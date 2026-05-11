// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

const (
	// LabelComplianceScanName is the label used to identify resources
	// connected to a ComplianceScan by name.
	LabelComplianceScanName = "compliancescan.diki.gardener.cloud/name"
	// LabelComplianceScanUID is the label used to identify resources
	// connected to a ComplianceScan by UID.
	LabelComplianceScanUID = "compliancescan.diki.gardener.cloud/uid"

	// ConfigMapGenerateNamePrefix is the prefix for diki config ConfigMap names.
	ConfigMapGenerateNamePrefix = "diki-config-"
	// DikiConfigKey is the key used to store the YAML configuration in the ConfigMap data.
	DikiConfigKey = "config.yaml"

	// RuleOptionsSuffix is the suffix appended to ruleset IDs when looking up rule options in ConfigMaps.
	RuleOptionsSuffix = "-rules"

	// ConditionReasonRunning is the reason for ComplianceScan condition when it is running.
	ConditionReasonRunning = "ComplianceScanRunning"
	// ConditionReasonCompleted is the reason for ComplianceScan condition when it has completed successfully.
	ConditionReasonCompleted = "ComplianceScanCompleted"
	// ConditionReasonFailed is the reason for ComplianceScan condition when it has failed.
	ConditionReasonFailed = "ComplianceScanFailed"
)
