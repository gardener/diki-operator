// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package outputs

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	dikireport "github.com/gardener/diki/pkg/report"

	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

// WebhookExporter is responsible for exporting the Diki report via an HTTP webhook.
type WebhookExporter struct {
	Config reportexporterv1alpha1.WebhookOutputConfig
}

var _ Output = &WebhookExporter{}

// WebhookDetails contains the details of the webhook export.
type WebhookDetails struct {
	URL        string `json:"url"`
	StatusCode int    `json:"statusCode"`
}

// NewWebhookExporter creates a new instance of WebhookExporter.
func NewWebhookExporter(config reportexporterv1alpha1.WebhookOutputConfig) *WebhookExporter {
	return &WebhookExporter{
		Config: config,
	}
}

// Type returns the type of the exporter.
func (w *WebhookExporter) Type() reportexporterv1alpha1.OutputType {
	return reportexporterv1alpha1.ExporterTypeWebhook
}

// Export exports the Diki report via an HTTP webhook.
func (w *WebhookExporter) Export(ctx context.Context, report dikireport.Report) (any, error) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	httpClient, err := w.buildHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.Config.URL, bytes.NewReader(reportJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range w.Config.Headers {
		req.Header.Set(key, value)
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

	return &WebhookDetails{
		URL:        w.Config.URL,
		StatusCode: resp.StatusCode,
	}, nil
}

func (w *WebhookExporter) buildHTTPClient() (*http.Client, error) {
	transport := &http.Transport{}

	if w.Config.TLS != nil {
		tlsConfig := &tls.Config{} //nolint:gosec

		if w.Config.TLS.InsecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}

		if len(w.Config.TLS.CACert) != 0 {
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM([]byte(w.Config.TLS.CACert)) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}

			tlsConfig.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}
