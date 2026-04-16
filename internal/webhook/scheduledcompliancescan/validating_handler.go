// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/cronexpr"
	admissionv1 "k8s.io/api/admission/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// Handler is an admission webhook handler that validates ScheduledComplianceScan resources.
type Handler struct {
	Client  client.Client
	Decoder admission.Decoder
}

// Handle handles an admission request for a ScheduledComplianceScan resource and validates
// the embedded ComplianceScan spec template.
func (h *Handler) Handle(_ context.Context, req admission.Request) admission.Response {
	scheduledScan := &dikiv1alpha1.ScheduledComplianceScan{}
	if err := h.Decoder.DecodeRaw(req.Object, scheduledScan); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if scheduledScan.Spec.Schedule != "" {
		if _, err := cronexpr.Parse(scheduledScan.Spec.Schedule); err != nil {
			allErrs = append(allErrs, field.Invalid(specPath.Child("schedule"), scheduledScan.Spec.Schedule, fmt.Sprintf("invalid cron expression: %s", err.Error())))
		}
	}
	if scheduledScan.Spec.SuccessfulScansHistoryLimit != nil && *scheduledScan.Spec.SuccessfulScansHistoryLimit < 0 {
		allErrs = append(allErrs, field.Invalid(specPath.Child("successfulScansHistoryLimit"), *scheduledScan.Spec.SuccessfulScansHistoryLimit, "must not be negative"))
	}
	if scheduledScan.Spec.FailedScansHistoryLimit != nil && *scheduledScan.Spec.FailedScansHistoryLimit < 0 {
		allErrs = append(allErrs, field.Invalid(specPath.Child("failedScansHistoryLimit"), *scheduledScan.Spec.FailedScansHistoryLimit, "must not be negative"))
	}

	if req.Operation == admissionv1.Update {
		oldScheduledScan := &dikiv1alpha1.ScheduledComplianceScan{}
		if err := h.Decoder.DecodeRaw(req.OldObject, oldScheduledScan); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if !apiequality.Semantic.DeepEqual(oldScheduledScan.Spec.ScanTemplate, scheduledScan.Spec.ScanTemplate) {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("scanTemplate"), "updating the ScheduledComplianceScan scanTemplate is not permitted"))
		}
	}

	if len(allErrs) > 0 {
		return admission.Denied(allErrs.ToAggregate().Error())
	}

	return admission.Allowed("")
}
