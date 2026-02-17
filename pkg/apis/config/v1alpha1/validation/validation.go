// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
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

	if conf.LogLevel != "" {
		if !sets.New(logger.AllLogLevels...).Has(conf.LogLevel) {
			allErrs = append(allErrs, field.NotSupported(field.NewPath("logLevel"), conf.LogLevel, logger.AllLogLevels))
		}
	}

	if conf.LogFormat != "" {
		if !sets.New(logger.AllLogFormats...).Has(conf.LogFormat) {
			allErrs = append(allErrs, field.NotSupported(field.NewPath("logFormat"), conf.LogFormat, logger.AllLogFormats))
		}
	}

	allErrs = append(allErrs, validateControllers(&conf.Controllers, field.NewPath("controllers"))...)
	allErrs = append(allErrs, validationutils.ValidateLeaderElectionConfiguration(conf.LeaderElection, field.NewPath("leaderElection"))...)

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

	if dikiRunner.WaitInterval != nil && dikiRunner.WaitInterval.Duration <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("waitInterval"), dikiRunner.WaitInterval, "waitInterval must be greater than 0"))
	}
	if dikiRunner.ExecTimeout != nil && dikiRunner.ExecTimeout.Duration <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("execTimeout"), dikiRunner.ExecTimeout, "execTimeout must be greater than 0"))
	}
	if dikiRunner.PodCompletionTimeout != nil && dikiRunner.PodCompletionTimeout.Duration <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("podCompletionTimeout"), dikiRunner.PodCompletionTimeout, "podCompletionTimeout must be greater than 0"))
	}

	return allErrs
}
