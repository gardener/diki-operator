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
	// Options are options for a ruleset.
	Options *RulesetOptions `json:"options,omitempty"`
}

// RulesetOptions are options for a ruleset.
type RulesetOptions struct {
	// Ruleset contains global options for the ruleset.
	Ruleset *Options `json:"ruleset,omitempty"`
	// Rules contains references to rule options.
	// Users can use these to configure the behaviour of specific rules.
	Rules *Options `json:"rules,omitempty"`
}

// Options contains references to options.
type Options struct {
	// ConfigMapRef is a reference to a ConfigMap containing options.
	ConfigMapRef *OptionsConfigMapRef `json:"configMapRef,omitempty"`
}

// OptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.
type OptionsConfigMapRef struct {
	// Name is the name of the ConfigMap.
	Name string `json:"name"`
	// Namespace is the namespace of the ConfigMap.
	Namespace string `json:"namespace"`
	// Key is the key within the ConfigMap, where the options are stored.
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
	// Results contains the results of the ruleset.
	Results RulesResults `json:"results"`
}

// RulesResults contains the results of the rules in a ruleset.
type RulesResults struct {
	// Summary contains information about the amount of rules per each status.
	Summary RulesSummary `json:"summary"`
	// Rules contains information about the specific rules that have errored/warned/failed.
	Rules *RulesFindings `json:"rules,omitempty"`
}

// RulesSummary contains information about the amount of rules per each status.
type RulesSummary struct {
	// Passed counts the amount of rules in a specific ruleset that have passed.
	Passed int32 `json:"passed"`
	// Skipped counts the amount of rules in a specific ruleset that have been skipped.
	Skipped int32 `json:"skipped"`
	// Accepted counts the amount of rules in a specific ruleset that have been accepted.
	Accepted int32 `json:"accepted"`
	// Warning counts the amount of rules in a specific ruleset that have returned a warning.
	Warning int32 `json:"warning"`
	// Failed counts the amount of rules in a specific ruleset that have failed.
	Failed int32 `json:"failed"`
	// Errored counts the amount of rules in a specific ruleset that have errored.
	Errored int32 `json:"errored"`
}

// RulesFindings contains information about the specific rules that have errored/warned/failed.
type RulesFindings struct {
	// Failed contains information about the rules that have a Failed status.
	Failed []Rule `json:"failed,omitempty"`
	// Errored contains information about the rules that have an Errored status.
	Errored []Rule `json:"errored,omitempty"`
	// Warning contains information about the rules that have a Warning status.
	Warning []Rule `json:"warning,omitempty"`
}

// Rule contains information about the ID and the name of the rule that contains the findings.
type Rule struct {
	// ID is the unique identifier of the rule which contains the finding.
	ID string `json:"id"`
	// Name is the name of the rule which contains the finding.
	Name string `json:"name"`
}

// Condition describes a condition of a ComplianceRun.
type Condition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type"`
	// Status is the status of the condition.
	Status ConditionStatus `json:"status"`
	// LastUpdateTime is the last time the condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	// LastTransitionTime is the last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is a brief reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is a human-readable message indicating details about the last transition.
	Message string `json:"message"`
}

// ConditionStatus is an alias for string representing the status of a condition.
type ConditionStatus string

// ConditionType is an alias for string representing the type of a condition.
type ConditionType string

const (
	// ConditionTrue means a resource is in the condition.
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse means a resource is not in the condition.
	ConditionFalse ConditionStatus = "False"
	// ConditionUnknown means that it cannot be decided if a resource is in the condition or not.
	ConditionUnknown ConditionStatus = "Unknown"

	// ConditionTypeCompleted indicates whether the ComplianceRun has completed.
	ConditionTypeCompleted ConditionType = "Completed"
	// ConditionTypeFailed indicates whether the ComplianceRun has failed.
	ConditionTypeFailed ConditionType = "Failed"
)
