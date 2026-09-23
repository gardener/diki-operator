// SPDX-FileCopyrightText: Contributors to the Gardener project
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
	// defaultCAKey is the default key used in a CA ConfigMap when Key is not specified.
	defaultCAKey = "ca.crt"
	// defaultClientCertKey is the default key used in a client TLS Secret for the certificate.
	defaultClientCertKey = "tls.crt"
	// defaultClientKeyKey is the default key used in a client TLS Secret for the private key.
	defaultClientKeyKey = "tls.key"
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

	if !complianceScan.Spec.DisableDefaultOutputs {
		for i := range r.Config.DefaultOutputs {
			exporterConfig.Outputs = append(exporterConfig.Outputs, *r.Config.DefaultOutputs[i].DeepCopy())
		}
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
		URL:    webhook.URL,
		Method: webhook.Method,
	}

	// Resolve headers from CredentialsRef.
	if webhook.CredentialsRef != nil {
		switch webhook.CredentialsRef.Kind {
		case "Secret", "":
			headers, err := r.resolveHeadersFromSecret(ctx, webhook.CredentialsRef)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve credentials: %w", err)
			}
			config.Headers = headers
		default:
			return nil, fmt.Errorf("unsupported credentialsRef kind %q, only Secret is supported for webhook output", webhook.CredentialsRef.Kind)
		}
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

func (r *Reconciler) resolveTLSConfig(ctx context.Context, tls *v1alpha1.TLSConfig) (*reportexporterv1alpha1.TLSConfig, error) {
	tlsConfig := &reportexporterv1alpha1.TLSConfig{}

	if tls.CAConfigMapRef != nil {
		configMap, err := r.getConfigMap(ctx, &tls.CAConfigMapRef.ResourceReference)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CA certificate: %w", err)
		}
		caCert, err := readConfigMapKey(configMap, ptr.Deref(tls.CAConfigMapRef.Key, defaultCAKey))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CA certificate: %w", err)
		}
		tlsConfig.CACert = caCert
	}

	if tls.MTLSSecretRef != nil {
		secret, err := r.getSecret(ctx, &tls.MTLSSecretRef.ResourceReference)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve mTLS secret: %w", err)
		}

		clientCert, err := readSecretKey(secret, ptr.Deref(tls.MTLSSecretRef.CertKey, defaultClientCertKey))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve client certificate: %w", err)
		}
		tlsConfig.ClientCert = string(clientCert)

		clientKey, err := readSecretKey(secret, ptr.Deref(tls.MTLSSecretRef.PrivateKey, defaultClientKeyKey))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve client key: %w", err)
		}
		tlsConfig.ClientKey = string(clientKey)
	}

	return tlsConfig, nil
}

func (r *Reconciler) resolveHeadersFromSecret(ctx context.Context, ref *v1alpha1.CredentialsRef) (map[string]string, error) {
	secret, err := r.getSecret(ctx, &ref.ResourceReference)
	if err != nil {
		return nil, err
	}

	data, err := readSecretKey(secret, ptr.Deref(ref.HeadersKey, defaultHeadersKey))
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	if err := json.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("failed to parse credentials as JSON map: %w", err)
	}

	return headers, nil
}

func (r *Reconciler) getSecret(ctx context.Context, ref *v1alpha1.ResourceReference) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: ref.Name, Namespace: ref.Namespace}, secret); err != nil {
		return nil, fmt.Errorf("failed to get Secret %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	return secret, nil
}

func (r *Reconciler) getConfigMap(ctx context.Context, ref *v1alpha1.ResourceReference) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: ref.Name, Namespace: ref.Namespace}, configMap); err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	return configMap, nil
}

func readSecretKey(secret *corev1.Secret, key string) ([]byte, error) {
	data, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in Secret %s/%s", key, secret.Namespace, secret.Name)
	}

	return data, nil
}

func readConfigMapKey(configMap *corev1.ConfigMap, key string) (string, error) {
	data, ok := configMap.Data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in ConfigMap %s/%s", key, configMap.Namespace, configMap.Name)
	}

	return data, nil
}
