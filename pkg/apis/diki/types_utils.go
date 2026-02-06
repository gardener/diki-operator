// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package diki

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition described a condition of a ComplianceRun.
type Condition struct {
	// Type of condition.
	Type ConditionType
	// Status of the condition.
	Status ConditionStatus
	// Last time the condition was updated.
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
	// ConditionTypeCompleted indicates whether the ComplianceRun has completed.
	ConditionTypeCompleted ConditionType = "Completed"
	// ConditionTypeFailed indicates whether the ComplianceRun has failed.
	ConditionTypeFailed ConditionType = "Failed"
)
