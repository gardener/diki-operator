// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"time"

	"github.com/gardener/gardener/pkg/controllerutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	// ControllerName is the name of the garbagecollector controller.
	ControllerName = "garbagecollector"
	// ReconciliationTimeout is the timeout passed to the context of the Reconcile call.
	ReconciliationTimeout = 1 * time.Minute
	// MaxConcurrentReconciles is the maximum number of concurrent Reconcile calls.
	MaxConcurrentReconciles = 1
)

// SetupWithManager specifies how the controller is built to periodically clean up Jobs.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	return builder.ControllerManagedBy(mgr).
		Named(ControllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: MaxConcurrentReconciles,
			ReconciliationTimeout:   ReconciliationTimeout,
		}).
		WatchesRawSource(controllerutils.EnqueueOnce).
		Complete(r)
}
