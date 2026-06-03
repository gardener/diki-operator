// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/version"
	"k8s.io/utils/ptr"

	"github.com/gardener/diki-operator/imagevector"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// deployDikiRunJob creates a Kubernetes Job that runs the diki compliance scan
// and exports the report to the configured outputs.
func (r *Reconciler) deployDikiRunJob(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, dikiConfigMapName string) (*batchv1.Job, error) {
	dikiImage, err := imagevector.ImageVector().FindImage("diki")
	if err != nil {
		return nil, err
	}

	reportExporterImage, err := imagevector.ImageVector().FindImage("report-exporter")
	if err != nil {
		return nil, err
	}
	reportExporterImage.WithOptionalTag(version.Get().GitVersion)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JobNamePrefix + string(complianceScan.UID),
			Namespace: r.Config.DikiRunner.Namespace,
			Labels:    r.getLabels(complianceScan),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.To(int32(0)),
			Suspend:      ptr.To(true),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.getLabels(complianceScan),
				},
				Spec: corev1.PodSpec{
					ActiveDeadlineSeconds: ptr.To(int64(r.Config.DikiRunner.PodCompletionTimeout.Seconds())),
					Containers: []corev1.Container{
						{
							Name:  DikiScanContainerName,
							Image: dikiImage.String(),
							Args: []string{
								"run",
								fmt.Sprintf("--config=%s/%s", DikiConfigMountPath, DikiConfigKey),
								"--all",
								fmt.Sprintf("--output=%s/%s", ReportMountPath, ReportFileName),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      DikiConfigVolumeName,
									MountPath: DikiConfigMountPath,
									ReadOnly:  true,
								},
								{
									Name:      ReportVolumeName,
									MountPath: ReportMountPath,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptr.To(false),
								ReadOnlyRootFilesystem:   ptr.To(true),
								Privileged:               ptr.To(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
						{
							Name:  ReportExporterContainerName,
							Image: reportExporterImage.String(),
							Args: []string{
								fmt.Sprintf("--config=%s/%s", DikiConfigMountPath, ExporterConfigKey),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      ReportVolumeName,
									MountPath: ReportMountPath,
									ReadOnly:  true,
								},
								{
									Name:      DikiConfigVolumeName,
									MountPath: DikiConfigMountPath,
									ReadOnly:  true,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptr.To(false),
								ReadOnlyRootFilesystem:   ptr.To(true),
								Privileged:               ptr.To(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: DikiConfigVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: dikiConfigMapName,
									},
									DefaultMode: ptr.To(int32(0440)),
								},
							},
						},
						{
							Name: ReportVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					ServiceAccountName: ServiceAccountNameDikiRun,
					RestartPolicy:      corev1.RestartPolicyNever,
					Tolerations: []corev1.Toleration{
						{
							Effect:   corev1.TaintEffectNoSchedule,
							Operator: corev1.TolerationOpExists,
						},
						{
							Effect:   corev1.TaintEffectNoExecute,
							Operator: corev1.TolerationOpExists,
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						FSGroup:      ptr.To(int64(65532)),
						RunAsUser:    ptr.To(int64(65532)),
						RunAsGroup:   ptr.To(int64(65532)),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
				},
			},
		},
	}

	if err := r.Client.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create diki runner job: %w", err)
	}

	return job, nil
}
