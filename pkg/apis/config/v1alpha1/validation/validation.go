// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"time"

	"github.com/gardener/gardener/pkg/logger"
	validationutils "github.com/gardener/gardener/pkg/utils/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
)

// ValidateDikiOperatorConfiguration validates the given `DikiOperatorConfiguration`.
func ValidateDikiOperatorConfiguration(conf *v1alpha1.DikiOperatorConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateLog(&conf.Log, field.NewPath("log"))...)
	allErrs = append(allErrs, validateControllers(&conf.Controllers, field.NewPath("controllers"))...)
	allErrs = append(allErrs, validationutils.ValidateLeaderElectionConfiguration(conf.LeaderElection, field.NewPath("leaderElection"))...)

	return allErrs
}

// validateLog validates the log configuration.
func validateLog(log *v1alpha1.Log, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if log.Level != "" {
		if !sets.New(logger.AllLogLevels...).Has(log.Level) {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("level"), log.Level, logger.AllLogLevels))
		}
	}

	if log.Format != "" {
		if !sets.New(logger.AllLogFormats...).Has(log.Format) {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("format"), log.Format, logger.AllLogFormats))
		}
	}

	return allErrs
}

// validateControllers validates the controllers configuration.
func validateControllers(controllers *v1alpha1.ControllerConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateDikiRunner(controllers.ComplianceRun.DikiRunner, fldPath.Child("complianceRun", "dikiRunner"))...)

	return allErrs
}

// validateDikiRunner validates the DikiRunner configuration.
func validateDikiRunner(dikiRunner v1alpha1.DikiRunnerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, metav1validation.ValidateLabels(dikiRunner.Labels, fldPath.Child("labels"))...)

	if dikiRunner.PodCompletionTimeout != nil && (dikiRunner.PodCompletionTimeout.Duration <= 0 || dikiRunner.PodCompletionTimeout.Duration > 1*time.Hour) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("podCompletionTimeout"), dikiRunner.PodCompletionTimeout, "podCompletionTimeout must be greater than 0 and less than or equal to 1 hour"))
	}

	return allErrs
}
