// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package compliancescan

import (
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// HandlerName is the name of this admission webhook handler.
	HandlerName = "compliancescan"
	// WebhookPath is the HTTP handler path for this admission webhook handler.
	WebhookPath = "/webhooks/compliancescan"
)

// AddToManager adds Handler to the given manager.
func AddToManager(mgr manager.Manager) error {
	webhook := &admission.Webhook{
		Handler: &Handler{
			Client:  mgr.GetClient(),
			Decoder: admission.NewDecoder(mgr.GetScheme()),
		},
		RecoverPanic: ptr.To(true),
	}

	mgr.GetWebhookServer().Register(WebhookPath, webhook)
	return nil
}
