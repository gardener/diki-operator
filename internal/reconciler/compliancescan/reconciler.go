// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

	if len(complianceScan.Status.Phase) > 0 {
		log.Info("ComplianceScan already processed, stop reconciling", "phase", complianceScan.Status.Phase)
		return reconcile.Result{}, nil
	}

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

	log.Info("Updated ComplianceScan phase to Running")

	// TODO(AleksandarSavchev): Create diki-runner job here.

	configMap, err := r.deployDikiConfigMap(ctx, complianceScan)
	if err != nil {
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, err)
	}

	log.Info(fmt.Sprintf("Created ConfigMap %s", client.ObjectKeyFromObject(configMap)))

	// Update phase to Completed
	patch = client.MergeFrom(complianceScan.DeepCopy())
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
		return reconcile.Result{}, r.handleFailedScan(ctx, complianceScan, log, err)
	}

	log.Info("Updated ComplianceScan phase to Completed")

	return ctrl.Result{}, nil
}
