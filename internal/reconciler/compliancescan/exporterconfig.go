// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

func (r *Reconciler) buildExporterConfig(ctx context.Context, complianceScan *v1alpha1.ComplianceScan) (*reportexporterv1alpha1.ReportExporterConfiguration, error) {
	exporterConfig := &reportexporterv1alpha1.ReportExporterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "exporter.diki.gardener.cloud/v1alpha1",
			Kind:       "ReportExporterConfiguration",
		},
		ReportPath:         ReportMountPath + "/" + ReportFileName,
		ComplianceScanName: complianceScan.Name,
		WaitForReport:      true,
	}

	for _, outputRef := range complianceScan.Spec.Outputs {
		reportOutput := &v1alpha1.ReportOutput{}
		if err := r.Client.Get(ctx, client.ObjectKey{Name: outputRef.Name}, reportOutput); err != nil {
			return nil, fmt.Errorf("failed to get ReportOutput %q: %w", outputRef.Name, err)
		}

		output, err := convertReportOutput(reportOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ReportOutput %q: %w", outputRef.Name, err)
		}

		exporterConfig.Outputs = append(exporterConfig.Outputs, *output)
	}

	return exporterConfig, nil
}

func convertReportOutput(reportOutput *v1alpha1.ReportOutput) (*reportexporterv1alpha1.Output, error) {
	if reportOutput.Spec.Output.ConfigMap != nil {
		configBytes, err := json.Marshal(reportOutput.Spec.Output.ConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ConfigMap output config: %w", err)
		}

		return &reportexporterv1alpha1.Output{
			Type: reportexporterv1alpha1.ExporterTypeConfigMap,
			Name: reportOutput.Name,
			Config: runtime.RawExtension{
				Raw: configBytes,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported output type in ReportOutput %q", reportOutput.Name)
}
