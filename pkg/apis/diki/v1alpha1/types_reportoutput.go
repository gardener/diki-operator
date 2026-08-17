// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster,path=reportoutputs,shortName=ro,singular=reportoutput

// ReportOutput describes a report output.
type ReportOutput struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec contains the specification of this report output.
	Spec ReportOutputSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReportOutputList describes a list of report outputs.
type ReportOutputList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items contains the list of ReportOutputs.
	Items []ReportOutput `json:"items"`
}

// ReportOutputSpec is the specification of a ReportOutput.
type ReportOutputSpec struct {
	// Output describes a specific output of a compliance scan.
	Output Output `json:"output"`
}

// Output describes a specific output of a compliance scan.
type Output struct {
	// ConfigMap contains the configuration for exporting the report to a ConfigMap.
	// +optional
	ConfigMap *OutputConfigMap `json:"configMap,omitempty"`
	// Webhook contains the configuration for exporting the report via an HTTP webhook.
	// +optional
	Webhook *OutputWebhook `json:"webhook,omitempty"`
}

// OutputConfigMap contains the configuration for exporting the report to a ConfigMap.
type OutputConfigMap struct {
	// Namespace is the namespace where the ConfigMap will be created.
	// Defaults to `kube-system`.
	// +kubebuilder:default="kube-system"
	Namespace string `json:"namespace,omitempty"`
	// NamePrefix is the prefix for the generated ConfigMap name.
	// Defaults to "compliance-scan-report-".
	// +kubebuilder:default="compliance-scan-report-"
	NamePrefix string `json:"namePrefix,omitempty"`
}

// OutputWebhook contains the configuration for exporting the report via an HTTP webhook.
type OutputWebhook struct {
	// URL is the destination endpoint to which the report will be POSTed.
	URL string `json:"url"`
	// CredentialsRef is a reference to a Secret whose data at the given key contains a JSON object
	// where keys are HTTP header names and values are the corresponding header values
	// to include in the webhook request.
	// +optional
	CredentialsRef *SecretReference `json:"credentialsRef,omitempty"`
	// TLS configures TLS settings for the webhook connection.
	// Only relevant when URL uses the HTTPS scheme.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
}

// SecretReference is a reference to a specific key within a Secret.
type SecretReference struct {
	// Name is the name of the Secret.
	Name string `json:"name"`
	// Namespace is the namespace of the Secret.
	Namespace string `json:"namespace"`
	// Key is the key within the Secret's data.
	// +optional
	Key *string `json:"key,omitempty"`
}

// TLSConfig configures TLS settings for output types that make outbound HTTPS connections.
type TLSConfig struct {
	// InsecureSkipVerify disables TLS certificate verification.
	// Use with caution; intended for development/testing environments.
	InsecureSkipVerify bool `json:"insecureSkipVerify"`
	// CASecretRef is a reference to a Secret containing a custom CA certificate bundle.
	// The referenced key should contain PEM-encoded CA certificate(s).
	// If not set, the system's root CA pool is used.
	// +optional
	CASecretRef *SecretReference `json:"caSecretRef,omitempty"`
}
