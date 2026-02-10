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
	// RuleOptions describes the rule options for a ruleset.
	RuleOptions *RuleOptions
}

// RuleOptions describes the rule options for a ruleset.
type RuleOptions struct {
	// ConfigMapRef references a ConfigMap containing rule options for the ruleset.
	ConfigMapRef *RuleOptionsConfigMapRef
}

// RuleOptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.
type RuleOptionsConfigMapRef struct {
	// Name is the name of the ConfigMap.
	Name string
	// Namespace is the namespace of the ConfigMap.
	Namespace string
	// Key is the key in the ConfigMap.
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
	// Summary contains information about the amount of rules per each status.
	Summary Summary
	// Findings contains information about the specific rules that have errored/warned/failed
	Findings *Findings
}

// Summary contains information about the amount of rules per each status.
type Summary struct {
	// Passed counts the amount of rules in a specific ruleset that have passed.
	Passed int
	// Skipped counts the amount of rules in a specific ruleset that have been skipped.
	Skipped int
	// Accepted counts the amount of rules in a specific ruleset that have been accepted.
	Accepted int
	// Warning counts the amount of rules in a specific ruleset that have returned a warning.
	Warning int
	// Failed counts the amount of rules in a specific ruleset that have failed.
	Failed int
	// Errored counts the amount of rules in a specific ruleset that have errored.
	Errored int
}

// Findings contains information about the specific rules that have errored/warned/failed.
type Findings struct {
	// Failed contains information about the rules that contain a Failed checkResult.
	Failed []Rule
	// Errored contains information about the rules that contain a Errored checkResult.
	Errored []Rule
	// Warning contains information about the rules that contain a Warning checkResult.
	Warning []Rule
}

// Rule contains information about the ID and the name of the rule that contains the findings.
type Rule struct {
	// ID is the unique identifier of the rule which contains the finding.
	ID string
	// Name is name of the rule which contains the finding.
	Name string
}
