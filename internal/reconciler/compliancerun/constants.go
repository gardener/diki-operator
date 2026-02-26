// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

const (
	// ComplianceRunLabel is the label used to identify resources connected to a ComplianceRun.
	ComplianceRunLabel = "diki.gardener.cloud/compliancerun"

	// LabelAppName is the standard Kubernetes label key for application name.
	LabelAppName = "app.kubernetes.io/name"
	// LabelAppManagedBy is the standard Kubernetes label key for the managing tool or operator.
	LabelAppManagedBy = "app.kubernetes.io/managed-by"

	// LabelValueDiki is the application name value used for diki-related resources.
	LabelValueDiki = "diki"
	// LabelValueDikiOperator is the managing operator value used for diki-operator managed resources.
	LabelValueDikiOperator = "diki-operator"

	// ConfigMapGenerateNamePrefix is the prefix for diki config ConfigMap names.
	ConfigMapGenerateNamePrefix = "diki-config-"
	// DikiConfigKey is the key used to store the YAML configuration in the ConfigMap data.
	DikiConfigKey = "config.yaml"

	// RuleOptionsSuffix is the suffix appended to ruleset IDs when looking up rule options in ConfigMaps.
	RuleOptionsSuffix = "-rules"
)
