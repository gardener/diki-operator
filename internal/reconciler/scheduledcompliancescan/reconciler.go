// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Reconciler reconciles scheduled compliance scans.
type Reconciler struct {
	Client client.Client
	Clock  clock.Clock
}

// Reconcile handles reconciliation requests for ScheduledComplianceScan resources.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	scheduledScan := &v1alpha1.ScheduledComplianceScan{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name}, scheduledScan); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Object is gone, stop reconciling")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("error retrieving ScheduledComplianceScan: %w", err)
	}

	childScans := &v1alpha1.ComplianceScanList{}
	if err := r.Client.List(ctx, childScans, client.MatchingLabels{
		LabelScheduledComplianceScanName: scheduledScan.Name,
		LabelScheduledComplianceScanUID:  string(scheduledScan.UID),
	}); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing child ComplianceScans: %w", err)
	}

	// Categorize child scans.
	var activeScan *v1alpha1.ComplianceScan
	var successfulScans, failedScans []v1alpha1.ComplianceScan
	for i := range childScans.Items {
		switch childScans.Items[i].Status.Phase {
		case v1alpha1.ComplianceScanCompleted:
			successfulScans = append(successfulScans, childScans.Items[i])
		case v1alpha1.ComplianceScanFailed:
			failedScans = append(failedScans, childScans.Items[i])
		default:
			activeScan = &childScans.Items[i]
		}
	}

	now := r.Clock.Now()

	// Detect active scan completion: if status references an active scan but it is now finished.
	if scheduledScan.Status.Active != nil && activeScan == nil {
		patch := client.MergeFrom(scheduledScan.DeepCopy())
		scheduledScan.Status.Active = nil
		scheduledScan.Status.LastCompletionTime = &metav1.Time{Time: now}
		if err := r.Client.Status().Patch(ctx, scheduledScan, patch); err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to update ScheduledComplianceScan status: %w", err)
		}
		log.Info("Active ComplianceScan finished, cleared active reference")
	}

	// If there is an active scan but Status.Active is not set, fix it up.
	// This handles the case where a ComplianceScan was created but the
	// subsequent status patch failed.
	if activeScan != nil && scheduledScan.Status.Active == nil {
		if err := r.setActiveScan(ctx, scheduledScan, activeScan, activeScan.CreationTimestamp.Time); err != nil {
			return reconcile.Result{}, err
		}
		log.Info("Recovered active reference for orphaned ComplianceScan", "childName", activeScan.Name)
		return reconcile.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	if activeScan != nil {
		return reconcile.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	expr, err := parseCronScheduleWithPanicRecovery(scheduledScan.Spec.Schedule)
	if err != nil {
		log.Error(err, "Invalid cron expression", "schedule", scheduledScan.Spec.Schedule)
		return reconcile.Result{}, nil
	}

	shouldCreate := false
	if scheduledScan.Status.LastScheduleTime == nil {
		shouldCreate = true
	} else {
		nextSchedule := expr.Next(scheduledScan.Status.LastScheduleTime.Time)
		if !now.Before(nextSchedule) {
			shouldCreate = true
		}
	}

	if shouldCreate {
		childScan, err := r.deployComplianceScan(ctx, scheduledScan, now)
		if err != nil {
			log.Error(err, "Failed to create ComplianceScan")
			return reconcile.Result{}, err
		}
		log.Info("Created ComplianceScan", "childName", childScan.Name)

		if err := r.setActiveScan(ctx, scheduledScan, childScan, now); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Clean up old scans per their respective history limits.
	r.cleanupOldScans(ctx, log, successfulScans, int(ptr.Deref(scheduledScan.Spec.SuccessfulScansHistoryLimit, 0)))
	r.cleanupOldScans(ctx, log, failedScans, int(ptr.Deref(scheduledScan.Spec.FailedScansHistoryLimit, 0)))

	// Calculate requeue time for the next scheduled run.
	var referenceTime time.Time
	if scheduledScan.Status.LastScheduleTime != nil {
		referenceTime = scheduledScan.Status.LastScheduleTime.Time
	} else {
		referenceTime = now
	}
	nextRun := expr.Next(referenceTime)
	requeueAfter := max(nextRun.Sub(now), 0)

	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}
