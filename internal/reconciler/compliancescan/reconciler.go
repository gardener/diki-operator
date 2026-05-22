// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gardener/diki-operator/imagevector"
	configv1alpha1 "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	dikiv1alpha1helper "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1/helper"
)

// Reconciler reconciles compliance scans.
type Reconciler struct {
	Client     client.Client
	RESTConfig *rest.Config
	Config     configv1alpha1.ComplianceScanConfig
}

// Reconcile handles reconciliation requests for ComplianceScan resources.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	complianceScan := &v1alpha1.ComplianceScan{}

	if err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name}, complianceScan); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Object is gone, stop reconciling")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("error retrieving complianceScan: %w", err)
	}

	if complianceScan.Status.Phase == v1alpha1.ComplianceScanCompleted || complianceScan.Status.Phase == v1alpha1.ComplianceScanFailed {
		log.Info("ComplianceScan already processed, stop reconciling", "phase", complianceScan.Status.Phase)
		return reconcile.Result{}, nil
	}

	if complianceScan.Status.Phase == v1alpha1.ComplianceScanRunning {
		return r.checkJobStatus(ctx, complianceScan, log)
	}

	return r.deployResources(ctx, complianceScan, log)
}

func (r *Reconciler) deployResources(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, log logr.Logger) (ctrl.Result, error) {
	// Update phase to Running
	patch := client.MergeFrom(complianceScan.DeepCopy())
	complianceScan.Status.Conditions = dikiv1alpha1helper.UpdateConditions(
		complianceScan.Status.Conditions,
		v1alpha1.ConditionTypeCompleted,
		v1alpha1.ConditionFalse,
		ConditionReasonRunning,
		"ComplianceScan is running",
		time.Now(),
	)
	complianceScan.Status.Phase = v1alpha1.ComplianceScanRunning
	if err := r.Client.Status().Patch(ctx, complianceScan, patch); err != nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, err)
	}

	configMapName := fmt.Sprintf("%s%s", ConfigMapGenerateNamePrefix, complianceScan.UID)

	log.Info("Updated ComplianceScan phase to Running")

	job, err := r.deployDikiRunJob(ctx, complianceScan, configMapName)
	if err != nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, err)
	}

	log.Info(fmt.Sprintf("Created Job %s", client.ObjectKeyFromObject(job)))

	configMap, err := r.deployDikiConfigMap(ctx, configMapName, complianceScan, job)
	if err != nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, err)
	}

	log.Info(fmt.Sprintf("Created ConfigMap %s", client.ObjectKeyFromObject(configMap)))

	jobPatch := client.MergeFrom(job.DeepCopy())
	r.upscaleDikiRunJob(job)
	if err := r.Client.Patch(ctx, job, jobPatch); err != nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, fmt.Errorf("failed to upscale diki runner job: %w", err))
	}

	log.Info(fmt.Sprintf("Upscaled Job %s", client.ObjectKeyFromObject(job)))

	return reconcile.Result{RequeueAfter: ReconciliationRequeueInterval}, nil
}

// DeployDikiRunJob creates a Kubernetes Job that runs the diki compliance scan.
func (r *Reconciler) deployDikiRunJob(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, dikiConfigMapName string) (*batchv1.Job, error) {
	dikiImage, err := imagevector.ImageVector().FindImage("diki")
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "diki-run-",
			Namespace:    r.Config.DikiRunner.Namespace,
			Labels:       r.getLabels(complianceScan),
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

func (r *Reconciler) checkJobStatus(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, log logr.Logger) (ctrl.Result, error) {
	job, err := r.findDikiRunJob(ctx, complianceScan.UID)
	if err != nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, err)
	}

	if job == nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, fmt.Errorf("job not found for ComplianceScan %s", complianceScan.Name))
	}

	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			log.Info("Job completed successfully")
			if err := r.Client.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				return reconcile.Result{}, err
			}

			patch := client.MergeFrom(complianceScan.DeepCopy())
			complianceScan.Status.Phase = v1alpha1.ComplianceScanCompleted
			complianceScan.Status.Conditions = dikiv1alpha1helper.UpdateConditions(
				complianceScan.Status.Conditions,
				v1alpha1.ConditionTypeCompleted,
				v1alpha1.ConditionTrue,
				ConditionReasonCompleted,
				"ComplianceScan has completed successfully",
				time.Now(),
			)
			if err := r.Client.Status().Patch(ctx, complianceScan, patch); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}

		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			if err := r.Client.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, fmt.Errorf("job failed: %s", condition.Message))
		}
	}

	return reconcile.Result{RequeueAfter: ReconciliationRequeueInterval}, nil
}
