// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package constants

const (
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
)
