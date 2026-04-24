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
	"k8s.io/utils/ptr"

	"github.com/gardener/diki-operator/imagevector"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// deployDikiRunJob creates a Kubernetes Job that runs the diki compliance scan.
func (r *Reconciler) deployDikiRunJob(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, dikiConfigMapName string) (*batchv1.Job, error) {
	dikiImage, err := imagevector.ImageVector().FindImage("diki")
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", JobNamePrefix, complianceScan.UID),
			Namespace: r.Config.DikiRunner.Namespace,
			Labels:    r.getLabels(complianceScan),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.To(int32(0)),
			Parallelism:  ptr.To(int32(0)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.getLabels(complianceScan),
				},
				Spec: corev1.PodSpec{
					ActiveDeadlineSeconds: ptr.To(int64(r.Config.DikiRunner.PodCompletionTimeout.Seconds())),
					Containers: []corev1.Container{
						{
							Name:  "diki-scan",
							Image: dikiImage.String(),
							Args: []string{
								"run",
								"--config=/config/config.yaml",
								"--all",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "diki-config",
									MountPath: "/config",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "diki-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: dikiConfigMapName,
									},
								},
							},
						},
					},
					ServiceAccountName: "diki-run",
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
				},
			},
		},
	}

	if err := r.Client.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create diki runner job: %w", err)
	}

	return job, nil
}
