// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"

	dikireport "github.com/gardener/diki/pkg/report"

	"github.com/gardener/diki-operator/internal/component/reportexporter/outputs"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

var _ outputs.Output = &Exporter{}

// Exporter is responsible for exporting the Diki report via an HTTP webhook.
type Exporter struct {
	Config         reportexporterv1alpha1.WebhookOutputConfig
	ComplianceScan *dikiv1alpha1.ComplianceScan
}

// ExportDetails contains the details of the webhook export.
type ExportDetails struct {
	URL          string `json:"url"`
	StatusCode   int    `json:"statusCode"`
	ResponseBody string `json:"responseBody,omitempty"`
}

// NewExporter creates a new instance of Exporter.
func NewExporter(config reportexporterv1alpha1.WebhookOutputConfig, complianceScan *dikiv1alpha1.ComplianceScan) *Exporter {
	return &Exporter{
		Config:         config,
		ComplianceScan: complianceScan,
	}
}

// Type returns the type of the exporter.
func (w *Exporter) Type() reportexporterv1alpha1.OutputType {
	return reportexporterv1alpha1.ExporterTypeWebhook
}

// Export exports the Diki report via an HTTP webhook.
func (w *Exporter) Export(ctx context.Context, report dikireport.Report) (any, error) {
	url, err := w.renderURL()
	if err != nil {
		return nil, fmt.Errorf("failed to render webhook URL: %w", err)
	}

	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	httpClient, err := w.buildHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP client: %w", err)
	}

	method := w.Config.Method
	if len(method) == 0 {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reportJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if len(w.Config.BearerTokenFile) != 0 {
		token, err := os.ReadFile(w.Config.BearerTokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read bearer token file: %w", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(bytes.TrimSpace(token))))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("webhook request failed with status %d: %s", resp.StatusCode, string(body))
	}

	details := &ExportDetails{
		URL:        url,
		StatusCode: resp.StatusCode,
	}

	if body, err := io.ReadAll(io.LimitReader(resp.Body, 1024)); err == nil && len(body) > 0 {
		details.ResponseBody = string(body)
	}

	return details, nil
}

// renderURL renders the configured URL as a Go text/template against the compliance scan
// metadata. A URL without any template actions is returned unchanged. Referencing an
// unsupported field (e.g. {{ .ComplianceScanMeta.Name }}) results in an error.
func (w *Exporter) renderURL() (string, error) {
	tmpl, err := template.New("webhookURL").Option("missingkey=error").Parse(w.Config.URL)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, reportexporterv1alpha1.WebhookURLTemplateContext{
		ComplianceScanMeta: reportexporterv1alpha1.ComplianceScanMeta{
			UID: string(w.ComplianceScan.UID),
		},
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (w *Exporter) buildHTTPClient() (*http.Client, error) {
	transport := &http.Transport{}

	if w.Config.TLS != nil {
		tlsConfig := &tls.Config{}

		if len(w.Config.TLS.CACertFile) != 0 {
			caCert, err := os.ReadFile(w.Config.TLS.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}

			tlsConfig.RootCAs = caCertPool
		}

		if len(w.Config.TLS.ClientCertFile) != 0 && len(w.Config.TLS.ClientKeyFile) != 0 {
			cert, err := tls.LoadX509KeyPair(w.Config.TLS.ClientCertFile, w.Config.TLS.ClientKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to parse client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}
