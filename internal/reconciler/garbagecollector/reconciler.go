// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gardener/diki-operator/internal/constants"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Config holds configuration for the garbagecollector controller.
type Config struct {
	Namespace       string
	RequeueInterval time.Duration
}

// Reconciler periodically cleans up diki-run Jobs that are no longer needed.
type Reconciler struct {
	Client client.Client
	Config Config
}

// Reconcile lists all diki-run Jobs and deletes those linked to a ComplianceScan
// that no longer exists or is in a terminal state (Completed/Failed).
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	complianceScanList := &v1alpha1.ComplianceScanList{}
	if err := r.Client.List(ctx, complianceScanList); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to list ComplianceScans: %w", err)
	}

	scanPhases := make(map[string]v1alpha1.ComplianceScanPhase, len(complianceScanList.Items))
	for _, complianceScan := range complianceScanList.Items {
		scanPhases[string(complianceScan.UID)] = complianceScan.Status.Phase
	}

	jobList := &batchv1.JobList{}
	if err := r.Client.List(ctx, jobList,
		client.InNamespace(r.Config.Namespace),
		client.HasLabels{constants.LabelComplianceScanUID},
	); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to list diki-run Jobs: %w", err)
	}

	var errs []error
	for i := range jobList.Items {
		job := &jobList.Items[i]
		complianceScanUID := job.Labels[constants.LabelComplianceScanUID]

		if !shouldDeleteJob(scanPhases, complianceScanUID) {
			continue
		}

		log.Info("Deleting Job", "job", client.ObjectKeyFromObject(job), "complianceScanUID", complianceScanUID)
		if err := r.Client.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("failed to delete Job %s: %w", client.ObjectKeyFromObject(job), err))
		}
	}

	return reconcile.Result{RequeueAfter: r.Config.RequeueInterval}, errors.Join(errs...)
}

func shouldDeleteJob(scanPhases map[string]v1alpha1.ComplianceScanPhase, complianceScanUID string) bool {
	phase, exists := scanPhases[complianceScanUID]
	return !exists || phase == v1alpha1.ComplianceScanCompleted || phase == v1alpha1.ComplianceScanFailed
}
