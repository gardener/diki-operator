// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha1config "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Reconciler reconciles compliance runs.
type Reconciler struct {
	Client     client.Client
	RESTConfig *rest.Config
	Config     v1alpha1config.ComplianceRunConfig
}

// Reconcile handles reconciliation requests for ComplianceRun resources.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	complianceRun := &v1alpha1.ComplianceRun{}

	if err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name}, complianceRun); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Object is gone, stop reconciling")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("error retrieving complianceRun: %w", err)
	}

	if len(complianceRun.Status.Phase) > 0 {
		log.Info("ComplianceRun already processed, stop reconciling", "name", complianceRun.Name, "phase", complianceRun.Status.Phase)
		return reconcile.Result{}, nil
	}

	// Update phase to Running
	patch := client.MergeFrom(complianceRun.DeepCopy())
	// TODO(AleksandarSavchev): Update conditions here.
	complianceRun.Status.Phase = v1alpha1.ComplianceRunRunning
	if err := r.Client.Status().Patch(ctx, complianceRun, patch); err != nil {
		return reconcile.Result{}, r.handleFailedRun(ctx, complianceRun, err)
	}

	log.Info("Updated ComplianceRun phase to Running", "name", complianceRun.Name)

	dikiConfig, err := r.deployDikiConfigMap(ctx, complianceRun)
	if err != nil {
		return reconcile.Result{}, r.handleFailedRun(ctx, complianceRun, err)
	}

	log.Info(fmt.Sprintf("Created ConfigMap %s/%s", dikiConfig.Namespace, dikiConfig.Name))

	return ctrl.Result{}, nil
}
