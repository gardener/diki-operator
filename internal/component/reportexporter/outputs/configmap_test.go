// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package outputs_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	dikireport "github.com/gardener/diki/pkg/report"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gardener/diki-operator/internal/component/reportexporter/outputs"
	dikiinstall "github.com/gardener/diki-operator/pkg/apis/diki/install"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("Controller", func() {
	var (
		ctx = logf.IntoContext(context.Background(), logzap.New(logzap.WriteTo(GinkgoWriter)))

		fakeClient client.Client
		dikiReport *dikireport.Report
		cmExporter outputs.ConfigMapExporter
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(dikiinstall.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).Build()

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

		cmExporter = outputs.ConfigMapExporter{
			Client: fakeClient,
			Config: dikiv1alpha1.OutputConfigMap{
				Namespace:  "default",
				NamePrefix: "diki-report-",
			},
		}
	})

	It("should create a ConfigMap with the Diki report", func() {
		details, err := cmExporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(details).ToNot(BeNil())

		cmDetails, ok := details.(*outputs.ConfigMapDetails)
		Expect(ok).To(BeTrue(), "details should be of type *ConfigMapDetails")
		Expect(cmDetails.ConfigMapRef.Name).To(HavePrefix("diki-report-"))
		Expect(cmDetails.ConfigMapRef.Namespace).To(Equal("default"))

		configMap := &corev1.ConfigMap{}
		err = fakeClient.Get(ctx, client.ObjectKey{
			Name:      cmDetails.ConfigMapRef.Name,
			Namespace: cmDetails.ConfigMapRef.Namespace,
		}, configMap)
		Expect(err).ToNot(HaveOccurred())

		reportData := configMap.BinaryData["report.json.gz"]
		Expect(reportData).ToNot(BeEmpty())

		gzReader, err := gzip.NewReader(bytes.NewReader(reportData))
		Expect(err).ToNot(HaveOccurred())
		decompressed, err := io.ReadAll(gzReader)
		Expect(err).ToNot(HaveOccurred())

		var unmarshaledReport dikireport.Report
		err = json.Unmarshal(decompressed, &unmarshaledReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(unmarshaledReport).To(Equal(*dikiReport))
	})

	It("should create a configMap with correct configurations", func() {
		cmExporter.Config.Namespace = "custom-namespace"
		cmExporter.Config.NamePrefix = "custom-prefix-"

		details, err := cmExporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(details).ToNot(BeNil())

		cmDetails, ok := details.(*outputs.ConfigMapDetails)
		Expect(ok).To(BeTrue(), "details should be of type *ConfigMapDetails")
		Expect(cmDetails.ConfigMapRef.Name).To(HavePrefix("custom-prefix-"))
		Expect(cmDetails.ConfigMapRef.Namespace).To(Equal("custom-namespace"))

		configMap := &corev1.ConfigMap{}
		err = fakeClient.Get(ctx, client.ObjectKey{
			Name:      cmDetails.ConfigMapRef.Name,
			Namespace: cmDetails.ConfigMapRef.Namespace,
		}, configMap)
		Expect(err).ToNot(HaveOccurred())
	})
})
