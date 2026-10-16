// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

const (
	// LabelScheduledComplianceScanName is the label used to identify ComplianceScans
	// created by a specific ScheduledComplianceScan by name.
	LabelScheduledComplianceScanName = "scheduledcompliancescan.diki.gardener.cloud/name"
	// LabelScheduledComplianceScanUID is the label used to identify ComplianceScans
	// created by a specific ScheduledComplianceScan by UID.
	LabelScheduledComplianceScanUID = "scheduledcompliancescan.diki.gardener.cloud/uid"

	// AnnotationScheduledComplianceScanOperation is the annotation used to request an operation
	// on a ScheduledComplianceScan.
	AnnotationScheduledComplianceScanOperation = "scheduledcompliancescan.diki.gardener.cloud/operation"
	// ScheduledComplianceScanOperationTrigger is a constant for the value of the operation annotation describing
	// a trigger of ComplianceScan creation outside of the normal schedule.
	ScheduledComplianceScanOperationTrigger = "trigger"
)
