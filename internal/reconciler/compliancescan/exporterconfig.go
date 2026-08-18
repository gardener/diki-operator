// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

const (
	// defaultHeadersKey is the default key used in a credentials Secret when Key is not specified.
	defaultHeadersKey = "headers"
	// defaultCAKey is the default key used in a CA Secret when Key is not specified.
	defaultCAKey = "ca.crt"
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

		output, err := r.convertReportOutput(ctx, reportOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ReportOutput %q: %w", outputRef.Name, err)
		}

		exporterConfig.Outputs = append(exporterConfig.Outputs, *output)
	}

	return exporterConfig, nil
}

func (r *Reconciler) convertReportOutput(ctx context.Context, reportOutput *v1alpha1.ReportOutput) (*reportexporterv1alpha1.Output, error) {
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

	if reportOutput.Spec.Output.Webhook != nil {
		webhookConfig, err := r.resolveWebhookConfig(ctx, reportOutput.Spec.Output.Webhook)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve webhook config: %w", err)
		}

		configBytes, err := json.Marshal(webhookConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Webhook output config: %w", err)
		}

		return &reportexporterv1alpha1.Output{
			Type: reportexporterv1alpha1.ExporterTypeWebhook,
			Name: reportOutput.Name,
			Config: runtime.RawExtension{
				Raw: configBytes,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported output type in ReportOutput %q", reportOutput.Name)
}

func (r *Reconciler) resolveWebhookConfig(ctx context.Context, webhook *v1alpha1.OutputWebhook) (*reportexporterv1alpha1.WebhookOutputConfig, error) {
	config := &reportexporterv1alpha1.WebhookOutputConfig{
		URL: webhook.URL,
	}

	// Resolve headers from CredentialsRef secret.
	if webhook.CredentialsRef != nil {
		headers, err := r.resolveHeadersFromSecret(ctx, webhook.CredentialsRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve credentials: %w", err)
		}
		config.Headers = headers
	}

	// Resolve TLS config.
	if webhook.TLS != nil {
		tlsConfig, err := r.resolveTLSConfig(ctx, webhook.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve TLS config: %w", err)
		}
		config.TLS = tlsConfig
	}

	return config, nil
}

func (r *Reconciler) resolveTLSConfig(ctx context.Context, tls *v1alpha1.TLSConfig) (*reportexporterv1alpha1.WebhookTLSConfig, error) {
	tlsConfig := &reportexporterv1alpha1.WebhookTLSConfig{
		InsecureSkipVerify: tls.InsecureSkipVerify,
	}

	if tls.CASecretRef != nil {
		secret, err := r.getSecret(ctx, tls.CASecretRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CA certificate: %w", err)
		}
		caCert, err := readSecretKey(secret, tls.CASecretRef, defaultCAKey)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CA certificate: %w", err)
		}
		tlsConfig.CACert = string(caCert)
	}

	return tlsConfig, nil
}

func (r *Reconciler) resolveHeadersFromSecret(ctx context.Context, ref *v1alpha1.SecretReference) (map[string]string, error) {
	secret, err := r.getSecret(ctx, ref)
	if err != nil {
		return nil, err
	}

	data, err := readSecretKey(secret, ref, defaultHeadersKey)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	if err := json.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("failed to parse credentials as JSON map: %w", err)
	}

	return headers, nil
}

func (r *Reconciler) getSecret(ctx context.Context, ref *v1alpha1.SecretReference) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: ref.Name, Namespace: ref.Namespace}, secret); err != nil {
		return nil, fmt.Errorf("failed to get Secret %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	return secret, nil
}

func readSecretKey(secret *corev1.Secret, ref *v1alpha1.SecretReference, defaultKey string) ([]byte, error) {
	key := ptr.Deref(ref.Key, defaultKey)

	data, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in Secret %s/%s", key, secret.Namespace, secret.Name)
	}

	return data, nil
}
