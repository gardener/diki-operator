// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package compliancescan

import (
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

const (
	// HandlerName is the name of this admission webhook handler.
	HandlerName = "compliancescan"
	// ValidatingWebhookPath is the HTTP handler path for the validating admission webhook.
	ValidatingWebhookPath = "/webhooks/compliancescan/validating"
)

// AddToManager adds the validating webhook handler to the given manager.
func AddToManager(mgr manager.Manager, defaultOutputs []reportexporterv1alpha1.Output) error {
	webhook := &admission.Webhook{
		Handler: &ValidatingHandler{
			Client:         mgr.GetClient(),
			Decoder:        admission.NewDecoder(mgr.GetScheme()),
			DefaultOutputs: defaultOutputs,
		},
		RecoverPanic: ptr.To(true),
	}

	mgr.GetWebhookServer().Register(ValidatingWebhookPath, webhook)
	return nil
}
