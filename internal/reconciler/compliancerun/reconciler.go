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

// Reconciler reconciles compliance runs.
type Reconciler struct {
	Client     client.Client
	RESTConfig *rest.Config
	Config     configv1alpha1.ComplianceRunConfig
}

// Reconcile handles reconciliation requests for ComplianceRun resources.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues("name", req.Name)

	complianceRun := &v1alpha1.ComplianceRun{}

	if err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name}, complianceRun); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Object is gone, stop reconciling")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("error retrieving complianceRun: %w", err)
	}

	if len(complianceRun.Status.Phase) > 0 {
		log.Info("ComplianceRun already processed, stop reconciling", "phase", complianceRun.Status.Phase)
		return reconcile.Result{}, nil
	}

	// Update phase to Running
	patch := client.MergeFrom(complianceRun.DeepCopy())
	complianceRun.Status.Conditions = dikiv1alpha1helper.UpdateConditions(
		complianceRun.Status.Conditions,
		v1alpha1.ConditionTypeCompleted,
		v1alpha1.ConditionFalse,
		ConditionReasonRunning,
		"ComplianceRun is running",
		time.Now(),
	)
	complianceRun.Status.Phase = v1alpha1.ComplianceRunRunning
	if err := r.Client.Status().Patch(ctx, complianceRun, patch); err != nil {
		return reconcile.Result{}, r.handleFailedRun(ctx, complianceRun, log, err)
	}

	log.Info("Updated ComplianceRun phase to Running")

	// TODO(AleksandarSavchev): Create diki-runner job here.

	configMap, err := r.deployDikiConfigMap(ctx, complianceRun)
	if err != nil {
		return reconcile.Result{}, r.handleFailedRun(ctx, complianceRun, log, err)
	}

	log.Info(fmt.Sprintf("Created ConfigMap %s", client.ObjectKeyFromObject(configMap)))

	// Update phase to Completed
	patch = client.MergeFrom(complianceRun.DeepCopy())
	complianceRun.Status.Phase = v1alpha1.ComplianceRunCompleted
	complianceRun.Status.Conditions = dikiv1alpha1helper.UpdateConditions(
		complianceRun.Status.Conditions,
		v1alpha1.ConditionTypeCompleted,
		v1alpha1.ConditionTrue,
		ConditionReasonCompleted,
		"ComplianceRun has completed successfully",
		time.Now(),
	)
	if err := r.Client.Status().Patch(ctx, complianceRun, patch); err != nil {
		return reconcile.Result{}, r.handleFailedRun(ctx, complianceRun, log, err)
	}

	log.Info("Updated ComplianceRun phase to Completed")

	return ctrl.Result{}, nil
}
