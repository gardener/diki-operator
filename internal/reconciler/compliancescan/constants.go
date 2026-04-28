// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import "time"

const (
	// ReconciliationRequeueInterval is the time window between different reconciliations of a running ComplianceScan.
	ReconciliationRequeueInterval = time.Second * 5

	// ComplianceScanLabel is the label used to identify resources connected to a ComplianceScan.
	ComplianceScanLabel = "diki.gardener.cloud/compliancescan"

	// LabelAppName is the standard Kubernetes label key for application name.
	LabelAppName = "app.kubernetes.io/name"
	// LabelAppManagedBy is the standard Kubernetes label key for the managing tool or operator.
	LabelAppManagedBy = "app.kubernetes.io/managed-by"

	// LabelValueDiki is the application name value used for diki-related resources.
	LabelValueDiki = "diki"
	// LabelValueDikiOperator is the managing operator value used for diki-operator managed resources.
	LabelValueDikiOperator = "diki-operator"

	// ConfigMapName is the name of the volume mounted in the diki-run Job pods.
	ConfigMapName = "diki-config"
	// ConfigMapNamePrefix is the prefix for diki config ConfigMap names.
	ConfigMapNamePrefix = "diki-config-"
	// JobName is the name for the diki-run Job-related resources (ServiceAccounts, ClusterRoles, etc.)
	JobName = "diki-run"
	// JobNamePrefix is the prefix for the diki-run Job names.
	JobNamePrefix = "diki-run-"
	// DikiConfigKey is the key used to store the YAML configuration in the ConfigMap data.
	DikiConfigKey = "config.yaml"
	// DikiConfigMountPath is the mount path for the configurations needed by the diki-run Job pod.
	DikiConfigMountPath = "/config"
	// DikiScanContainerName is the name of the container that is used for executing the scan withn the diki-run Job pod.
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
