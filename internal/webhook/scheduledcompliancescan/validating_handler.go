// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	scheduledcompliancescan "github.com/gardener/diki-operator/internal/reconciler/scheduledcompliancescan"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// ValidatingHandler is an admission webhook handler that validates ScheduledComplianceScan resources.
type ValidatingHandler struct {
	Decoder admission.Decoder
}

var _ admission.Handler = &ValidatingHandler{}

// Handle handles an admission request for a ScheduledComplianceScan resource and validates
// the embedded ComplianceScan spec template.
func (h *ValidatingHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	scheduledScan := &dikiv1alpha1.ScheduledComplianceScan{}
	if err := h.Decoder.DecodeRaw(req.Object, scheduledScan); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if scheduledScan.Spec.Schedule != "" {
		if _, err := scheduledcompliancescan.ParseCronScheduleWithPanicRecovery(scheduledScan.Spec.Schedule); err != nil {
			allErrs = append(allErrs, field.Invalid(specPath.Child("schedule"), scheduledScan.Spec.Schedule, fmt.Sprintf("invalid cron expression: %s", err.Error())))
		}
	}
	allErrs = append(allErrs, validateTimeZone(scheduledScan.Spec.TimeZone, specPath.Child("timeZone"))...)
	if scheduledScan.Spec.SuccessfulScansHistoryLimit != nil && *scheduledScan.Spec.SuccessfulScansHistoryLimit < 0 {
		allErrs = append(allErrs, field.Invalid(specPath.Child("successfulScansHistoryLimit"), *scheduledScan.Spec.SuccessfulScansHistoryLimit, "must not be negative"))
	}
	if scheduledScan.Spec.FailedScansHistoryLimit != nil && *scheduledScan.Spec.FailedScansHistoryLimit < 0 {
		allErrs = append(allErrs, field.Invalid(specPath.Child("failedScansHistoryLimit"), *scheduledScan.Spec.FailedScansHistoryLimit, "must not be negative"))
	}

	annotationsPath := field.NewPath("metadata").Child("annotations")
	if annotationValue, ok := scheduledScan.Annotations[scheduledcompliancescan.AnnotationScheduledComplianceScanOperation]; ok && annotationValue != scheduledcompliancescan.ScheduledComplianceScanOperationTrigger {
		allErrs = append(allErrs,
			field.NotSupported(annotationsPath.Key(scheduledcompliancescan.AnnotationScheduledComplianceScanOperation),
				annotationValue,
				[]string{scheduledcompliancescan.ScheduledComplianceScanOperationTrigger},
			),
		)
	}

	if len(allErrs) > 0 {
		return admission.Denied(allErrs.ToAggregate().Error())
	}

	return admission.Allowed("")
}

var validTimeZoneCharacters = regexp.MustCompile(`^[A-Za-z\.\-_0-9+]{1,14}$`)

// validateTimeZone was created by reusing validation from the kubernetes/kubernetes project
// https://github.com/kubernetes/kubernetes/blob/19c504053ed551f6e5ab5b5fd7c3345a6dcd8593/pkg/apis/batch/validation/validation.go#L966
func validateTimeZone(timeZone *string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if timeZone == nil {
		return allErrs
	}

	if len(*timeZone) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, timeZone, "timeZone must be nil or non-empty string"))
		return allErrs
	}

	for _, part := range strings.Split(*timeZone, "/") {
		if part == "." || part == ".." || strings.HasPrefix(part, "-") || !validTimeZoneCharacters.MatchString(part) {
			allErrs = append(allErrs, field.Invalid(fldPath, timeZone, fmt.Sprintf("unknown time zone %s", *timeZone)))
			return allErrs
		}
	}

	if strings.EqualFold(*timeZone, "Local") {
		allErrs = append(allErrs, field.Invalid(fldPath, timeZone, "timeZone must be an explicit time zone as defined in https://www.iana.org/time-zones"))
	}

	if _, err := time.LoadLocation(*timeZone); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, timeZone, err.Error()))
	}

	return allErrs
}
