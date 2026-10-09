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
	"time"

	dikireport "github.com/gardener/diki/pkg/report"

	"github.com/gardener/diki-operator/internal/component/reportexporter/outputs"
	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

var _ outputs.Output = &Exporter{}

// Exporter is responsible for exporting the Diki report via an HTTP webhook.
type Exporter struct {
	Config reportexporterv1alpha1.WebhookOutputConfig
}

// ExportDetails contains the details of the webhook export.
type ExportDetails struct {
	URL          string `json:"url"`
	StatusCode   int    `json:"statusCode"`
	ResponseBody string `json:"responseBody,omitempty"`
}

// NewExporter creates a new instance of Exporter.
func NewExporter(config reportexporterv1alpha1.WebhookOutputConfig) *Exporter {
	return &Exporter{
		Config: config,
	}
}

// Type returns the type of the exporter.
func (w *Exporter) Type() reportexporterv1alpha1.OutputType {
	return reportexporterv1alpha1.ExporterTypeWebhook
}

// Export exports the Diki report via an HTTP webhook.
func (w *Exporter) Export(ctx context.Context, report dikireport.Report) (any, error) {
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

	req, err := http.NewRequestWithContext(ctx, method, w.Config.URL, bytes.NewReader(reportJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if len(w.Config.HeadersFile) != 0 {
		headersData, err := os.ReadFile(w.Config.HeadersFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read headers file: %w", err)
		}
		headers := make(map[string]string)
		if err := json.Unmarshal(headersData, &headers); err != nil {
			return nil, fmt.Errorf("failed to parse headers file as JSON map: %w", err)
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
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
		URL:        w.Config.URL,
		StatusCode: resp.StatusCode,
	}

	if body, err := io.ReadAll(io.LimitReader(resp.Body, 1024)); err == nil && len(body) > 0 {
		details.ResponseBody = string(body)
	}

	return details, nil
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
