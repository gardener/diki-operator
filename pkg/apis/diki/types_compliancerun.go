// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package diki

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComplianceRun describes a compliance run.
type ComplianceRun struct {
	metav1.TypeMeta
	// Standard object metadata.
	metav1.ObjectMeta

	// Spec contains the specification of this compliance run.
	Spec ComplianceRunSpec
	// Status contains the status of this compliance run.
	Status ComplianceRunStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComplianceRunList describes a list of compliance runs.
type ComplianceRunList struct {
	metav1.TypeMeta
	metav1.ListMeta

	// Items contains the list of ComplianceRuns.
	Items []ComplianceRun
}

// ComplianceRunSpec is the specification of a ComplianceRun.
type ComplianceRunSpec struct {
	// Rulesets describe the rulesets to be applied during the compliance run.
	Rulesets []RulesetConfig
}

// RulesetConfig describes the configuration of a ruleset.
type RulesetConfig struct {
	// ID is the identifier of the ruleset.
	ID string
	// Version is the version of the ruleset.
	Version string
	// Options are options for a ruleset.
	Options *RulesetOptions
}

// RulesetOptions are options for a ruleset.
type RulesetOptions struct {
	// Ruleset contains global options for the ruleset.
	Ruleset *Options
	// Rules contains references to rule options.
	// Users can use these to configure the behaviour of specific rules.
	Rules *Options
}

// Options contains references to options.
type Options struct {
	// ConfigMapRef is a reference to a ConfigMap containing options.
	ConfigMapRef *OptionsConfigMapRef
}

// OptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.
type OptionsConfigMapRef struct {
	// Name is the name of the ConfigMap.
	Name string
	// Namespace is the namespace of the ConfigMap.
	Namespace string
	// Key is the key within the ConfigMap, where the options are stored.
	Key *string
}

// ComplianceRunStatus contains the status of a ComplianceRun.
type ComplianceRunStatus struct {
	// Conditions contains the conditions of the ComplianceRun.
	Conditions []Condition
	// Phase represents the current phase of the ComplianceRun.
	Phase ComplianceRunPhase
	// Rulesets contains the ruleset summaries of the ComplianceRun.
	Rulesets []RulesetSummary
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
	ID string
	// Version is the version of the ruleset that is summarized.
	Version string
	// Results contains the results of the ruleset.
	Results RulesResults
}

// RulesResults contains the results of the rules in a ruleset.
type RulesResults struct {
	// Summary contains information about the amount of rules per each status.
	Summary RulesSummary
	// Rules contains information about the specific rules that have errored/warned/failed.
	Rules *RulesFindings
}

// RulesSummary contains information about the amount of rules per each status.
type RulesSummary struct {
	// Passed counts the amount of rules in a specific ruleset that have passed.
	Passed int32
	// Skipped counts the amount of rules in a specific ruleset that have been skipped.
	Skipped int32
	// Accepted counts the amount of rules in a specific ruleset that have been accepted.
	Accepted int32
	// Warning counts the amount of rules in a specific ruleset that have returned a warning.
	Warning int32
	// Failed counts the amount of rules in a specific ruleset that have failed.
	Failed int32
	// Errored counts the amount of rules in a specific ruleset that have errored.
	Errored int32
}

// RulesFindings contains information about the specific rules that have errored/warned/failed.
type RulesFindings struct {
	// Failed contains information about the rules that have a Failed status.
	Failed []Rule
	// Errored contains information about the rules that have an Errored status.
	Errored []Rule
	// Warning contains information about the rules that have a Warning status.
	Warning []Rule
}

// Rule contains information about the ID and the name of the rule that contains the findings.
type Rule struct {
	// ID is the unique identifier of the rule which contains the finding.
	ID string
	// Name is the name of the rule which contains the finding.
	Name string
}

// Condition describes a condition of a ComplianceRun.
type Condition struct {
	// Type is the type of the condition.
	Type ConditionType
	// Status is the status of the condition.
	Status ConditionStatus
	// LastUpdateTime is the last time the condition was updated.
	LastUpdateTime metav1.Time
	// LastTransitionTime is the last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time
	// Reason is a brief reason for the condition's last transition.
	Reason string
	// Message is a human-readable message indicating details about the last transition.
	Message string
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
