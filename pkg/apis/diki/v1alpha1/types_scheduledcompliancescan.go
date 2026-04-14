// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster,path=scheduledcompliancescans,shortName=scscan,singular=scheduledcompliancescan
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`,description="Cron schedule of the compliance scan"
// +kubebuilder:printcolumn:name="Active",type=string,JSONPath=`.status.active.name`,description="Name of the currently active ComplianceScan"
// +kubebuilder:printcolumn:name="Last Schedule",type=date,JSONPath=`.status.lastScheduleTime`,description="Last time a ComplianceScan was scheduled"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Creation timestamp"

// ScheduledComplianceScan describes a scheduled compliance scan.
type ScheduledComplianceScan struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec contains the specification of this scheduled compliance scan.
	Spec ScheduledComplianceScanSpec `json:"spec,omitempty"`
	// Status contains the status of this scheduled compliance scan.
	Status ScheduledComplianceScanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduledComplianceScanList describes a list of scheduled compliance scans.
type ScheduledComplianceScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items contains the list of ScheduledComplianceScans.
	Items []ScheduledComplianceScan `json:"items"`
}

// ScheduledComplianceScanSpec is the specification of a ScheduledComplianceScan.
type ScheduledComplianceScanSpec struct {
	// Schedule is a cron expression defining when the compliance scan should run.
	// +optional
	// +kubebuilder:default="0 0 * * 0"
	// +kubebuilder:validation:MinLength=9
	Schedule string `json:"schedule,omitempty"`
	// SuccessfulScansHistoryLimit is the number of completed compliance scans to keep.
	// +optional
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=0
	SuccessfulScansHistoryLimit *int32 `json:"successfulScansHistoryLimit,omitempty"`
	// FailedScansHistoryLimit is the number of failed compliance scans to keep.
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	FailedScansHistoryLimit *int32 `json:"failedScansHistoryLimit,omitempty"`
	// ScanTemplate is the template for the ComplianceScan that will be created on each scheduled scan.
	ScanTemplate ScheduledComplianceScanTemplate `json:"scanTemplate"`
}

// ScheduledComplianceScanTemplate is the template for the ComplianceScan that will be created.
type ScheduledComplianceScanTemplate struct {
	// Spec is the spec of the ComplianceScan that will be created.
	Spec ComplianceScanSpec `json:"spec"`
}

// ScheduledComplianceScanStatus contains the status of a ScheduledComplianceScan.
type ScheduledComplianceScanStatus struct {
	// Active is a reference to the currently active ComplianceScan, if any.
	// +optional
	Active *corev1.ObjectReference `json:"active,omitempty"`
	// LastScheduleTime is the last time a ComplianceScan was scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
	// LastCompletionTime is the last time a scheduled ComplianceScan completed.
	// +optional
	LastCompletionTime *metav1.Time `json:"lastCompletionTime,omitempty"`
}
