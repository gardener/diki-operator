// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package compliancescan

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	compscanreconciler "github.com/gardener/diki-operator/internal/reconciler/compliancescan"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Handler is an admission webhook handler that restricts creation or updates to
// certain ComplianceScan resources.
type Handler struct {
	Client  client.Client
	Decoder admission.Decoder
}

// Handle handles an admission request for a ComplianceScan resource and restricts updates
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
		var (
			specFieldPath = field.NewPath("spec", "rulesets")
			allErrs       field.ErrorList
		)

		for rIdx, ruleset := range complianceScan.Spec.Rulesets {
			var (
				indexedRulesetConfigPath = specFieldPath.Index(rIdx).Child("options")
				rulesetOptionsPath       = indexedRulesetConfigPath.Child("ruleset")
				ruleOptionsPath          = indexedRulesetConfigPath.Child("rules")
				defaultRulesetOptionsKey = ruleset.ID
				defaultRuleOptionsKey    = fmt.Sprintf("%s%s", ruleset.ID, compscanreconciler.RuleOptionsSuffix)
			)

			if ruleset.Options == nil {
				continue
			}

			if ruleset.Options.Ruleset != nil && ruleset.Options.Ruleset.ConfigMapRef != nil {
				allErrs = append(allErrs, validateConfigMapReference(ctx, h.Client, ruleset.Options.Ruleset.ConfigMapRef, defaultRulesetOptionsKey, rulesetOptionsPath)...)
			}

			if ruleset.Options.Rules != nil && ruleset.Options.Rules.ConfigMapRef != nil {
				allErrs = append(allErrs, validateConfigMapReference(ctx, h.Client, ruleset.Options.Rules.ConfigMapRef, defaultRuleOptionsKey, ruleOptionsPath)...)
			}
		}

		if len(allErrs) > 0 {
			return admission.Denied(allErrs.ToAggregate().Error())
		}
		return admission.Allowed("")
	}

	return admission.Allowed("")
}

// TODO(georgibaltiev): Remove the defaultConfigMapKey once a mutating webhook for the compliance scan resource has been introduced.
func validateConfigMapReference(ctx context.Context, c client.Client, configMapRef *dikiv1alpha1.OptionsConfigMapRef, defaultConfigMapKey string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	optionsConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapRef.Name,
			Namespace: configMapRef.Namespace,
		},
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(optionsConfigMap), optionsConfigMap); err != nil {
		if apierrors.IsNotFound(err) {
			return append(allErrs, field.NotFound(fldPath, "the referenced configMap does not exist"))
		}
		return append(allErrs, field.InternalError(fldPath, err))
	}

	configMapKey := defaultConfigMapKey
	if configMapRef.Key != nil {
		configMapKey = *configMapRef.Key
	}

	if _, ok := optionsConfigMap.Data[configMapKey]; !ok {
		return append(allErrs, field.NotFound(fldPath.Child("key"), "the referenced key within the configMap does not exist"))
	}

	return allErrs
}
