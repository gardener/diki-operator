// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"time"

	"github.com/gardener/gardener/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_DikiOperatorConfiguration sets defaults for the configuration of the diki operator.
func SetDefaults_DikiOperatorConfiguration(obj *DikiOperatorConfiguration) {
	if obj.LogLevel == "" {
		obj.LogLevel = logger.InfoLevel
	}
	if obj.LogFormat == "" {
		obj.LogFormat = logger.FormatJSON
	}
	if obj.LeaderElection == nil {
		obj.LeaderElection = &componentbaseconfigv1alpha1.LeaderElectionConfiguration{}
	}
}

// SetDefaults_ComplianceRunConfig sets defaults for the ComplianceRunConfig object.
func SetDefaults_ComplianceRunConfig(obj *ComplianceRunConfig) {
	if obj.SyncPeriod == nil {
		obj.SyncPeriod = &metav1.Duration{Duration: time.Hour}
	}
}

// SetDefaults_DikiRunnerConfig sets defaults for the DikiRunnerConfig object.
func SetDefaults_DikiRunnerConfig(obj *DikiRunnerConfig) {
	if obj.Namespace == "" {
		obj.Namespace = DefaultDikiRunnerNamespace
	}
	if obj.WaitInterval == nil {
		obj.WaitInterval = &metav1.Duration{Duration: DefaultWaitInterval}
	}
	if obj.ExecTimeout == nil {
		obj.ExecTimeout = &metav1.Duration{Duration: DefaultExecTimeout}
	}
	if obj.PodCompletionTimeout == nil {
		obj.PodCompletionTimeout = &metav1.Duration{Duration: DefaultPodCompletionTimeout}
	}
}

// SetDefaults_ServerConfiguration sets defaults for the ServerConfiguration object.
func SetDefaults_ServerConfiguration(obj *ServerConfiguration) {
	if obj.HealthProbes == nil {
		obj.HealthProbes = &Server{}
	}
	if obj.HealthProbes.Port == 0 {
		obj.HealthProbes.Port = 8081
	}
	if obj.Metrics == nil {
		obj.Metrics = &Server{}
	}
	if obj.Metrics.Port == 0 {
		obj.Metrics.Port = 8080
	}
}

// SetDefaults_LeaderElectionConfiguration sets defaults for the LeaderElectionConfiguration object.
func SetDefaults_LeaderElectionConfiguration(obj *componentbaseconfigv1alpha1.LeaderElectionConfiguration) {
	if obj.ResourceLock == "" {
		obj.ResourceLock = "leases"
	}

	componentbaseconfigv1alpha1.RecommendedDefaultLeaderElectionConfiguration(obj)

	if obj.ResourceNamespace == "" {
		obj.ResourceNamespace = DefaultLockObjectNamespace
	}
	if obj.ResourceName == "" {
		obj.ResourceName = DefaultLockObjectName
	}
}
