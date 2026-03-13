// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package outputs_test

import (
	"context"
	"encoding/json"

	dikireport "github.com/gardener/diki/pkg/report"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gardener/diki-operator/internal/component/dikiexporter/outputs"
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
			Config: dikiv1alpha1.ConfigMapOutput{
				Namespace:  ptr.To("default"),
				NamePrefix: ptr.To("diki-report-"),
				Labels: map[string]string{
					"app": "diki-exporter",
				},
			},
		}
	})

	It("should create a ConfigMap with the Diki report", func() {
		details, err := cmExporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(details).ToNot(BeNil())

		cmDetails, ok := details.(*outputs.ConfigMapDetails)
		Expect(ok).To(BeTrue(), "details should be of type *ConfigMapDetails")
		Expect(cmDetails.Name).To(HavePrefix("diki-report-"))
		Expect(cmDetails.Namespace).To(Equal("default"))

		configMap := &corev1.ConfigMap{}
		err = fakeClient.Get(ctx, client.ObjectKey{
			Name:      cmDetails.Name,
			Namespace: cmDetails.Namespace,
		}, configMap)
		Expect(err).ToNot(HaveOccurred())

		Expect(configMap.Labels).To(HaveKeyWithValue("app", "diki-exporter"))

		reportData := configMap.Data["report.json"]
		Expect(reportData).ToNot(BeEmpty())

		var unmarshaledReport dikireport.Report
		err = json.Unmarshal([]byte(reportData), &unmarshaledReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(unmarshaledReport).To(Equal(*dikiReport))
	})

	It("should create a configMap with correct configurations", func() {
		cmExporter.Config.Namespace = ptr.To("custom-namespace")
		cmExporter.Config.NamePrefix = ptr.To("custom-prefix-")
		cmExporter.Config.Labels = map[string]string{
			"custom": "label",
		}

		details, err := cmExporter.Export(ctx, *dikiReport)
		Expect(err).ToNot(HaveOccurred())
		Expect(details).ToNot(BeNil())

		cmDetails, ok := details.(*outputs.ConfigMapDetails)
		Expect(ok).To(BeTrue(), "details should be of type *ConfigMapDetails")
		Expect(cmDetails.Name).To(HavePrefix("custom-prefix-"))
		Expect(cmDetails.Namespace).To(Equal("custom-namespace"))

		configMap := &corev1.ConfigMap{}
		err = fakeClient.Get(ctx, client.ObjectKey{
			Name:      cmDetails.Name,
			Namespace: cmDetails.Namespace,
		}, configMap)
		Expect(err).ToNot(HaveOccurred())

		Expect(configMap.Labels).To(HaveKeyWithValue("custom", "label"))
	})
})
