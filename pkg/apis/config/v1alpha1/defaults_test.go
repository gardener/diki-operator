// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"

	. "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener/pkg/logger"
)

var _ = Describe("Defaults", func() {
	Describe("#SetDefaults_DikiOperatorConfiguration", func() {
		var obj *DikiOperatorConfiguration

		BeforeEach(func() {
			obj = &DikiOperatorConfiguration{}
		})

		Context("LogLevel", func() {
			It("should default log level", func() {
				SetDefaults_DikiOperatorConfiguration(obj)

				Expect(obj.LogLevel).To(Equal(logger.InfoLevel))
			})

			It("should not overwrite already set value for log level", func() {
				obj.LogLevel = "warning"

				SetDefaults_DikiOperatorConfiguration(obj)

				Expect(obj.LogLevel).To(Equal("warning"))
			})
		})

		Context("LogFormat", func() {
			It("should default log format", func() {
				SetDefaults_DikiOperatorConfiguration(obj)

				Expect(obj.LogFormat).To(Equal(logger.FormatJSON))
			})

			It("should not overwrite already set value for log format", func() {
				obj.LogFormat = "md"

				SetDefaults_DikiOperatorConfiguration(obj)

				Expect(obj.LogFormat).To(Equal("md"))
			})
		})

		Context("LeaderElection", func() {
			It("should initialize LeaderElection when nil", func() {
				SetDefaults_DikiOperatorConfiguration(obj)

				Expect(obj.LeaderElection).NotTo(BeNil())
			})
		})
	})

	Describe("#SetDefaults_ComplianceRunConfig", func() {
		var obj *ComplianceRunConfig

		BeforeEach(func() {
			obj = &ComplianceRunConfig{}
		})

		Context("SyncPeriod", func() {
			It("should default sync period", func() {
				SetDefaults_ComplianceRunConfig(obj)

				Expect(obj.SyncPeriod).To(Equal(&metav1.Duration{Duration: time.Hour}))
			})

			It("should not overwrite already set value for sync period", func() {
				obj.SyncPeriod = &metav1.Duration{Duration: time.Minute}

				SetDefaults_ComplianceRunConfig(obj)

				Expect(obj.SyncPeriod).To(Equal(&metav1.Duration{Duration: time.Minute}))
			})
		})
	})

	Describe("#SetDefaults_DikiRunnerConfig", func() {
		var obj *DikiRunnerConfig

		BeforeEach(func() {
			obj = &DikiRunnerConfig{}
		})

		Context("Namespace", func() {
			It("should default namespace", func() {
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.Namespace).To(Equal("kube-system"))
			})

			It("should not overwrite already set value for namespace", func() {
				obj.Namespace = "default"
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.Namespace).To(Equal("default"))
			})
		})

		Context("WaitInterval", func() {
			It("should default wait interval", func() {
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.WaitInterval).To(Equal(&metav1.Duration{Duration: 5 * time.Second}))
			})

			It("should not overwrite already set value for wait interval", func() {
				obj.WaitInterval = &metav1.Duration{Duration: 10 * time.Second}
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.WaitInterval).To(Equal(&metav1.Duration{Duration: 10 * time.Second}))
			})
		})

		Context("ExecTimeout", func() {
			It("should default exec timeout", func() {
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.ExecTimeout).To(Equal(&metav1.Duration{Duration: 30 * time.Second}))
			})

			It("should not overwrite already set value for exec timeout", func() {
				obj.ExecTimeout = &metav1.Duration{Duration: 10 * time.Second}
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.ExecTimeout).To(Equal(&metav1.Duration{Duration: 10 * time.Second}))
			})
		})

		Context("PodCompletionTimeout", func() {
			It("should default pod completion timeout", func() {
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.PodCompletionTimeout).To(Equal(&metav1.Duration{Duration: 10 * time.Minute}))
			})

			It("should not overwrite already set value for pod completion timeout", func() {
				obj.PodCompletionTimeout = &metav1.Duration{Duration: 10 * time.Second}
				SetDefaults_DikiRunnerConfig(obj)

				Expect(obj.PodCompletionTimeout).To(Equal(&metav1.Duration{Duration: 10 * time.Second}))
			})
		})
	})

	Describe("#SetDefaults_ServerConfiguration", func() {
		var obj *ServerConfiguration

		BeforeEach(func() {
			obj = &ServerConfiguration{}
		})

		Context("HealthProbes", func() {
			It("should default HealthProbes when nil", func() {
				SetDefaults_ServerConfiguration(obj)

				Expect(obj.HealthProbes).NotTo(BeNil())
				Expect(obj.HealthProbes.Port).To(Equal(8081))
			})
		})

		Context("Metrics", func() {
			It("should default Metrics when nil", func() {
				SetDefaults_ServerConfiguration(obj)

				Expect(obj.Metrics).NotTo(BeNil())
				Expect(obj.Metrics.Port).To(Equal(8080))
			})
		})

		Context("should not overwrite already set values", func() {
			It("should not overwrite already set HealthProbes", func() {
				obj.HealthProbes = &Server{Port: 9090}

				SetDefaults_ServerConfiguration(obj)

				Expect(obj.HealthProbes.Port).To(Equal(9090))
			})

			It("should not overwrite already set Metrics", func() {
				obj.Metrics = &Server{Port: 9092}

				SetDefaults_ServerConfiguration(obj)

				Expect(obj.Metrics.Port).To(Equal(9092))
			})
		})
	})

	Describe("#SetDefaults_LeaderElectionConfiguration", func() {
		var obj *componentbaseconfigv1alpha1.LeaderElectionConfiguration

		BeforeEach(func() {
			obj = &componentbaseconfigv1alpha1.LeaderElectionConfiguration{}
		})

		Context("should default to recommended leader election values", func() {
			It("should set default recommended leader election values", func() {
				SetDefaults_LeaderElectionConfiguration(obj)

				expectedLeaderElectionConfig := &componentbaseconfigv1alpha1.LeaderElectionConfiguration{
					LeaderElect:       ptr.To(true),
					LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
					RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
					RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
					ResourceLock:      "leases",
					ResourceName:      DefaultLockObjectName,
					ResourceNamespace: DefaultLockObjectNamespace,
				}
				Expect(obj).To(Equal(expectedLeaderElectionConfig))
			})

			It("should not overwrite already set values for leader election", func() {
				obj.LeaderElect = ptr.To(false)
				obj.LeaseDuration = metav1.Duration{Duration: 30 * time.Second}
				obj.RenewDeadline = metav1.Duration{Duration: 20 * time.Second}
				obj.RetryPeriod = metav1.Duration{Duration: 5 * time.Second}
				obj.ResourceLock = "lock"
				obj.ResourceName = "name"
				obj.ResourceNamespace = "namespace"

				SetDefaults_LeaderElectionConfiguration(obj)

				expectedLeaderElectionConfig := &componentbaseconfigv1alpha1.LeaderElectionConfiguration{
					LeaderElect:       ptr.To(false),
					LeaseDuration:     metav1.Duration{Duration: 30 * time.Second},
					RenewDeadline:     metav1.Duration{Duration: 20 * time.Second},
					RetryPeriod:       metav1.Duration{Duration: 5 * time.Second},
					ResourceLock:      "lock",
					ResourceName:      "name",
					ResourceNamespace: "namespace",
				}
				Expect(obj).To(Equal(expectedLeaderElectionConfig))
			})
		})
	})
})
