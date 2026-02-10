// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition described a condition of a ComplianceRun.
type Condition struct {
	// Type of condition.
	Type ConditionType `json:"type"`
	// Status of the condition.
	Status ConditionStatus `json:"status"`
	// Last time the condition was updated.
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
	// ConditionUnknown means diki-operator cannot decide if a resource is in the condition or not.
	ConditionUnknown ConditionStatus = "Unknown"

	// ConditionTypeCompleted indicates whether the ComplianceRun has completed.
	ConditionTypeCompleted ConditionType = "Completed"
	// ConditionTypeFailed indicates whether the ComplianceRun has failed.
	ConditionTypeFailed ConditionType = "Failed"
)
