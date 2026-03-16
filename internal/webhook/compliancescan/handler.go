// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package compliancescan

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Handler is an admission webhook handler that restricts updates to certain fields
// of managed OpenIDConnect resources.
type Handler struct {
	Client  client.Client
	Logger  logr.Logger
	Decoder admission.Decoder
}

// Handle handles an admission request for an ComplianceScan resource and restricts updates
// and creations if it contains references to invalid ConfigMaps.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.Logger.Info("ComplianceScan input validation invoked",
		"operation", req.Operation,
		"resource", req.Resource.Resource,
		"name", req.Name,
		"username", req.UserInfo.Username,
	)

	if req.Operation == admissionv1.Update {
		return admission.Denied("updating ComplianceScan resources is not permitted")
	}

	if req.Operation != admissionv1.Create {
		return admission.Allowed("")
	}

	complianceScan := &dikiv1alpha1.ComplianceScan{}
	if err := h.Decoder.DecodeRaw(req.Object, complianceScan); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var specFieldPath = field.NewPath("spec", "rulesets")

	for rIdx, ruleset := range complianceScan.Spec.Rulesets {
		var rulesetFieldPath = specFieldPath.Index(rIdx).Child("options")

		if ruleset.Options == nil {
			return admission.Allowed("")
		}

		if ruleset.Options.Ruleset != nil && ruleset.Options.Ruleset.ConfigMapRef != nil {
			rulesetOptionsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ruleset.Options.Ruleset.ConfigMapRef.Name,
					Namespace: ruleset.Options.Ruleset.ConfigMapRef.Namespace,
				},
			}

			if err := h.Client.Get(ctx, client.ObjectKeyFromObject(rulesetOptionsConfigMap), rulesetOptionsConfigMap); err != nil {
				if apierrors.IsNotFound(err) {
					return admission.Denied(fmt.Sprintf("%s: the referenced configMap does not exist", rulesetFieldPath.Child("ruleset").String()))
				}
				return admission.Errored(http.StatusBadRequest, err)
			}
		}

		if ruleset.Options.Rules != nil && ruleset.Options.Rules.ConfigMapRef != nil {
			ruleOptionsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ruleset.Options.Rules.ConfigMapRef.Name,
					Namespace: ruleset.Options.Rules.ConfigMapRef.Namespace,
				},
			}

			if err := h.Client.Get(ctx, client.ObjectKeyFromObject(ruleOptionsConfigMap), ruleOptionsConfigMap); err != nil {
				if apierrors.IsNotFound(err) {
					return admission.Denied(fmt.Sprintf("%s: the referenced configMap does not exist", rulesetFieldPath.Child("rules").String()))
				}
				return admission.Errored(http.StatusBadRequest, err)
			}
		}
	}

	return admission.Allowed("")
}
