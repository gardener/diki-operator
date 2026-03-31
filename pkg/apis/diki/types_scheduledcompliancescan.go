// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package diki

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduledComplianceScan describes a scheduled compliance scan.
type ScheduledComplianceScan struct {
	metav1.TypeMeta
	// Standard object metadata.
	metav1.ObjectMeta

	// Spec contains the specification of this scheduled compliance scan.
	Spec ScheduledComplianceScanSpec
	// Status contains the status of this scheduled compliance scan.
	Status ScheduledComplianceScanStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduledComplianceScanList describes a list of scheduled compliance scans.
type ScheduledComplianceScanList struct {
	metav1.TypeMeta
	metav1.ListMeta

	// Items contains the list of ScheduledComplianceScans.
	Items []ScheduledComplianceScan
}

// ScheduledComplianceScanSpec is the specification of a ScheduledComplianceScan.
type ScheduledComplianceScanSpec struct {
	// Schedule is a cron expression defining when the compliance scan should run.
	Schedule string
	// ScansHistoryLimit is the number of completed compliance scans to keep.
	ScansHistoryLimit *int32
	// ScanTemplate is the template for the ComplianceScan that will be created on each scheduled scan.
	ScanTemplate ScheduledComplianceScanTemplate
}

// ScheduledComplianceScanTemplate is the template for the ComplianceScan that will be created.
type ScheduledComplianceScanTemplate struct {
	// Spec is the spec of the ComplianceScan that will be created.
	Spec ComplianceScanSpec
}

// ScheduledComplianceScanStatus contains the status of a ScheduledComplianceScan.
type ScheduledComplianceScanStatus struct {
	// Active is a reference to the currently active ComplianceScan, if any.
	Active *corev1.ObjectReference
	// LastScheduleTime is the last time a ComplianceScan was scheduled.
	LastScheduleTime *metav1.Time
	// LastCompletionTime is the last time a scheduled ComplianceScan completed.
	LastCompletionTime *metav1.Time
}
