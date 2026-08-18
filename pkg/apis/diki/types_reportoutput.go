// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package diki

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReportOutput describes a report output.
type ReportOutput struct {
	metav1.TypeMeta
	// Standard object metadata.
	metav1.ObjectMeta

	// Spec contains the specification of this report output.
	Spec ReportOutputSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReportOutputList describes a list of report outputs.
type ReportOutputList struct {
	metav1.TypeMeta
	metav1.ListMeta

	// Items contains the list of ReportOutputs.
	Items []ReportOutput
}

// ReportOutputSpec is the specification of a ReportOutput.
type ReportOutputSpec struct {
	// Output describes a specific output of a compliance scan.
	Output Output
}

// Output describes a specific output of a compliance scan.
type Output struct {
	// ConfigMap contains the configuration for exporting the report to a ConfigMap.
	ConfigMap *OutputConfigMap
	// Webhook contains the configuration for exporting the report via an HTTP webhook.
	Webhook *OutputWebhook
}

// OutputConfigMap contains the configuration for exporting the report to a ConfigMap.
type OutputConfigMap struct {
	// Namespace is the namespace where the ConfigMap will be created.
	// Defaults to `kube-system`.
	Namespace string
	// NamePrefix is the prefix for the generated ConfigMap name.
	// Defaults to "compliance-scan-report-".
	NamePrefix string
}

// OutputWebhook contains the configuration for exporting the report via an HTTP webhook.
type OutputWebhook struct {
	// URL is the destination endpoint to which the report will be POSTed.
	URL string
	// CredentialsRef is a reference to a Secret whose data at the given key contains a JSON object
	// where keys are HTTP header names and values are the corresponding header values
	// to include in the webhook request.
	CredentialsRef *SecretReference
	// TLS configures TLS settings for the webhook connection.
	// Only relevant when URL uses the HTTPS scheme.
	TLS *TLSConfig
}

// SecretReference is a reference to a specific key within a Secret.
type SecretReference struct {
	// Name is the name of the Secret.
	Name string
	// Namespace is the namespace of the Secret.
	Namespace string
	// Key is the key within the Secret's data.
	Key *string
}

// TLSConfig configures TLS settings for output types that make outbound HTTPS connections.
type TLSConfig struct {
	// InsecureSkipVerify disables TLS certificate verification.
	// Use with caution; intended for development/testing environments.
	InsecureSkipVerify bool
	// CASecretRef is a reference to a Secret containing a custom CA certificate bundle.
	// The referenced key should contain PEM-encoded CA certificate(s).
	// If not set, the system's root CA pool is used.
	CASecretRef *SecretReference
}
