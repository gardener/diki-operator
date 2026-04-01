// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reportexporter_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	dikireport "github.com/gardener/diki/pkg/report"
	"github.com/gardener/diki/pkg/rule"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/gardener/diki-operator/internal/component/reportexporter"
	dikiinstall "github.com/gardener/diki-operator/pkg/apis/diki/install"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

var _ = Describe("ReportExporter", func() {
	var (
		ctx = context.TODO()

		fakeClient     client.Client
		exporter       *reportexporter.ReportExporter
		complianceScan *dikiv1alpha1.ComplianceScan
		tempDir        string
		reportPath     string
		dikiReport     *dikireport.Report
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "report-exporter-test-")
		Expect(err).NotTo(HaveOccurred())

		reportPath = filepath.Join(tempDir, "report.json")

		scheme := runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(dikiinstall.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).
			Build()

		complianceScan = &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-compliancescan",
			},
			Spec: dikiv1alpha1.ComplianceScanSpec{
				Rulesets: []dikiv1alpha1.RulesetConfig{
					{
						ID:      "test-ruleset",
						Version: "v1.0.0",
					},
				},
			},
			Status: dikiv1alpha1.ComplianceScanStatus{
				Phase: dikiv1alpha1.ComplianceScanRunning,
			},
		}

		dikiReport = &dikireport.Report{
			Providers: []dikireport.Provider{
				{
					ID:   "test-provider",
					Name: "Test Provider",
					Rulesets: []dikireport.Ruleset{
						{
							ID:      "test-ruleset",
							Name:    "Test Ruleset",
							Version: "v1.0.0",
							Rules: []dikireport.Rule{
								{
									ID:   "rule-1",
									Name: "Test Rule 1",
									Checks: []dikireport.Check{
										{
											Status: rule.Passed,
										},
									},
								},
								{
									ID:   "rule-2",
									Name: "Test Rule 2",
									Checks: []dikireport.Check{
										{
											Status: rule.Passed,
										},
									},
								},
								{
									ID:   "rule-3",
									Name: "Test Rule 3",
									Checks: []dikireport.Check{
										{
											Status: rule.Passed,
										},
									},
								},
								{
									ID:   "rule-4",
									Name: "Test Rule 4",
									Checks: []dikireport.Check{
										{
											Status: rule.Failed,
										},
									},
								},
								{
									ID:   "rule-5",
									Name: "Test Rule 5",
									Checks: []dikireport.Check{
										{
											Status: rule.Failed,
										},
									},
								},
								{
									ID:   "rule-6",
									Name: "Test Rule 6",
									Checks: []dikireport.Check{
										{
											Status: rule.Errored,
										},
									},
								},
								{
									ID:   "rule-7",
									Name: "Test Rule 7",
									Checks: []dikireport.Check{
										{
											Status: rule.Errored,
										},
									},
								},
								{
									ID:   "rule-8",
									Name: "Test Rule 8",
									Checks: []dikireport.Check{
										{
											Status: rule.Errored,
										},
									},
								},
								{
									ID:   "rule-9",
									Name: "Test Rule 9",
									Checks: []dikireport.Check{
										{
											Status: rule.Warning,
										},
									},
								},
								{
									ID:   "rule-10",
									Name: "Test Rule 10",
									Checks: []dikireport.Check{
										{
											Status: rule.Skipped,
										},
									},
								},
								{
									ID:   "rule-11",
									Name: "Test Rule 11",
									Checks: []dikireport.Check{
										{
											Status: rule.Skipped,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		reportData, err := json.Marshal(dikiReport)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(reportPath, reportData, 0600)).To(Succeed())

		exporter = reportexporter.NewReportExporter(
			fakeClient,
			v1alpha1.ReportExporterConfiguration{
				ReportPath:         reportPath,
				ComplianceScanName: complianceScan.Name,
				Outputs: []v1alpha1.Output{
					{
						Type: v1alpha1.ExporterTypeConfigMap,
						Name: "test-output",
						Config: runtime.RawExtension{
							Raw: []byte(`{"namespace":"kube-system","namePrefix":"diki-report-"}`),
						},
					},
				},
			},
		)
	})

	Describe("Export", func() {
		It("should successfully export the report and update ComplianceScan status", func() {
			err := exporter.Export(ctx)
			Expect(err).NotTo(HaveOccurred())

			updatedScan := &dikiv1alpha1.ComplianceScan{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, updatedScan)).To(Succeed())

			Expect(updatedScan.Status.Rulesets).To(HaveLen(1))
			Expect(updatedScan.Status.Rulesets[0]).To(MatchFields(IgnoreExtras, Fields{
				"ID":      Equal("test-ruleset"),
				"Version": Equal("v1.0.0"),
			}))

			summary := updatedScan.Status.Rulesets[0].Results.Summary
			Expect(summary.Passed).To(Equal(int32(3)))
			Expect(summary.Failed).To(Equal(int32(2)))
			Expect(summary.Errored).To(Equal(int32(3)))
			Expect(summary.Warning).To(Equal(int32(1)))
			Expect(summary.Skipped).To(Equal(int32(2)))

			findings := updatedScan.Status.Rulesets[0].Results.Rules
			Expect(findings.Failed).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"ID":   Equal("rule-4"),
					"Name": Equal("Test Rule 4"),
				}),
				MatchFields(IgnoreExtras, Fields{
					"ID":   Equal("rule-5"),
					"Name": Equal("Test Rule 5"),
				}),
			))
			Expect(findings.Errored).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"ID":   Equal("rule-6"),
					"Name": Equal("Test Rule 6"),
				}),
				MatchFields(IgnoreExtras, Fields{
					"ID":   Equal("rule-7"),
					"Name": Equal("Test Rule 7"),
				}),
				MatchFields(IgnoreExtras, Fields{
					"ID":   Equal("rule-8"),
					"Name": Equal("Test Rule 8"),
				}),
			))
			Expect(findings.Warning).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"ID":   Equal("rule-9"),
					"Name": Equal("Test Rule 9"),
				}),
			))

			// Verify output statuses
			Expect(updatedScan.Status.Outputs).To(HaveLen(1))
			Expect(updatedScan.Status.Outputs[0].OutputName).To(Equal("test-output"))
			Expect(updatedScan.Status.Outputs[0].Phase).To(Equal(dikiv1alpha1.OutputStatusCompleted))
			Expect(updatedScan.Status.Outputs[0].Details.Raw).NotTo(BeEmpty())
		})

		It("should handle multiple outputs correctly", func() {
			exporter.Config.Outputs = append(exporter.Config.Outputs, v1alpha1.Output{
				Type: v1alpha1.ExporterTypeConfigMap,
				Name: "test-output-2",
				Config: runtime.RawExtension{
					Raw: []byte(`{"namespace":"default","namePrefix":"report-"}`),
				},
			})

			err := exporter.Export(ctx)
			Expect(err).NotTo(HaveOccurred())

			updatedScan := &dikiv1alpha1.ComplianceScan{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, updatedScan)).To(Succeed())

			Expect(updatedScan.Status.Outputs).To(HaveLen(2))

			outputNames := []string{updatedScan.Status.Outputs[0].OutputName, updatedScan.Status.Outputs[1].OutputName}
			Expect(outputNames).To(ConsistOf("test-output", "test-output-2"))

			for _, output := range updatedScan.Status.Outputs {
				Expect(output.Phase).To(Equal(dikiv1alpha1.OutputStatusCompleted))
			}
		})

		It("should handle output export failures gracefully", func() {
			// Invalid config to cause unmarshal error
			exporter.Config.Outputs = []v1alpha1.Output{
				{
					Type: v1alpha1.ExporterTypeConfigMap,
					Name: "invalid-output",
					Config: runtime.RawExtension{
						Raw: []byte(`{invalid json}`),
					},
				},
			}

			err := exporter.Export(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to unmarshal ConfigMapOutput"))
		})

		It("should not export if ComplianceScan is already completed", func() {
			patch := client.MergeFrom(complianceScan.DeepCopy())
			complianceScan.Status.Phase = dikiv1alpha1.ComplianceScanCompleted
			Expect(fakeClient.Status().Patch(ctx, complianceScan, patch)).To(Succeed())

			err := exporter.Export(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("complianceScan is already completed"))
		})

		It("should return error if ComplianceScan does not exist", func() {
			exporter.Config.ComplianceScanName = "non-existent-scan"

			err := exporter.Export(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error retrieving complianceScan"))
		})

		It("should return error if report file does not exist", func() {
			exporter.Config.ReportPath = filepath.Join(tempDir, "non-existent-report.json")

			err := exporter.Export(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error reading report file"))
		})

		It("should return error if report file contains invalid JSON", func() {
			Expect(os.WriteFile(reportPath, []byte("invalid json"), 0600)).To(Succeed())

			err := exporter.Export(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error unmarshaling report"))
		})

		It("should handle empty ruleset report correctly", func() {
			dikiReport.Providers[0].Rulesets[0].Rules = []dikireport.Rule{}
			reportData, err := json.Marshal(dikiReport)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(reportPath, reportData, 0600)).To(Succeed())

			err = exporter.Export(ctx)
			Expect(err).NotTo(HaveOccurred())

			updatedScan := &dikiv1alpha1.ComplianceScan{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, updatedScan)).To(Succeed())

			Expect(updatedScan.Status.Rulesets).To(HaveLen(1))
			summary := updatedScan.Status.Rulesets[0].Results.Summary
			Expect(summary.Passed).To(Equal(int32(0)))
			Expect(summary.Failed).To(Equal(int32(0)))
			Expect(summary.Errored).To(Equal(int32(0)))
			Expect(summary.Warning).To(Equal(int32(0)))
			Expect(summary.Skipped).To(Equal(int32(0)))
		})

		It("should handle multiple rulesets in report", func() {
			dikiReport.Providers[0].Rulesets = append(dikiReport.Providers[0].Rulesets, dikireport.Ruleset{
				ID:      "test-ruleset-2",
				Name:    "Test Ruleset 2",
				Version: "v2.0.0",
				Rules: []dikireport.Rule{
					{
						ID:   "rule-5",
						Name: "Test Rule 5",
						Checks: []dikireport.Check{
							{
								Status: rule.Passed,
							},
						},
					},
				},
			})

			reportData, err := json.Marshal(dikiReport)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(reportPath, reportData, 0600)).To(Succeed())

			err = exporter.Export(ctx)
			Expect(err).NotTo(HaveOccurred())

			updatedScan := &dikiv1alpha1.ComplianceScan{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, updatedScan)).To(Succeed())

			Expect(updatedScan.Status.Rulesets).To(HaveLen(2))
			Expect(updatedScan.Status.Rulesets[0].ID).To(Equal("test-ruleset"))
			Expect(updatedScan.Status.Rulesets[1].ID).To(Equal("test-ruleset-2"))
		})

		It("should handle rules with multiple checks of different statuses", func() {
			dikiReport.Providers[0].Rulesets[0].Rules = []dikireport.Rule{
				{
					ID:   "rule-mixed",
					Name: "Mixed Status Rule",
					Checks: []dikireport.Check{
						{Status: rule.Passed},
						{Status: rule.Failed},
						{Status: rule.Warning},
					},
				},
			}

			reportData, err := json.Marshal(dikiReport)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(reportPath, reportData, 0600)).To(Succeed())

			err = exporter.Export(ctx)
			Expect(err).NotTo(HaveOccurred())

			updatedScan := &dikiv1alpha1.ComplianceScan{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, updatedScan)).To(Succeed())

			findings := updatedScan.Status.Rulesets[0].Results.Rules
			// Should appear in both failed and warning lists
			Expect(findings.Failed).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"ID": Equal("rule-mixed"),
			})))
			Expect(findings.Warning).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"ID": Equal("rule-mixed"),
			})))
		})

		It("should fail when unsupported output type is passed", func() {
			exporter.Config.Outputs = []v1alpha1.Output{
				{
					Type: "UnsupportedType",
					Name: "unsupported-output",
				},
			}

			err := exporter.Export(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported output type"))
		})

		It("should handle output export failures and mark output as failed", func() {
			scheme := runtime.NewScheme()
			Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
			Expect(dikiinstall.AddToScheme(scheme)).To(Succeed())

			fakeClientWithError := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).
				WithObjects(complianceScan).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						// Fail ConfigMap creation to trigger export error
						if _, ok := obj.(*corev1.ConfigMap); ok {
							return errors.New("simulated ConfigMap creation failure")
						}
						return client.Create(ctx, obj, opts...)
					},
				}).
				Build()

			exporter.Client = fakeClientWithError

			err := exporter.Export(ctx)
			Expect(err).NotTo(HaveOccurred())

			updatedScan := &dikiv1alpha1.ComplianceScan{}
			Expect(fakeClientWithError.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, updatedScan)).To(Succeed())

			// Verify ruleset summaries were still updated
			Expect(updatedScan.Status.Rulesets).To(HaveLen(1))

			// Verify output status shows failure
			Expect(updatedScan.Status.Outputs).To(HaveLen(1))
			Expect(updatedScan.Status.Outputs[0].OutputName).To(Equal("test-output"))
			Expect(updatedScan.Status.Outputs[0].Phase).To(Equal(dikiv1alpha1.OutputStatusFailed))

			// Verify error details are present
			var errorDetails map[string]any
			Expect(json.Unmarshal(updatedScan.Status.Outputs[0].Details.Raw, &errorDetails)).To(Succeed())
			Expect(errorDetails).To(HaveKey("error"))
			Expect(errorDetails["error"]).To(ContainSubstring("simulated ConfigMap creation failure"))
		})
	})

	Describe("NewReportExporter", func() {
		It("should create a new ReportExporter instance", func() {
			config := v1alpha1.ReportExporterConfiguration{
				ReportPath:         "/path/to/report.json",
				ComplianceScanName: "test-scan",
				Outputs:            []v1alpha1.Output{},
			}

			exporter := reportexporter.NewReportExporter(fakeClient, config)

			Expect(exporter).NotTo(BeNil())
			Expect(exporter.Client).To(Equal(fakeClient))
			Expect(exporter.Config).To(Equal(config))
		})
	})
})
