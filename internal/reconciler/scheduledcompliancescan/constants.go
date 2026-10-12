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
	// AnnotationScheduledComplianceScanTriggerNow is the annotation used to trigger an immediate
	// ComplianceScan creation outside of the normal schedule.
	// The annotation is removed after being processed to prevent re-triggering.
	AnnotationScheduledComplianceScanTriggerNow = "scheduledcompliancescan.diki.gardener.cloud/trigger-now"
)
