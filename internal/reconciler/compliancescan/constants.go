// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import "time"

const (
	// ReconciliationRequeueInterval is the time window between different reconciliations of a running ComplianceScan.
	ReconciliationRequeueInterval = time.Second * 5

	// ConfigMapNamePrefix is the prefix for diki config ConfigMap names.
	ConfigMapNamePrefix = "diki-config-"
	// ServiceAccountNameDikiRunner is the name for the diki-run Job related ServiceAccount.
	ServiceAccountNameDikiRunner = "diki-runner"
	// JobNamePrefix is the prefix for the diki-run Job names.
	JobNamePrefix = "diki-run-"
	// DikiConfigVolumeName is the name of the volume mounted in the diki-run Job pods.
	DikiConfigVolumeName = "diki-config"
	// DikiConfigKey is the key used to store the YAML configuration in the ConfigMap data.
	DikiConfigKey = "config.yaml"
	// DikiConfigMountPath is the mount path for the configurations needed by the diki-run Job pod.
	DikiConfigMountPath = "/config"
	// DikiScanContainerName is the name of the container that is used for executing the scan within the diki-run Job pod.
	DikiScanContainerName = "diki-scan"

	// RuleOptionsSuffix is the suffix appended to ruleset IDs when looking up rule options in ConfigMaps.
	RuleOptionsSuffix = "-rules"

	// ConditionReasonRunning is the reason for ComplianceScan condition when it is running.
	ConditionReasonRunning = "ComplianceScanRunning"
	// ConditionReasonCompleted is the reason for ComplianceScan condition when it has completed successfully.
	ConditionReasonCompleted = "ComplianceScanCompleted"
	// ConditionReasonFailed is the reason for ComplianceScan condition when it has failed.
	ConditionReasonFailed = "ComplianceScanFailed"
)
