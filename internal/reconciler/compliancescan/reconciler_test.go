// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler_test

import (
	"context"
	"errors"
	"time"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	compliancescan "github.com/gardener/diki-operator/internal/reconciler/compliancescan"
	configv1alpha1 "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	dikiinstall "github.com/gardener/diki-operator/pkg/apis/diki/install"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("Controller", func() {
	var (
		ctx = logf.IntoContext(context.Background(), logzap.New(logzap.WriteTo(GinkgoWriter)))

		cr         *compliancescan.Reconciler
		fakeClient client.Client
		fakeConfig *rest.Config

		request reconcile.Request

		scheme         *runtime.Scheme
		complianceScan *dikiv1alpha1.ComplianceScan
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(dikiinstall.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).Build()
		fakeConfig = &rest.Config{
			Host: "foo",
		}
		cr = &compliancescan.Reconciler{
			Client:       fakeClient,
			SourceClient: fakeClient,
			RESTConfig:   fakeConfig,
			Config: configv1alpha1.ComplianceScanConfig{
				SyncPeriod: &metav1.Duration{Duration: time.Hour},
				DikiRunner: configv1alpha1.DikiRunnerConfig{
					PodCompletionTimeout: &metav1.Duration{Duration: time.Second * 5},
				},
			},
		}

		complianceScan = &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "compliancescan",
				UID:  types.UID("1"),
			},
			Spec: dikiv1alpha1.ComplianceScanSpec{
				Rulesets: []dikiv1alpha1.RulesetConfig{
					{
						ID:      "FAKE",
						Version: "FAKE",
					},
				},
			},
		}

		request = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: complianceScan.Name,
			},
		}
	})

	Describe("early exit paths", func() {
		It("should stop reconciling when ComplianceScan is not found", func() {
			res, err := cr.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should stop reconciling when the ComplianceScan is already failed", func() {
			complianceScan.Status.Phase = dikiv1alpha1.ComplianceScanFailed
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())
			Expect(fakeClient.Status().Update(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should stop reconciling when the ComplianceScan is already completed", func() {
			complianceScan.Status.Phase = dikiv1alpha1.ComplianceScanCompleted
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())
			Expect(fakeClient.Status().Update(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
		})

		It("should return error when ComplianceScan Get fails with non-NotFound error", func() {
			cr.Client = fake.NewClientBuilder().
				WithScheme(scheme).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if _, ok := obj.(*dikiv1alpha1.ComplianceScan); ok {
							return errors.New("api-server-error")
						}
						return c.Get(ctx, key, obj, opts...)
					},
				}).Build()

			res, err := cr.Reconcile(ctx, request)
			Expect(err).To(MatchError(ContainSubstring("api-server-error")))
			Expect(res).To(Equal(reconcile.Result{}))
		})
	})

	Describe("deploy resources", func() {
		It("should create and set the ComplianceScan's phase to Running on its first reconcile", func() {
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))
			Expect(complianceScan.Status.Conditions).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeCompleted),
					"Status":  Equal(dikiv1alpha1.ConditionFalse),
					"Reason":  Equal(compliancescan.ConditionReasonRunning),
					"Message": Equal("ComplianceScan is running"),
				}),
			))
		})

		It("should set the ComplianceScan's phase to Failed when patchRunning fails", func() {
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			cr.Client = fake.NewClientBuilder().
				WithScheme(fakeClient.Scheme()).
				WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).
				WithObjects(complianceScan).
				WithInterceptorFuncs(interceptor.Funcs{
					SubResourcePatch: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
						var (
							cr        = obj.(*dikiv1alpha1.ComplianceScan)
							fakeError = errors.New("err-foo")
						)

						if cr.Status.Phase == dikiv1alpha1.ComplianceScanRunning {
							return fakeError
						}

						return client.SubResource(subResourceName).Patch(ctx, obj, patch, opts...)
					},
				}).Build()

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(cr.Client.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanFailed))
			Expect(complianceScan.Status.Conditions).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Reason":  Equal(compliancescan.ConditionReasonFailed),
					"Message": Equal("ComplianceScan failed with error: failed to update ComplianceScan status to Running: err-foo"),
				}),
			))
		})

		Describe("diki run Job", func() {
			var jobList *batchv1.JobList

			BeforeEach(func() {
				jobList = &batchv1.JobList{}
				Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())
			})

			It("should create a Job with the correct spec", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList, client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": string(complianceScan.UID)})).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				job := jobList.Items[0]
				expectedLabels := map[string]string{
					"app.kubernetes.io/name":                  "diki",
					"app.kubernetes.io/managed-by":            "diki-operator",
					"compliancescan.diki.gardener.cloud/uid":  string(complianceScan.UID),
					"compliancescan.diki.gardener.cloud/name": complianceScan.Name,
				}

				Expect(job.Labels).To(Equal(expectedLabels))

				Expect(job.Spec).To(MatchFields(IgnoreExtras, Fields{
					"BackoffLimit": PointTo(Equal(int32(0))),
					"Suspend":      PointTo(BeFalse()),
					"Template": MatchFields(IgnoreExtras, Fields{
						"ObjectMeta": MatchFields(IgnoreExtras, Fields{
							"Labels": Equal(expectedLabels),
						}),
						"Spec": MatchFields(IgnoreExtras, Fields{
							"ActiveDeadlineSeconds": PointTo(Equal(int64(5))),
							"ServiceAccountName":    Equal("diki-run"),
							"RestartPolicy":         Equal(corev1.RestartPolicyNever),
							"Containers": ConsistOf(
								MatchFields(IgnoreExtras, Fields{
									"Name": Equal("diki-scan"),
									"Args": Equal([]string{
										"run",
										"--config=/config/config.yaml",
										"--all",
										"--output=/report/report.json",
									}),
									"VolumeMounts": ConsistOf(
										MatchFields(IgnoreExtras, Fields{
											"Name":      Equal("diki-config"),
											"MountPath": Equal("/config"),
											"ReadOnly":  BeTrue(),
										}),
										MatchFields(IgnoreExtras, Fields{
											"Name":      Equal("diki-report"),
											"MountPath": Equal("/report"),
										}),
									),
									"SecurityContext": PointTo(MatchFields(IgnoreExtras, Fields{
										"AllowPrivilegeEscalation": PointTo(BeFalse()),
										"ReadOnlyRootFilesystem":   PointTo(BeTrue()),
										"Privileged":               PointTo(BeFalse()),
										"Capabilities": PointTo(MatchFields(IgnoreExtras, Fields{
											"Drop": ConsistOf(corev1.Capability("ALL")),
										})),
									})),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Name": Equal("report-exporter"),
									"Args": Equal([]string{
										"--config=/config/exporter-config.yaml",
									}),
									"VolumeMounts": ConsistOf(
										MatchFields(IgnoreExtras, Fields{
											"Name":      Equal("diki-report"),
											"MountPath": Equal("/report"),
											"ReadOnly":  BeTrue(),
										}),
										MatchFields(IgnoreExtras, Fields{
											"Name":      Equal("diki-config"),
											"MountPath": Equal("/config"),
											"ReadOnly":  BeTrue(),
										}),
									),
									"SecurityContext": PointTo(MatchFields(IgnoreExtras, Fields{
										"AllowPrivilegeEscalation": PointTo(BeFalse()),
										"ReadOnlyRootFilesystem":   PointTo(BeTrue()),
										"Privileged":               PointTo(BeFalse()),
										"Capabilities": PointTo(MatchFields(IgnoreExtras, Fields{
											"Drop": ConsistOf(corev1.Capability("ALL")),
										})),
									})),
								}),
							),
							"Volumes": ConsistOf(
								MatchFields(IgnoreExtras, Fields{
									"Name": Equal("diki-config"),
									"VolumeSource": MatchFields(IgnoreExtras, Fields{
										"ConfigMap": PointTo(MatchFields(IgnoreExtras, Fields{
											"LocalObjectReference": MatchFields(IgnoreExtras, Fields{
												"Name": Equal(compliancescan.ConfigMapNamePrefix + string(complianceScan.UID)),
											}),
											"DefaultMode": PointTo(Equal(int32(0440))),
										})),
									}),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Name": Equal("diki-report"),
									"VolumeSource": MatchFields(IgnoreExtras, Fields{
										"EmptyDir": Not(BeNil()),
									}),
								}),
							),
							"Tolerations": ConsistOf(
								MatchFields(IgnoreExtras, Fields{
									"Effect":   Equal(corev1.TaintEffectNoSchedule),
									"Operator": Equal(corev1.TolerationOpExists),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Effect":   Equal(corev1.TaintEffectNoExecute),
									"Operator": Equal(corev1.TolerationOpExists),
								}),
							),
							"SecurityContext": PointTo(MatchFields(IgnoreExtras, Fields{
								"RunAsNonRoot": PointTo(BeTrue()),
								"FSGroup":      PointTo(Equal(int64(65532))),
								"RunAsUser":    PointTo(Equal(int64(65532))),
								"RunAsGroup":   PointTo(Equal(int64(65532))),
								"SeccompProfile": PointTo(MatchFields(IgnoreExtras, Fields{
									"Type": Equal(corev1.SeccompProfileTypeRuntimeDefault),
								})),
							})),
						}),
					}),
				}))
			})

			It("should handle failed Job creation", func() {
				interceptingClient := fake.NewClientBuilder().
					WithScheme(fakeClient.Scheme()).
					WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).
					WithObjects(complianceScan).
					WithInterceptorFuncs(interceptor.Funcs{
						Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
							if _, ok := obj.(*batchv1.Job); ok {
								return errors.New("create-failed")
							}
							return c.Create(ctx, obj, opts...)
						},
					}).Build()
				cr.Client = interceptingClient
				cr.SourceClient = interceptingClient

				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{}))

				Expect(cr.Client.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
				Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanFailed))
				Expect(complianceScan.Status.Conditions).To(ContainElement(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
						"Status":  Equal(dikiv1alpha1.ConditionTrue),
						"Message": ContainSubstring("create-failed"),
					}),
				))
			})

			It("should create a Job with kubeconfig projected volume when kubeconfig is set", func() {
				cr.Config.DikiRunner.Kubeconfig = &configv1alpha1.KubeconfigConfig{
					SecretRef: configv1alpha1.SecretRef{
						Name: "target-kubeconfig",
					},
				}

				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList, client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": string(complianceScan.UID)})).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				job := jobList.Items[0]
				Expect(job.Spec.Template.Spec.Volumes).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("kubeconfig"),
					"VolumeSource": MatchFields(IgnoreExtras, Fields{
						"Projected": PointTo(MatchFields(IgnoreExtras, Fields{
							"DefaultMode": PointTo(Equal(int32(0440))),
							"Sources": ConsistOf(MatchFields(IgnoreExtras, Fields{
								"Secret": PointTo(MatchFields(IgnoreExtras, Fields{
									"LocalObjectReference": MatchFields(IgnoreExtras, Fields{
										"Name": Equal("target-kubeconfig"),
									}),
									"Items": ConsistOf(MatchFields(IgnoreExtras, Fields{
										"Key":  Equal("kubeconfig"),
										"Path": Equal("kubeconfig"),
									})),
									"Optional": PointTo(BeFalse()),
								})),
							})),
						})),
					}),
				})))
				Expect(job.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("kubeconfig"),
					"MountPath": Equal("/var/run/secrets/gardener.cloud/shoot/generic-kubeconfig"),
					"ReadOnly":  BeTrue(),
				})))
			})

			It("should create a Job with projected volume containing both kubeconfig and token when both refs are set", func() {
				cr.Config.DikiRunner.Kubeconfig = &configv1alpha1.KubeconfigConfig{
					SecretRef: configv1alpha1.SecretRef{
						Name: "target-kubeconfig",
					},
					TokenSecretRef: &configv1alpha1.SecretRef{
						Name: "target-token",
					},
				}

				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList, client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": string(complianceScan.UID)})).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				job := jobList.Items[0]
				Expect(job.Spec.Template.Spec.Volumes).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("kubeconfig"),
					"VolumeSource": MatchFields(IgnoreExtras, Fields{
						"Projected": PointTo(MatchFields(IgnoreExtras, Fields{
							"DefaultMode": PointTo(Equal(int32(0440))),
							"Sources": ConsistOf(
								MatchFields(IgnoreExtras, Fields{
									"Secret": PointTo(MatchFields(IgnoreExtras, Fields{
										"LocalObjectReference": MatchFields(IgnoreExtras, Fields{
											"Name": Equal("target-kubeconfig"),
										}),
										"Items": ConsistOf(MatchFields(IgnoreExtras, Fields{
											"Key":  Equal("kubeconfig"),
											"Path": Equal("kubeconfig"),
										})),
										"Optional": PointTo(BeFalse()),
									})),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Secret": PointTo(MatchFields(IgnoreExtras, Fields{
										"LocalObjectReference": MatchFields(IgnoreExtras, Fields{
											"Name": Equal("target-token"),
										}),
										"Items": ConsistOf(MatchFields(IgnoreExtras, Fields{
											"Key":  Equal("token"),
											"Path": Equal("token"),
										})),
										"Optional": PointTo(BeFalse()),
									})),
								}),
							),
						})),
					}),
				})))
				Expect(job.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("kubeconfig"),
					"MountPath": Equal("/var/run/secrets/gardener.cloud/shoot/generic-kubeconfig"),
					"ReadOnly":  BeTrue(),
				})))
			})
		})
	})

	Describe("check Job status", func() {
		var (
			dikiRunJob        *batchv1.Job
			fakeClientBuilder *fake.ClientBuilder
		)
		BeforeEach(func() {
			complianceScan.Status.Phase = dikiv1alpha1.ComplianceScanRunning
			complianceScan.Status.Conditions = []dikiv1alpha1.Condition{
				{
					Type:               dikiv1alpha1.ConditionTypeCompleted,
					Status:             dikiv1alpha1.ConditionFalse,
					Reason:             compliancescan.ConditionReasonRunning,
					Message:            "ComplianceScan is running",
					LastTransitionTime: metav1.Now(),
					LastUpdateTime:     metav1.Now(),
				},
			}

			dikiRunJob = &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name: compliancescan.JobNamePrefix + string(complianceScan.UID),
					Labels: map[string]string{
						"diki.gardener.cloud/compliancescan": "1",
					},
				},
			}

			fakeClientBuilder = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&dikiv1alpha1.ComplianceScan{})
		})

		DescribeTable("should reconcile correctly when Job is deployed",
			func(jobConditions []batchv1.JobCondition, suspend *bool, expectedResult reconcile.Result, expectedPhase dikiv1alpha1.ComplianceScanPhase, conditionMatcher gomegatypes.GomegaMatcher) {
				dikiRunJob.Status.Conditions = jobConditions
				dikiRunJob.Spec.Suspend = suspend
				fakeClient = fakeClientBuilder.WithObjects(complianceScan, dikiRunJob).Build()
				cr.Client = fakeClient
				cr.SourceClient = fakeClient

				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(expectedResult))

				Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
				Expect(complianceScan.Status.Phase).To(Equal(expectedPhase))
				if conditionMatcher != nil {
					Expect(complianceScan.Status.Conditions).To(ContainElement(conditionMatcher))
				}
			},
			Entry("Job succeeds",
				[]batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}},
				nil,
				reconcile.Result{},
				dikiv1alpha1.ComplianceScanCompleted,
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal(dikiv1alpha1.ConditionTypeCompleted),
					"Status": Equal(dikiv1alpha1.ConditionTrue),
					"Reason": Equal(compliancescan.ConditionReasonCompleted),
				}),
			),
			Entry("Job fails",
				[]batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "BackoffLimitExceeded"}},
				nil,
				reconcile.Result{},
				dikiv1alpha1.ComplianceScanFailed,
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": Equal("ComplianceScan failed with error: job failed: BackoffLimitExceeded"),
				}),
			),
			Entry("Job is still running",
				nil,
				nil,
				reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval},
				dikiv1alpha1.ComplianceScanRunning,
				nil,
			),
			Entry("Job is suspended",
				nil,
				ptr.To(true),
				reconcile.Result{},
				dikiv1alpha1.ComplianceScanFailed,
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": ContainSubstring("job is unexpectedly suspended"),
				}),
			),
		)

		It("should set phase to Completed when Job succeeds and there are no outputs", func() {
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			}

			fakeClient = fakeClientBuilder.WithObjects(complianceScan, dikiRunJob).Build()
			cr.Client = fakeClient
			cr.SourceClient = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeCompleted),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": ContainSubstring("ComplianceScan has completed successfully"),
				}),
			))
		})

		It("should set phase to Completed when Job succeeds and all outputs are successful", func() {
			complianceScan.Status.Outputs = []dikiv1alpha1.OutputStatus{
				{OutputName: "output-1", Phase: dikiv1alpha1.OutputStatusCompleted},
				{OutputName: "output-2", Phase: dikiv1alpha1.OutputStatusCompleted},
			}
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			}

			fakeClient = fakeClientBuilder.WithObjects(complianceScan, dikiRunJob).Build()
			cr.Client = fakeClient
			cr.SourceClient = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeCompleted),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": ContainSubstring("ComplianceScan has completed successfully"),
				}),
			))
		})

		It("should set phase to Failed when Job succeeds but outputs have failed", func() {
			complianceScan.Status.Outputs = []dikiv1alpha1.OutputStatus{
				{OutputName: "output-1", Phase: dikiv1alpha1.OutputStatusCompleted},
				{OutputName: "output-2", Phase: dikiv1alpha1.OutputStatusFailed},
			}
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			}

			fakeClient = fakeClientBuilder.WithObjects(complianceScan, dikiRunJob).Build()
			cr.Client = fakeClient
			cr.SourceClient = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanFailed))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": ContainSubstring("1/2 output(s) failed: output-2"),
				}),
			))
		})

		It("should fail when the Job is not found", func() {
			fakeClient = fakeClientBuilder.WithObjects(complianceScan).Build()
			cr.Client = fakeClient
			cr.SourceClient = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanFailed))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": ContainSubstring("failed to get diki runner job"),
				}),
			))
		})

		It("should return error when patchCompleted fails", func() {
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{
					Type:   batchv1.JobComplete,
					Status: corev1.ConditionTrue,
				},
			}
			interceptingClient := fakeClientBuilder.
				WithObjects(complianceScan, dikiRunJob).
				WithInterceptorFuncs(interceptor.Funcs{
					SubResourcePatch: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
						cs := obj.(*dikiv1alpha1.ComplianceScan)
						if cs.Status.Phase == dikiv1alpha1.ComplianceScanCompleted {
							return errors.New("status-patch-failed")
						}
						return client.SubResource(subResourceName).Patch(ctx, obj, patch, opts...)
					},
				}).Build()
			cr.Client = interceptingClient
			cr.SourceClient = interceptingClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).To(MatchError(ContainSubstring("status-patch-failed")))
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(interceptingClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))
		})

		It("should return error when patchFailed fails", func() {
			interceptingClient := fakeClientBuilder.
				WithObjects(complianceScan).
				WithInterceptorFuncs(interceptor.Funcs{
					SubResourcePatch: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
						cs := obj.(*dikiv1alpha1.ComplianceScan)
						if cs.Status.Phase == dikiv1alpha1.ComplianceScanFailed {
							return errors.New("status-patch-failed")
						}
						return client.SubResource(subResourceName).Patch(ctx, obj, patch, opts...)
					},
				}).Build()
			cr.Client = interceptingClient
			cr.SourceClient = interceptingClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).To(MatchError(ContainSubstring("failed to update ComplianceScan status to Failed")))
			Expect(err).To(MatchError(ContainSubstring("failed to get diki runner job")))
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(interceptingClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))
		})
	})

	Describe("diki config ConfigMap", func() {
		var (
			defaultRulesetOptions = `
          foo: bar`
			defaultRuleOptions = `
          - ruleID: "1111"
            args:
              foo: bar
          - ruleID: "2222"
            args:
              foo: baz`
			setRulesetOptions = `
          foo: baz`
			setRuleOptions = `
          - ruleID: "1111"
            args:
              foo: baz
          - ruleID: "2222"
            args:
              foo: bar`
			optionsConfigMap *corev1.ConfigMap
			configMapList    *corev1.ConfigMapList

			disaConfigWith = func(version, rulesetOptions, ruleOptions string) string {
				return `
      - id: disa-kubernetes-stig
        name: DISA Kubernetes Security Technical Implementation Guide
        version: ` + version + `
        ruleOptions:` + ruleOptions + `
        args:` + rulesetOptions
			}
			secK8sConfigWith = func(version, rulesetOptions, ruleOptions string) string {
				return `
      - id: security-hardened-k8s
        name: Security Hardened Kubernetes Cluster
        version: ` + version + `
        ruleOptions:` + ruleOptions + `
        args:` + rulesetOptions
			}
			configFor = func(rulesets ...string) string {
				config := `providers:
  - id: managedk8s
    name: Managed Kubernetes
    metadata: {}
    rulesets:`

				rulesetsConfig := ""
				for _, ruleset := range rulesets {
					rulesetsConfig += ruleset
				}

				if len(rulesetsConfig) > 0 {
					config += rulesetsConfig
				} else {
					config += ` []`
				}
				return config + `
    args: null
`
			}
			configForWithKubeconfig = func(rulesets ...string) string {
				config := `providers:
  - id: managedk8s
    name: Managed Kubernetes
    metadata: {}
    rulesets:`

				rulesetsConfig := ""
				for _, ruleset := range rulesets {
					rulesetsConfig += ruleset
				}

				if len(rulesetsConfig) > 0 {
					config += rulesetsConfig
				} else {
					config += ` []`
				}
				return config + `
    args:
      kubeconfigPath: /var/run/secrets/gardener.cloud/shoot/generic-kubeconfig/kubeconfig
`
			}
		)

		BeforeEach(func() {
			optionsConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "options-configmap",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"disa-kubernetes-stig":        defaultRulesetOptions,
					"disa-kubernetes-stig-rules":  defaultRuleOptions,
					"security-hardened-k8s":       defaultRulesetOptions,
					"security-hardened-k8s-rules": defaultRuleOptions,
					"set-ruleset-options":         setRulesetOptions,
					"set-rule-options":            setRuleOptions,
				},
			}
			Expect(fakeClient.Create(ctx, optionsConfigMap)).To(Succeed())
			configMapList = &corev1.ConfigMapList{}
		})

		It("should create a diki config ConfigMap", func() {
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))

			Expect(fakeClient.List(ctx, configMapList,
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/name": "compliancescan"},
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": "1"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			configMap := configMapList.Items[0]

			Expect(configMap.Data).To(HaveKey("config.yaml"))
			Expect(configMap.Data["config.yaml"]).To(Equal(configFor()))
		})

		It("should create a diki config ConfigMap with kubeconfigPath when kubeconfig is set", func() {
			cr.Config.DikiRunner.Kubeconfig = &configv1alpha1.KubeconfigConfig{
				SecretRef: configv1alpha1.SecretRef{
					Name: "target-kubeconfig",
				},
			}
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.List(ctx, configMapList,
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/name": "compliancescan"},
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": "1"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			configMap := configMapList.Items[0]
			Expect(configMap.Data).To(HaveKey("config.yaml"))
			Expect(configMap.Data["config.yaml"]).To(Equal(configForWithKubeconfig()))
		})

		It("should create a diki config for all rulesets without options", func() {
			complianceScan.Spec.Rulesets = []dikiv1alpha1.RulesetConfig{
				{
					ID:      "disa-kubernetes-stig",
					Version: "v1",
				},
				{
					ID:      "security-hardened-k8s",
					Version: "v1",
				},
			}
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))

			Expect(fakeClient.List(ctx, configMapList,
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/name": "compliancescan"},
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": "1"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			var (
				configMap    = configMapList.Items[0]
				disaConfig   = disaConfigWith("v1", " null", " []")
				secK8sConfig = secK8sConfigWith("v1", " null", " []")
			)

			Expect(configMap.Data).To(HaveKey("config.yaml"))
			Expect(configMap.Data["config.yaml"]).To(Equal(configFor(disaConfig, secK8sConfig)))
		})

		It("should create a diki config for all rulesets with options", func() {
			complianceScan.Spec.Rulesets = []dikiv1alpha1.RulesetConfig{
				{
					ID:      "disa-kubernetes-stig",
					Version: "v1",
					Options: &dikiv1alpha1.RulesetOptions{
						Ruleset: &dikiv1alpha1.Options{
							ConfigMapRef: &dikiv1alpha1.OptionsConfigMapRef{
								Name:      "options-configmap",
								Namespace: "kube-system",
							},
						},
						Rules: &dikiv1alpha1.Options{
							ConfigMapRef: &dikiv1alpha1.OptionsConfigMapRef{
								Name:      "options-configmap",
								Namespace: "kube-system",
							},
						},
					},
				},
				{
					ID:      "security-hardened-k8s",
					Version: "v1",
					Options: &dikiv1alpha1.RulesetOptions{
						Ruleset: &dikiv1alpha1.Options{
							ConfigMapRef: &dikiv1alpha1.OptionsConfigMapRef{
								Name:      "options-configmap",
								Namespace: "kube-system",
								Key:       ptr.To("set-ruleset-options"),
							},
						},
						Rules: &dikiv1alpha1.Options{
							ConfigMapRef: &dikiv1alpha1.OptionsConfigMapRef{
								Name:      "options-configmap",
								Namespace: "kube-system",
								Key:       ptr.To("set-rule-options"),
							},
						},
					},
				},
			}
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))

			Expect(fakeClient.List(ctx, configMapList,
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/name": "compliancescan"},
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/uid": "1"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			var (
				configMap    = configMapList.Items[0]
				disaConfig   = disaConfigWith("v1", defaultRulesetOptions, defaultRuleOptions)
				secK8sConfig = secK8sConfigWith("v1", setRulesetOptions, setRuleOptions)
			)

			Expect(configMap.Data).To(HaveKey("config.yaml"))
			Expect(configMap.Data["config.yaml"]).To(Equal(configFor(disaConfig, secK8sConfig)))
		})
	})

	Describe("exporter config in ConfigMap", func() {
		var configMapList *corev1.ConfigMapList

		BeforeEach(func() {
			configMapList = &corev1.ConfigMapList{}
		})

		It("should create exporter config with waitForReport and no outputs", func() {
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.List(ctx, configMapList,
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/name": "compliancescan"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			configMap := configMapList.Items[0]
			Expect(configMap.Data).To(HaveKey("exporter-config.yaml"))
			Expect(configMap.Data["exporter-config.yaml"]).To(Equal(`apiVersion: exporter.diki.gardener.cloud/v1alpha1
complianceScanName: compliancescan
kind: ReportExporterConfiguration
outputs: null
reportPath: /report/report.json
waitForReport: true
`))
		})

		It("should create exporter config with resolved ReportOutput", func() {
			reportOutput := &dikiv1alpha1.ReportOutput{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-output",
				},
				Spec: dikiv1alpha1.ReportOutputSpec{
					Output: dikiv1alpha1.Output{
						ConfigMap: &dikiv1alpha1.OutputConfigMap{
							Namespace:  "kube-system",
							NamePrefix: "scan-report-",
						},
					},
				},
			}
			Expect(fakeClient.Create(ctx, reportOutput)).To(Succeed())

			complianceScan.Spec.Outputs = []dikiv1alpha1.ReportOutputRef{
				{Name: "my-output"},
			}
			Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.List(ctx, configMapList,
				client.MatchingLabels{"compliancescan.diki.gardener.cloud/name": "compliancescan"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			configMap := configMapList.Items[0]
			Expect(configMap.Data).To(HaveKey("exporter-config.yaml"))
			Expect(configMap.Data["exporter-config.yaml"]).To(Equal(`apiVersion: exporter.diki.gardener.cloud/v1alpha1
complianceScanName: compliancescan
kind: ReportExporterConfiguration
outputs:
  - config:
      namePrefix: scan-report-
      namespace: kube-system
    name: my-output
    type: ConfigMap
reportPath: /report/report.json
waitForReport: true
`))
		})
	})

})
