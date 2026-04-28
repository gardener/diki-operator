// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configv1alpha1 "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
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
		job, err := r.findDikiRunJob(ctx, complianceScan.UID, r.Config.DikiRunner.Namespace)
		if err != nil {
			return reconcile.Result{}, r.patchFailed(ctx, complianceScan, log, err)
		}

		for _, condition := range job.Status.Conditions {
			if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
				log.Info("Job completed successfully")
				return reconcile.Result{}, r.patchCompleted(ctx, complianceScan)
			}

			if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
				return reconcile.Result{}, r.patchFailed(ctx, complianceScan, log, fmt.Errorf("job failed: %s", condition.Message))
			}
		}

		return reconcile.Result{RequeueAfter: ReconciliationRequeueInterval}, nil
	}

	if err := r.patchRunning(ctx, complianceScan); err != nil {
		return reconcile.Result{}, r.patchFailed(ctx, complianceScan, log, err)
	}

	log.Info("Updated ComplianceScan phase to Running")

	if err := r.deployResources(ctx, complianceScan, log); err != nil {
		return reconcile.Result{}, r.patchFailed(ctx, complianceScan, log, err)
	}

	return reconcile.Result{RequeueAfter: ReconciliationRequeueInterval}, nil
}

func (r *Reconciler) deployResources(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, log logr.Logger) error {
	configMapName := ConfigMapNamePrefix + string(complianceScan.UID)

	job, err := r.deployDikiRunJob(ctx, complianceScan, configMapName)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Created Job %s", client.ObjectKeyFromObject(job)))

	configMap, err := r.deployDikiConfigMap(ctx, configMapName, complianceScan, job)
	if err != nil {
		if deleteErr := r.Client.Delete(ctx, job); deleteErr != nil {
			log.Error(deleteErr, "Failed to delete orphaned Job during cleanup", "job", client.ObjectKeyFromObject(job))
		}
		return err
	}
	log.Info(fmt.Sprintf("Created ConfigMap %s", client.ObjectKeyFromObject(configMap)))

	if err := r.startDikiRunJob(ctx, job); err != nil {
		if deleteErr := r.Client.Delete(ctx, job); deleteErr != nil {
			log.Error(deleteErr, "Failed to delete orphaned Job during cleanup", "job", client.ObjectKeyFromObject(job))
		}
		return fmt.Errorf("failed to upscale diki runner job: %w", err)
	}
	log.Info(fmt.Sprintf("Upscaled Job %s", client.ObjectKeyFromObject(job)))

	return nil
}
