// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DikiExporterConfiguration defines the configuration for the diki-exporter.
type DikiExporterConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// ReportPath is the path to the Diki report file to be exported.
	ReportPath string `json:"reportPath"`
	// ComplianceScanName is the name of the compliance scan, which generated the report.
	ComplianceScanName string `json:"complianceScanName"`
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
)
