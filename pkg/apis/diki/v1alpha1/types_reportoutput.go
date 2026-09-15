// SPDX-FileCopyrightText: Contributors to the Gardener project
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
	// URL is the destination endpoint to which the report will be sent.
	URL string `json:"url"`
	// Method is the HTTP method used to send the report.
	// The report payload is always sent as the full JSON body regardless of the method.
	// This is useful when the receiving endpoint expects a specific method (e.g. PUT for upsert semantics).
	// Defaults to "POST".
	// +optional
	// +kubebuilder:default="POST"
	// +kubebuilder:validation:Enum={"POST","PUT"}
	Method string `json:"method,omitempty"`
	// CredentialsRef is a reference to a Secret whose data at the given key contains a JSON object
	// where keys are HTTP header names and values are the corresponding header values
	// to include in the webhook request.
	// +optional
	CredentialsRef *CredentialsSecretRef `json:"credentialsRef,omitempty"`
	// TLS configures TLS settings for the webhook connection.
	// Only relevant when URL uses the HTTPS scheme.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
}

// CredentialsSecretRef is a reference to a Secret containing HTTP headers for webhook authentication.
type CredentialsSecretRef struct {
	ResourceReference `json:",inline"`

	// HeadersKey is the key within the Secret's data that contains the JSON-encoded headers.
	// Defaults to `headers`.
	// +optional
	HeadersKey *string `json:"headersKey,omitempty"`
}

// ResourceReference is a reference to a namespaced Kubernetes resource.
type ResourceReference struct {
	// Name is the name of the resource.
	Name string `json:"name"`
	// Namespace is the namespace of the resource.
	Namespace string `json:"namespace"`
}

// CAConfigMapRef is a reference to a ConfigMap containing a PEM-encoded CA certificate bundle.
type CAConfigMapRef struct {
	ResourceReference `json:",inline"`

	// Key is the key within the ConfigMap's data that contains the CA certificate(s).
	// Defaults to `ca.crt`.
	// +optional
	Key *string `json:"key,omitempty"`
}

// MTLSSecretRef is a reference to a Kubernetes TLS Secret containing a client certificate
// and key for mutual TLS (mTLS) authentication.
type MTLSSecretRef struct {
	ResourceReference `json:",inline"`

	// CertKey is the key within the Secret's data that contains the PEM-encoded client certificate.
	// Defaults to `tls.crt`.
	// +optional
	CertKey *string `json:"certKey,omitempty"`
	// PrivateKey is the key within the Secret's data that contains the PEM-encoded client private key.
	// Defaults to `tls.key`.
	// +optional
	PrivateKey *string `json:"privateKey,omitempty"`
}

// TLSConfig configures TLS settings for output types that make outbound HTTPS connections.
type TLSConfig struct {
	// InsecureSkipVerify disables TLS certificate verification.
	// Use with caution; intended for development/testing environments.
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// CAConfigMapRef is a reference to a ConfigMap containing a custom CA certificate bundle.
	// If not set, the system's root CA pool is used.
	// +optional
	CAConfigMapRef *CAConfigMapRef `json:"caConfigMapRef,omitempty"`
	// MTLSSecretRef is a reference to a Kubernetes TLS Secret containing a client certificate
	// and key for mutual TLS (mTLS) authentication.
	// +optional
	MTLSSecretRef *MTLSSecretRef `json:"mtlsSecretRef,omitempty"`
}
