// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package outputs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/andybalholm/brotli"
	dikireport "github.com/gardener/diki/pkg/report"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

// ConfigMapExporter is responsible for exporting the Diki report to a ConfigMap.
type ConfigMapExporter struct {
	Client client.Client
	Config dikiv1alpha1.OutputConfigMap
}

var _ Output = &ConfigMapExporter{}

// ConfigMapDetails contains the details of the created ConfigMap.
type ConfigMapDetails struct {
	ConfigMapRef ConfigMapRef `json:"configMapRef"`
}

// ConfigMapRef contains the reference to a ConfigMap.
type ConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// NewConfigMapExporter creates a new instance of ConfigMapExporter.
func NewConfigMapExporter(client client.Client, config dikiv1alpha1.OutputConfigMap) *ConfigMapExporter {
	return &ConfigMapExporter{
		Client: client,
		Config: config,
	}
}

const reportKey = "report.json.br"

// Type returns the type of the exporter.
func (c *ConfigMapExporter) Type() v1alpha1.OutputType {
	return v1alpha1.ExporterTypeConfigMap
}

// Export exports the Diki report to a ConfigMap.
func (c *ConfigMapExporter) Export(ctx context.Context, report dikireport.Report) (any, error) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	var buf bytes.Buffer
	brWriter := brotli.NewWriter(&buf)
	if _, err := brWriter.Write(reportJSON); err != nil {
		return nil, fmt.Errorf("failed to compress report with brotli: %w", err)
	}
	if err := brWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close brotli writer: %w", err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: c.Config.NamePrefix,
			Namespace:    c.Config.Namespace,
		},
		BinaryData: map[string][]byte{
			reportKey: buf.Bytes(),
		},
	}

	if err := c.Client.Create(ctx, configMap); err != nil {
		return nil, fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	return &ConfigMapDetails{
		ConfigMapRef: ConfigMapRef{
			Name:      configMap.Name,
			Namespace: configMap.Namespace,
		},
	}, nil
}
