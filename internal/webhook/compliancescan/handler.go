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
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Handler is an admission webhook handler that restricts creation or updates to
// certain ComplianceScan resources.
type Handler struct {
	Client  client.Client
	Logger  logr.Logger
	Decoder admission.Decoder
}

// Handle handles an admission request for an ComplianceScan resource and restricts updates
// and creations if it contains references to invalid ConfigMaps.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	complianceScan := &dikiv1alpha1.ComplianceScan{}
	if err := h.Decoder.DecodeRaw(req.Object, complianceScan); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == admissionv1.Update {
		oldComplianceScan := &dikiv1alpha1.ComplianceScan{}
		if err := h.Decoder.DecodeRaw(req.OldObject, oldComplianceScan); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if !apiequality.Semantic.DeepEqual(oldComplianceScan.Spec, complianceScan.Spec) {
			return admission.Denied("updating the ComplianceScan spec is not permitted")
		}
		return admission.Allowed("")
	}

	if req.Operation == admissionv1.Create {
		var specFieldPath = field.NewPath("spec", "rulesets")

		for rIdx, ruleset := range complianceScan.Spec.Rulesets {
			var (
				indexedRulesetConfigPath = specFieldPath.Index(rIdx).Child("options")
				rulesetOptionsPath       = indexedRulesetConfigPath.Child("ruleset")
				ruleOptionsPath          = indexedRulesetConfigPath.Child("rules")
			)

			if ruleset.Options == nil {
				continue
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
						return admission.Denied(fmt.Sprintf("%s: the referenced configMap does not exist", rulesetOptionsPath.String()))
					}
					return admission.Denied(fmt.Sprintf("failed to retrieve the referenced configMap %s: %s", rulesetOptionsPath.String(), err.Error()))
				}

				if ruleset.Options.Ruleset.ConfigMapRef.Key != nil {
					if _, ok := rulesetOptionsConfigMap.Data[*ruleset.Options.Ruleset.ConfigMapRef.Key]; !ok {
						return admission.Denied(fmt.Sprintf("%s: the referenced key within the configMap does not exist", rulesetOptionsPath.Child("key").String()))
					}
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
						return admission.Denied(fmt.Sprintf("%s: the referenced configMap does not exist", ruleOptionsPath.String()))
					}
					return admission.Denied(fmt.Sprintf("failed to retrieve the referenced configMap %s: %s", ruleOptionsPath.String(), err.Error()))
				}

				if ruleset.Options.Rules.ConfigMapRef.Key != nil {
					if _, ok := ruleOptionsConfigMap.Data[*ruleset.Options.Rules.ConfigMapRef.Key]; !ok {
						return admission.Denied(fmt.Sprintf("%s: the referenced key within the configMap does not exist", ruleOptionsPath.Child("key").String()))
					}
				}
			}
		}
		return admission.Allowed("")
	}

	return admission.Allowed("")
}
