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
			Log: v1alpha1.Log{
				Level:  "info",
				Format: "json",
			},
			Controllers: v1alpha1.ControllerConfiguration{
				ComplianceScan: v1alpha1.ComplianceScanConfig{
					DikiRunner: v1alpha1.DikiRunnerConfig{
						Namespace: "diki-runner",
						Labels: map[string]string{
							"app": "diki-runner",
						},
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
				Webhooks: v1alpha1.HTTPSServer{
					Server: v1alpha1.Server{
						Port: 10443,
					},
					TLS: v1alpha1.TLS{
						ServerCertDir: "/etc/diki-operator/webhooks/tls",
					},
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
		conf.Log.Level = "invalid"

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeNotSupported),
			"Field":    Equal("log.level"),
			"BadValue": Equal("invalid"),
		}))))
	})

	It("should fail validation when LogFormat is invalid", func() {
		conf.Log.Format = "invalid"

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeNotSupported),
			"Field":    Equal("log.format"),
			"BadValue": Equal("invalid"),
		}))))
	})

	It("should fail validation when labels contain invalid characters", func() {
		conf.Controllers.ComplianceScan.DikiRunner.Labels = map[string]string{
			"!invalid": "value",
		}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceScan.dikiRunner.labels"),
			"BadValue": Equal("!invalid"),
		}))))
	})

	It("should fail validation when PodCompletionTimeout is less than or equal to 0", func() {
		conf.Controllers.ComplianceScan.DikiRunner.PodCompletionTimeout = &metav1.Duration{Duration: -5 * time.Minute}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceScan.dikiRunner.podCompletionTimeout"),
			"BadValue": Equal(&metav1.Duration{Duration: -5 * time.Minute}),
		}))))
	})

	It("should fail validation when PodCompletionTimeout is greater than 1 hour", func() {
		conf.Controllers.ComplianceScan.DikiRunner.PodCompletionTimeout = &metav1.Duration{Duration: 2 * time.Hour}

		errorList := ValidateDikiOperatorConfiguration(conf)
		Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal("controllers.complianceScan.dikiRunner.podCompletionTimeout"),
			"BadValue": Equal(&metav1.Duration{Duration: 2 * time.Hour}),
		}))))
	})

	Describe("ServerConfiguration", func() {
		It("should forbid negative HealthProbes port", func() {
			conf.Server.HealthProbes.Port = -1

			errs := ValidateDikiOperatorConfiguration(conf)
			Expect(errs).To(ConsistOf(PointTo(
				MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("server.healthProbes.port"),
				}),
			)))
		})

		It("should forbid negative Metrics port", func() {
			conf.Server.Metrics.Port = -1

			errs := ValidateDikiOperatorConfiguration(conf)
			Expect(errs).To(ConsistOf(PointTo(
				MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("server.metrics.port"),
				}),
			)))
		})

		It("should forbid negative Webhooks port", func() {
			conf.Server.Webhooks.Port = -1

			errs := ValidateDikiOperatorConfiguration(conf)
			Expect(errs).To(ConsistOf(PointTo(
				MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("server.webhooks.port"),
				}),
			)))
		})

		It("should forbid empty Webhooks TLS ServerCertDir", func() {
			conf.Server.Webhooks.TLS.ServerCertDir = ""

			errs := ValidateDikiOperatorConfiguration(conf)
			Expect(errs).To(ConsistOf(PointTo(
				MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("server.webhooks.tls.serverCertDir"),
				}),
			)))
		})
	})
})
