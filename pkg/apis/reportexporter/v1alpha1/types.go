// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReportExporterConfiguration defines the configuration for the report-exporter.
type ReportExporterConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// ReportPath is the path to the Diki report file to be exported.
	ReportPath string `json:"reportPath"`
	// ComplianceScanName is the name of the compliance scan, which generated the report.
	ComplianceScanName string `json:"complianceScanName"`
	// WaitForReport specifies whether the exporter should wait for the report file to appear before reading it.
	// +optional
	WaitForReport bool `json:"waitForReport,omitempty"`
	// ReportWaitTimeout is the maximum duration to wait for the report file to appear.
	// Only used when WaitForReport is true. If not set, the exporter waits indefinitely.
	// +optional
	ReportWaitTimeout *metav1.Duration `json:"reportWaitTimeout,omitempty"`
	// Outputs contains the list of output configurations.
	Outputs []Output `json:"outputs"`
}

// Output describes a specific output.
type Output struct {
	// Type is the type of the output.
	Type OutputType `json:"type"`
	// Name is the name of the output, used for identification purposes.
	Name string `json:"name"`
	// Config contains the configuration for the output.
	// +optional
	Config runtime.RawExtension `json:"config,omitempty"`
}

// OutputType is an alias for string representing the type of an exporter.
type OutputType string

const (
	// ExporterTypeConfigMap is the type for exporting reports to a ConfigMap.
	ExporterTypeConfigMap OutputType = "ConfigMap"
	// ExporterTypeWebhook is the type for exporting reports via an HTTP webhook.
	ExporterTypeWebhook OutputType = "Webhook"
)

// WebhookOutputConfig is the resolved configuration for a webhook output.
// All secrets are resolved by the operator at reconciliation time.
// The exporter receives plain values and does not need access to Secrets.
type WebhookOutputConfig struct {
	// URL is the destination endpoint to which the report will be sent.
	URL string `json:"url"`
	// Method is the HTTP method used to send the report.
	// The report payload is always sent as the full JSON body regardless of the method.
	// This is useful when the receiving endpoint expects a specific method (e.g. PUT for upsert semantics).
	Method string `json:"method"`
	// Headers contains HTTP headers to include in the webhook request.
	// +optional
	Headers map[string]string `json:"headers,omitempty"`
	// TLS contains resolved TLS settings for the webhook connection.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
}

// TLSConfig contains resolved TLS settings for the webhook exporter.
type TLSConfig struct {
	// InsecureSkipVerify disables TLS certificate verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// CACert contains a PEM-encoded CA certificate bundle.
	CACert string `json:"caCert,omitempty"`
}
