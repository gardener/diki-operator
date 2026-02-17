// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

const (
	// DefaultLockObjectNamespace is the default lock namespace for leader election.
	DefaultLockObjectNamespace = "kube-system"
	// DefaultLockObjectName is the default lock name for leader election.
	DefaultLockObjectName = "diki-operator-leader-election"
	// DefaultDikiRunnerNamespace is the default namespace where DikiRunner pods are created.
	DefaultDikiRunnerNamespace = "kube-system"
	// DefaultWaitInterval is the default duration to wait between status checks.
	DefaultWaitInterval = 5 * time.Second
	// DefaultPodCompletionTimeout is the default maximum duration to wait for pod completion.
	DefaultPodCompletionTimeout = 10 * time.Minute
	// DefaultExecTimeout is the default maximum duration to wait for exec commands.
	DefaultExecTimeout = 30 * time.Second
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DikiOperatorConfiguration defines the configuration for the diki-operator.
type DikiOperatorConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// LeaderElection defines the configuration of leader election client.
	// +optional
	LeaderElection *componentbaseconfigv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// LogLevel is the level/severity for the logs. Must be one of [info,debug,error].
	LogLevel string `json:"logLevel"`
	// LogFormat is the output format for the logs. Must be one of [text,json].
	LogFormat string `json:"logFormat"`
	// Controllers defines the configuration of the controllers.
	Controllers ControllerConfiguration `json:"controllers"`
	// Server defines the configuration of the HTTP server.
	Server ServerConfiguration `json:"server"`
}

// ControllerConfiguration defines the configuration of the controllers.
type ControllerConfiguration struct {
	// ComplianceRun is the configuration for the compliance run controller.
	ComplianceRun ComplianceRunConfig `json:"complianceRun"`
}

// ComplianceRunConfig contains configuration for the ComplianceRun controller.
type ComplianceRunConfig struct {
	// SyncPeriod is the duration how often the controller performs its reconciliation.
	// +optional
	SyncPeriod *metav1.Duration `json:"syncPeriod,omitempty"`
	// DikiRunner is the configuration for the DikiRunner.
	// +optional
	DikiRunner DikiRunnerConfig `json:"dikiRunner,omitempty"`
}

// DikiRunnerConfig contains configuration for the DikiRunner.
type DikiRunnerConfig struct {
	// Namespace is the namespace where DikiRunner pods are created.
	Namespace string `json:"namespace"`
	// Labels are the labels to be added to DikiRunner pods.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// WaitInterval is the duration to wait between status checks for pod creation and completion.
	// +optional
	WaitInterval *metav1.Duration `json:"waitInterval,omitempty"`
	// PodCompletionTimeout is the maximum duration to wait for a DikiRunner pod to complete.
	// +optional
	PodCompletionTimeout *metav1.Duration `json:"podCompletionTimeout,omitempty"`
	// ExecTimeout is the maximum duration to wait for a pod exec command to complete.
	// +optional
	ExecTimeout *metav1.Duration `json:"execTimeout,omitempty"`
}

// ServerConfiguration contains details for the HTTP(S) servers.
type ServerConfiguration struct {
	// HealthProbes is the configuration for serving the healthz and readyz endpoints.
	// +optional
	HealthProbes *Server `json:"healthProbes,omitempty"`
	// Metrics is the configuration for serving the metrics endpoint.
	// +optional
	Metrics *Server `json:"metrics,omitempty"`
}

// Server contains information for HTTP(S) server configuration.
type Server struct {
	// Port is the port on which to serve requests.
	Port int `json:"port"`
	// BindAddress is the IP address on which to listen for the specified port.
	BindAddress string `json:"bindAddress"`
}
