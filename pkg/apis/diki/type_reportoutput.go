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
	// Outputs describes a specific output of a compliance scan
	Output Output
}

// Output describes a specific output of a compliance scan
type Output struct {
	// ConfigMap contains the configuration for exporting the report to a ConfigMap.
	ConfigMap *ConfigMapOutput
}

// ConfigMapOutput contains the configuration for exporting the report to a ConfigMap.
type ConfigMapOutput struct {
	// Namespace is the namespace where the ConfigMap will be created.
	// Defaults to `kube-system`.
	Namespace *string
	// NamePrefix is the prefix for the generated ConfigMap name.
	// Defaults to "diki-report-".
	NamePrefix *string
	// Labels are additional labels to add to the ConfigMap.
	Labels map[string]string
}
