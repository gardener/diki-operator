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
	// ServiceAccountNameDikiRun is the name for the diki-run Job related ServiceAccount.
	ServiceAccountNameDikiRun = "diki-run"
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
	// ReportExporterContainerName is the name of the report-exporter container in the diki-run Job pod.
	ReportExporterContainerName = "report-exporter"

	// ReportVolumeName is the name of the shared emptyDir volume used to pass the report between containers.
	ReportVolumeName = "diki-report"
	// ReportMountPath is the mount path for the shared report volume.
	ReportMountPath = "/report"
	// ReportFileName is the name of the report file written by the diki-scan container.
	ReportFileName = "report.json"

	// ExporterConfigKey is the key used to store the exporter configuration in the ConfigMap data.
	ExporterConfigKey = "exporter-config.yaml"

	// RuleOptionsSuffix is the suffix appended to ruleset IDs when looking up rule options in ConfigMaps.
	RuleOptionsSuffix = "-rules"

	// ConditionReasonRunning is the reason for ComplianceScan condition when it is running.
	ConditionReasonRunning = "ComplianceScanRunning"
	// ConditionReasonCompleted is the reason for ComplianceScan condition when it has completed successfully.
	ConditionReasonCompleted = "ComplianceScanCompleted"
	// ConditionReasonFailed is the reason for ComplianceScan condition when it has failed.
	ConditionReasonFailed = "ComplianceScanFailed"
)
