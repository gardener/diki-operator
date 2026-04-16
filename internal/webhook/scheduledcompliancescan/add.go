// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan

import (
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// HandlerName is the name of this admission webhook handler.
	HandlerName = "scheduledcompliancescan"
	// ValidatingWebhookPath is the HTTP handler path for the validating admission webhook.
	ValidatingWebhookPath = "/webhooks/scheduledcompliancescan/validate"
	// MutatingWebhookPath is the HTTP handler path for the mutating admission webhook.
	MutatingWebhookPath = "/webhooks/scheduledcompliancescan/mutate"
)

// AddToManager registers the validating and mutating webhook handlers with the given manager.
func AddToManager(mgr manager.Manager) error {
	decoder := admission.NewDecoder(mgr.GetScheme())

	mgr.GetWebhookServer().Register(ValidatingWebhookPath, &admission.Webhook{
		Handler: &Handler{
			Client:  mgr.GetClient(),
			Decoder: decoder,
		},
		RecoverPanic: ptr.To(true),
	})

	mgr.GetWebhookServer().Register(MutatingWebhookPath, &admission.Webhook{
		Handler: &MutatingHandler{
			Decoder: decoder,
		},
		RecoverPanic: ptr.To(true),
	})

	return nil
}
