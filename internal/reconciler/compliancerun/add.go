// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"time"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

const (
	// ControllerName is the name of the compliancerun controller.
	ControllerName = "compliancerun"
	// ReconciliationTimeout is the timeout passed to the context of the Reconcile call.
	ReconciliationTimeout = 10 * time.Minute
	// MaxConcurrentReconciles is the maximum number of concurrent Reconcile calls.
	MaxConcurrentReconciles = 50
)

// SetupWithManager specifies how the controller is built to watch ComplianceRun resources.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	if r.RESTConfig == nil {
		r.RESTConfig = mgr.GetConfig()
	}

	return builder.ControllerManagedBy(mgr).
		Named(ControllerName).
		For(&dikiv1alpha1.ComplianceRun{}, builder.WithPredicates(r.Predicate())).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: MaxConcurrentReconciles,
			RateLimiter: workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](5*time.Second, 2*time.Minute),
				&workqueue.TypedBucketRateLimiter[reconcile.Request]{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
			),
			ReconciliationTimeout: ReconciliationTimeout,
		}).
		Complete(r)
}

// Predicate returns a predicate to filter ComplianceRun events.
func (r *Reconciler) Predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(_ event.CreateEvent) bool { return true },
		UpdateFunc:  func(_ event.UpdateEvent) bool { return false },
		DeleteFunc:  func(_ event.DeleteEvent) bool { return false },
		GenericFunc: func(_ event.GenericEvent) bool { return false },
	}
}
