// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster,path=complianceruns,shortName=crun,singular=compliancerun
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Current phase of the compliance run"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Creation timestamp"

// ComplianceRun describes a compliance run.
type ComplianceRun struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec contains the specification of this compliance run.
	Spec ComplianceRunSpec `json:"spec,omitempty"`
	// Status contains the status of this compliance run.
	Status ComplianceRunStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComplianceRunList describes a list of compliance runs.
type ComplianceRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items contains the list of ComplianceRuns.
	Items []ComplianceRun `json:"items"`
}

// ComplianceRunSpec is the specification of a ComplianceRun.
type ComplianceRunSpec struct {
	// Rulesets describe the rulesets to be applied during the compliance run.
	Rulesets []RulesetConfig `json:"rulesets,omitempty"`
}

// RulesetConfig describes the configuration of a ruleset.
type RulesetConfig struct {
	// ID is the identifier of the ruleset.
	ID string `json:"id"`
	// Version is the version of the ruleset.
	Version string `json:"version"`
	// RuleOptions describes the rule options for a ruleset.
	RuleOptions *RuleOptions `json:"ruleOptions,omitempty"`
}

// RuleOptions describes the rule options for a ruleset.
type RuleOptions struct {
	// ConfigMapRef references a ConfigMap containing rule options for the ruleset.
	ConfigMapRef *RuleOptionsConfigMapRef `json:"configMapRef,omitempty"`
}

// RuleOptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.
type RuleOptionsConfigMapRef struct {
	// Name is the name of the ConfigMap.
	Name string `json:"name"`
	// Namespace is the namespace of the ConfigMap.
	Namespace string `json:"namespace"`
	// Key is the key in the ConfigMap.
	Key *string `json:"key,omitempty"`
}

// ComplianceRunStatus contains the status of a ComplianceRun.
type ComplianceRunStatus struct {
	// Conditions contains the conditions of the ComplianceRun.
	Conditions []Condition `json:"conditions,omitempty"`
	// Phase represents the current phase of the ComplianceRun.
	Phase ComplianceRunPhase `json:"phase"`
	// Rulesets contains the ruleset summaries of the ComplianceRun.
	Rulesets []RulesetSummary `json:"rulesets,omitempty"`
}

// ComplianceRunPhase is an alias for string representing the phase of a ComplianceRun.
type ComplianceRunPhase string

const (
	// ComplianceRunPending means that the ComplianceRun is pending execution.
	ComplianceRunPending ComplianceRunPhase = "Pending"
	// ComplianceRunRunning means that the ComplianceRun is running.
	ComplianceRunRunning ComplianceRunPhase = "Running"
	// ComplianceRunCompleted means that the ComplianceRun has completed successfully.
	ComplianceRunCompleted ComplianceRunPhase = "Completed"
	// ComplianceRunFailed means that the ComplianceRun has failed.
	ComplianceRunFailed ComplianceRunPhase = "Failed"
)

// RulesetSummary contains the identifiers and the summary for a specific ruleset.
type RulesetSummary struct {
	// ID is the identifier of the ruleset that is summarized.
	ID string `json:"id"`
	// Version is the version of the ruleset that is summarized.
	Version string `json:"version"`
	// Summary contains information about the amount of rules per each status.
	Summary Summary `json:"summary"`
	// Findings contains information about the specific rules that have errored/warned/failed
	Findings *Findings `json:"findings,omitempty"`
}

// Summary contains information about the amount of rules per each status.
type Summary struct {
	// Passed counts the amount of rules in a specific ruleset that have passed.
	Passed int `json:"passed"`
	// Skipped counts the amount of rules in a specific ruleset that have been skipped.
	Skipped int `json:"skipped"`
	// Accepted counts the amount of rules in a specific ruleset that have been accepted.
	Accepted int `json:"accepted"`
	// Warning counts the amount of rules in a specific ruleset that have returned a warning.
	Warning int `json:"warning"`
	// Failed counts the amount of rules in a specific ruleset that have failed.
	Failed int `json:"failed"`
	// Errored counts the amount of rules in a specific ruleset that have errored.
	Errored int `json:"errored"`
}

// Findings contains information about the specific rules that have errored/warned/failed.
type Findings struct {
	// Failed contains information about the rules that contain a Failed checkResult.
	Failed []Rule `json:"failed,omitempty"`
	// Errored contains information about the rules that contain a Errored checkResult.
	Errored []Rule `json:"errored,omitempty"`
	// Warning contains information about the rules that contain a Warning checkResult.
	Warning []Rule `json:"warning,omitempty"`
}

// Rule contains information about the ID and the name of the rule that contains the findings.
type Rule struct {
	// ID is the unique identifier of the rule which contains the finding.
	ID string `json:"ruleID"`
	// Name is name of the rule which contains the finding.
	Name string `json:"ruleName"`
}
