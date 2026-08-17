// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package outputs_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	dikireport "github.com/gardener/diki/pkg/report"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/diki-operator/internal/component/reportexporter/outputs"
	reportexporterv1alpha1 "github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

var _ = Describe("WebhookExporter", func() {
	var (
		ctx        = context.Background()
		dikiReport *dikireport.Report
	)

	BeforeEach(func() {
		dikiReport = &dikireport.Report{
			Providers: []dikireport.Provider{
				{
					ID:   "FAKE",
					Name: "FAKE",
					Rulesets: []dikireport.Ruleset{
						{
							ID:   "FAKE",
							Name: "FAKE",
						},
					},
				},
			},
		}
	})

	It("should POST the report to the configured URL", func() {
		var receivedBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Method).To(Equal(http.MethodPost))
			Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))
			var err error
			receivedBody, err = io.ReadAll(r.Body)
			Expect(err).ToNot(HaveOccurred())
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
		})

		details, err := exporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(details).ToNot(BeNil())

		webhookDetails, ok := details.(*outputs.WebhookDetails)
		Expect(ok).To(BeTrue(), "details should be of type *WebhookDetails")
		Expect(webhookDetails.URL).To(Equal(server.URL))
		Expect(webhookDetails.StatusCode).To(Equal(http.StatusOK))

		var receivedReport dikireport.Report
		Expect(json.Unmarshal(receivedBody, &receivedReport)).To(Succeed())
		Expect(receivedReport).To(Equal(*dikiReport))
	})

	It("should apply headers to the request", func() {
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
			Headers: map[string]string{
				"Authorization":   "Bearer my-token",
				"X-Custom-Header": "custom-value",
			},
		})

		_, err := exporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(receivedHeaders.Get("Authorization")).To(Equal("Bearer my-token"))
		Expect(receivedHeaders.Get("X-Custom-Header")).To(Equal("custom-value"))
	})

	It("should return an error when the server responds with a non-2xx status", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal server error"))
		}))
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
		})

		_, err := exporter.Export(ctx, *dikiReport)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("webhook request failed with status 500"))
		Expect(err.Error()).To(ContainSubstring("internal server error"))
	})

	It("should use insecureSkipVerify when configured", func() {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
			TLS: &reportexporterv1alpha1.WebhookTLSConfig{
				InsecureSkipVerify: true,
			},
		})

		details, err := exporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())

		webhookDetails, ok := details.(*outputs.WebhookDetails)
		Expect(ok).To(BeTrue())
		Expect(webhookDetails.StatusCode).To(Equal(http.StatusOK))
	})

	It("should fail TLS verification without insecureSkipVerify for self-signed certs", func() {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
		})

		_, err := exporter.Export(ctx, *dikiReport)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to send webhook request"))
	})

	It("should use a custom CA certificate", func() {
		// Generate a self-signed CA certificate
		caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Expect(err).ToNot(HaveOccurred())

		caTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName: "Test CA",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(time.Hour),
			IsCA:                  true,
			BasicConstraintsValid: true,
			KeyUsage:              x509.KeyUsageCertSign,
		}

		caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
		Expect(err).ToNot(HaveOccurred())

		caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

		// Generate a server certificate signed by the CA
		serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Expect(err).ToNot(HaveOccurred())

		serverTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(2),
			Subject: pkix.Name{
				CommonName: "localhost",
			},
			DNSNames:    []string{"localhost"},
			IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
			NotBefore:   time.Now(),
			NotAfter:    time.Now().Add(time.Hour),
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageServerAuth,
			},
		}

		caCert, err := x509.ParseCertificate(caCertDER)
		Expect(err).ToNot(HaveOccurred())

		serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
		Expect(err).ToNot(HaveOccurred())

		serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
		serverKeyDER, err := x509.MarshalECPrivateKey(serverKey)
		Expect(err).ToNot(HaveOccurred())
		serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyDER})

		serverTLSCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
		Expect(err).ToNot(HaveOccurred())

		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		server.TLS = &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
		}
		server.StartTLS()
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
			TLS: &reportexporterv1alpha1.WebhookTLSConfig{
				CACert: string(caCertPEM),
			},
		})

		details, err := exporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())

		webhookDetails, ok := details.(*outputs.WebhookDetails)
		Expect(ok).To(BeTrue())
		Expect(webhookDetails.StatusCode).To(Equal(http.StatusOK))
	})

	It("should return an error when the CA certificate is invalid", func() {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		exporter := outputs.NewWebhookExporter(reportexporterv1alpha1.WebhookOutputConfig{
			URL: server.URL,
			TLS: &reportexporterv1alpha1.WebhookTLSConfig{
				CACert: "not a valid certificate",
			},
		})

		_, err := exporter.Export(ctx, *dikiReport)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to parse CA certificate"))
	})
})
