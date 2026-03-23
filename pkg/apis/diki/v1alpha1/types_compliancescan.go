// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster,path=compliancescans,shortName=cscan,singular=compliancescan
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Current phase of the compliance scan"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Creation timestamp"

// ComplianceScan describes a compliance scan.
type ComplianceScan struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec contains the specification of this compliance scan.
	Spec ComplianceScanSpec `json:"spec,omitempty"`
	// Status contains the status of this compliance scan.
	Status ComplianceScanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComplianceScanList describes a list of compliance scans.
type ComplianceScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items contains the list of ComplianceScans.
	Items []ComplianceScan `json:"items"`
}

// ComplianceScanSpec is the specification of a ComplianceScan.
type ComplianceScanSpec struct {
	// Rulesets describe the rulesets to be applied during the compliance scan.
	Rulesets []RulesetConfig `json:"rulesets,omitempty"`
}

// RulesetConfig describes the configuration of a ruleset.
type RulesetConfig struct {
	// ID is the identifier of the ruleset.
	ID string `json:"id"`
	// Version is the version of the ruleset.
	Version string `json:"version"`
	// Options are options for a ruleset.
	// +optional
	Options *RulesetOptions `json:"options,omitempty"`
}

// RulesetOptions are options for a ruleset.
type RulesetOptions struct {
	// Ruleset contains global options for the ruleset.
	// +optional
	Ruleset *Options `json:"ruleset,omitempty"`
	// Rules contains references to rule options.
	// Users can use these to configure the behaviour of specific rules.
	// +optional
	Rules *Options `json:"rules,omitempty"`
}

// Options contains references to options.
type Options struct {
	// ConfigMapRef is a reference to a ConfigMap containing options.
	// +optional
	ConfigMapRef *OptionsConfigMapRef `json:"configMapRef,omitempty"`
}

// OptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.
type OptionsConfigMapRef struct {
	// Name is the name of the ConfigMap.
	Name string `json:"name"`
	// Namespace is the namespace of the ConfigMap.
	Namespace string `json:"namespace"`
	// Key is the key within the ConfigMap, where the options are stored.
	// +optional
	Key *string `json:"key,omitempty"`
}

// ComplianceScanStatus contains the status of a ComplianceScan.
type ComplianceScanStatus struct {
	// Conditions contains the conditions of the ComplianceScan.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// Phase represents the current phase of the ComplianceScan.
	Phase ComplianceScanPhase `json:"phase"`
	// Rulesets contains the ruleset summaries of the ComplianceScan.
	// +optional
	Rulesets []RulesetSummary `json:"rulesets,omitempty"`
}

// ComplianceScanPhase is an alias for string representing the phase of a ComplianceScan.
type ComplianceScanPhase string

const (
	// ComplianceScanPending means that the ComplianceScan is pending execution.
	ComplianceScanPending ComplianceScanPhase = "Pending"
	// ComplianceScanRunning means that the ComplianceScan is running.
	ComplianceScanRunning ComplianceScanPhase = "Running"
	// ComplianceScanCompleted means that the ComplianceScan has completed successfully.
	ComplianceScanCompleted ComplianceScanPhase = "Completed"
	// ComplianceScanFailed means that the ComplianceScan has failed.
	ComplianceScanFailed ComplianceScanPhase = "Failed"
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
	// +optional
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
	// +optional
	Failed []Rule `json:"failed,omitempty"`
	// Errored contains information about the rules that have an Errored status.
	// +optional
	Errored []Rule `json:"errored,omitempty"`
	// Warning contains information about the rules that have a Warning status.
	// +optional
	Warning []Rule `json:"warning,omitempty"`
}

// Rule contains information about the ID and the name of the rule that contains the findings.
type Rule struct {
	// ID is the unique identifier of the rule which contains the finding.
	ID string `json:"id"`
	// Name is the name of the rule which contains the finding.
	Name string `json:"name"`
}

// Condition describes a condition of a ComplianceScan.
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
	// ConditionTypeCompleted indicates whether the ComplianceScan has completed.
	ConditionTypeCompleted ConditionType = "Completed"
	// ConditionTypeFailed indicates whether the ComplianceScan has failed.
	ConditionTypeFailed ConditionType = "Failed"
)
