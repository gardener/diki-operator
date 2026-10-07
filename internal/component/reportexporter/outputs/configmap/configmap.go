// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package configmap

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"

	dikireport "github.com/gardener/diki/pkg/report"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/internal/component/reportexporter/outputs"
	"github.com/gardener/diki-operator/internal/constants"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

var _ outputs.Output = &Exporter{}

// Exporter is responsible for exporting the Diki report to a ConfigMap.
type Exporter struct {
	Client         client.Client
	Config         dikiv1alpha1.OutputConfigMap
	ComplianceScan *dikiv1alpha1.ComplianceScan
}

// ExportDetails contains the details of the created ConfigMap.
type ExportDetails struct {
	ConfigMapRef Ref `json:"configMapRef"`
}

// Ref contains the reference to a ConfigMap.
type Ref struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// NewExporter creates a new instance of Exporter.
func NewExporter(client client.Client, config dikiv1alpha1.OutputConfigMap, complianceScan *dikiv1alpha1.ComplianceScan) *Exporter {
	return &Exporter{
		Client:         client,
		Config:         config,
		ComplianceScan: complianceScan,
	}
}

const reportKey = "report.json.gz"

// Type returns the type of the exporter.
func (c *Exporter) Type() v1alpha1.OutputType {
	return v1alpha1.ExporterTypeConfigMap
}

// Export exports the Diki report to a ConfigMap.
func (c *Exporter) Export(ctx context.Context, report dikireport.Report) (any, error) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(reportJSON); err != nil {
		// call gzWriter.Close for the sake of completeness
		// ignore the error as this would probably be the same error as the error returned by gzWriter.Write
		_ = gzWriter.Close()
		return nil, fmt.Errorf("failed to compress report with gzip: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: c.Config.NamePrefix,
			Namespace:    c.Config.Namespace,
			Labels:       c.getLabels(),
		},
		BinaryData: map[string][]byte{
			reportKey: buf.Bytes(),
		},
	}

	if err := c.Client.Create(ctx, configMap); err != nil {
		return nil, fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	return &ExportDetails{
		ConfigMapRef: Ref{
			Name:      configMap.Name,
			Namespace: configMap.Namespace,
		},
	}, nil
}

func (c *Exporter) getLabels() map[string]string {
	return map[string]string{
		constants.LabelAppName:            constants.LabelValueDiki,
		constants.LabelAppManagedBy:       constants.LabelValueDikiOperator,
		constants.LabelComplianceScanName: c.ComplianceScan.Name,
		constants.LabelComplianceScanUID:  string(c.ComplianceScan.UID),
	}
}
