// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"

	"github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	. "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1/validation"
)

var _ = Describe("#ValidateDikiOperatorConfiguration", func() {
	var conf *v1alpha1.DikiOperatorConfiguration

	BeforeEach(func() {
		conf = &v1alpha1.DikiOperatorConfiguration{
			LogLevel:  "info",
			LogFormat: "json",
			Controllers: v1alpha1.ControllerConfiguration{
				ComplianceRun: v1alpha1.ComplianceRunConfig{
					DikiRunner: v1alpha1.DikiRunnerConfig{
						Namespace: "diki-runner",
						Labels: map[string]string{
							"app": "diki-runner",
						},
						WaitInterval:         &metav1.Duration{Duration: 30 * time.Second},
						ExecTimeout:          &metav1.Duration{Duration: 5 * time.Minute},
						PodCompletionTimeout: &metav1.Duration{Duration: 10 * time.Minute},
					},
				},
			},
			Server: v1alpha1.ServerConfiguration{
				HealthProbes: &v1alpha1.Server{
					Port: 8081,
				},
				Metrics: &v1alpha1.Server{
					Port: 8080,
				},
			},
			LeaderElection: &componentbaseconfigv1alpha1.LeaderElectionConfiguration{
				LeaderElect:       ptr.To(true),
				LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
				RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
				RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
				ResourceLock:      "leases",
				ResourceName:      "diki-operator-leader-election",
				ResourceNamespace: "kube-system",
			},
		}
	})

	It("should pass validation with valid configuration", func() {
		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(BeEmpty())
	})

	It("should pass validation when LeaderElectionConfiguration is nil", func() {
		conf.LeaderElection = nil

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(BeEmpty())
	})

	It("should fail validation when LogLevel is invalid", func() {
		conf.LogLevel = "invalid"

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeNotSupported),
			"Field":    Equal("logLevel"),
			"BadValue": Equal("invalid"),
		}))))
	})

	It("should fail validation when LogFormat is invalid", func() {
		conf.LogFormat = "invalid"

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeNotSupported),
			"Field":    Equal("logFormat"),
			"BadValue": Equal("invalid"),
		}))))
	})

	It("should fail validation when labels contain invalid characters", func() {
		conf.Controllers.ComplianceRun.DikiRunner.Labels = map[string]string{
			"!invalid": "value",
		}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceRun.dikiRunner.labels"),
			"BadValue": Equal("!invalid"),
		}))))
	})

	It("should faile validation when WaitInterval is less than or equal to 0", func() {
		conf.Controllers.ComplianceRun.DikiRunner.WaitInterval = &metav1.Duration{Duration: -1 * time.Second}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceRun.dikiRunner.waitInterval"),
			"BadValue": Equal(&metav1.Duration{Duration: -1 * time.Second}),
		}))))
	})

	It("should fail validation when ExecTimeout is less than or equal to 0", func() {
		conf.Controllers.ComplianceRun.DikiRunner.ExecTimeout = &metav1.Duration{Duration: 0}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceRun.dikiRunner.execTimeout"),
			"BadValue": Equal(&metav1.Duration{Duration: 0}),
		}))))
	})

	It("should fail validation when PodCompletionTimeout is less than or equal to 0", func() {
		conf.Controllers.ComplianceRun.DikiRunner.PodCompletionTimeout = &metav1.Duration{Duration: -5 * time.Minute}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceRun.dikiRunner.podCompletionTimeout"),
			"BadValue": Equal(&metav1.Duration{Duration: -5 * time.Minute}),
		}))))
	})
})
