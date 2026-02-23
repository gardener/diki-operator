// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1config "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
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

	log.Info(fmt.Sprintf("Reconciling ComplianceRun %s", req.Name))
	return ctrl.Result{}, nil
}
