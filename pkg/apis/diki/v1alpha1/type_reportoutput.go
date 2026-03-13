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
	// Outputs describes a specific output of a compliance scan
	Output Output `json:"output"`
}

// Output describes a specific output of a compliance scan
type Output struct {
	// ConfigMap contains the configuration for exporting the report to a ConfigMap.
	// +optional
	ConfigMap *ConfigMapOutput `json:"configMap,omitempty"`
}

// ConfigMapOutput contains the configuration for exporting the report to a ConfigMap.
type ConfigMapOutput struct {
	// Namespace is the namespace where the ConfigMap will be created.
	// Defaults to `kube-system`.
	// +kubebuilder:default="kube-system"
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// NamePrefix is the prefix for the generated ConfigMap name.
	// Defaults to "diki-report-".
	// +kubebuilder:default="diki-report-"
	// +optional
	NamePrefix *string `json:"namePrefix,omitempty"`
	// Labels are additional labels to add to the ConfigMap.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}
