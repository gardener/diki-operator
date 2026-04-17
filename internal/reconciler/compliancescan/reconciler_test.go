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
			Client:     fakeClient,
			RESTConfig: fakeConfig,
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

	Describe("deploy resources", func() {
		It("should set phase to Running and create resources on first reconcile", func() {
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

		It("should handle failed status patch during deploy", func() {
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
					"Message": Equal("ComplianceScan failed with error: err-foo"),
				}),
			))
		})

		Describe("diki run Job", func() {
			var (
				jobList *batchv1.JobList
			)

			BeforeEach(func() {
				jobList = &batchv1.JobList{}
				Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())
			})

			It("should create a Job with correct metadata", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				job := jobList.Items[0]
				Expect(job.Labels).To(Equal(map[string]string{
					"app.kubernetes.io/name":           "diki",
					"app.kubernetes.io/managed-by":     "diki-operator",
					compliancescan.ComplianceScanLabel: string(complianceScan.UID),
				}))
			})

			It("should configure the diki-scan container", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				containers := jobList.Items[0].Spec.Template.Spec.Containers
				Expect(containers).To(HaveLen(1))
				Expect(containers[0].Name).To(Equal("diki-scan"))
				Expect(containers[0].Args).To(Equal([]string{
					"run",
					"--config=/config/config.yaml",
					"--all",
				}))
			})

			It("should configure volumes correctly", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				volumes := jobList.Items[0].Spec.Template.Spec.Volumes
				Expect(volumes).To(HaveLen(1))
				Expect(volumes[0].Name).To(Equal("diki-config"))
				Expect(volumes[0].VolumeSource.ConfigMap).NotTo(BeNil())
				Expect(volumes[0].VolumeSource.ConfigMap.Name).To(Equal(compliancescan.ConfigMapGenerateNamePrefix + string(complianceScan.UID)))
			})

			It("should set volume mounts on the container", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				containers := jobList.Items[0].Spec.Template.Spec.Containers
				Expect(containers[0].VolumeMounts).To(HaveLen(1))
				Expect(containers[0].VolumeMounts[0].Name).To(Equal("diki-config"))
				Expect(containers[0].VolumeMounts[0].MountPath).To(Equal("/config"))
				Expect(containers[0].VolumeMounts[0].ReadOnly).To(BeTrue())
			})

			It("should set pod spec fields correctly", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				podSpec := jobList.Items[0].Spec.Template.Spec
				Expect(podSpec.ActiveDeadlineSeconds).NotTo(BeNil())
				Expect(*podSpec.ActiveDeadlineSeconds).To(Equal(int64(5)))
				Expect(podSpec.ServiceAccountName).To(Equal("diki-run"))
				Expect(podSpec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
			})

			It("should set tolerations for NoSchedule and NoExecute", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				tolerations := jobList.Items[0].Spec.Template.Spec.Tolerations
				Expect(tolerations).To(HaveLen(2))
				Expect(tolerations[0].Effect).To(Equal(corev1.TaintEffectNoSchedule))
				Expect(tolerations[0].Operator).To(Equal(corev1.TolerationOpExists))
				Expect(tolerations[1].Effect).To(Equal(corev1.TaintEffectNoExecute))
				Expect(tolerations[1].Operator).To(Equal(corev1.TolerationOpExists))
			})

			It("should propagate labels to the pod template", func() {
				res, err := cr.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

				Expect(fakeClient.List(ctx, jobList,
					client.MatchingLabels{compliancescan.ComplianceScanLabel: string(complianceScan.UID)},
				)).To(Succeed())
				Expect(jobList.Items).To(HaveLen(1))

				Expect(jobList.Items[0].Spec.Template.Labels).To(Equal(map[string]string{
					"app.kubernetes.io/name":           "diki",
					"app.kubernetes.io/managed-by":     "diki-operator",
					compliancescan.ComplianceScanLabel: string(complianceScan.UID),
				}))
			})

			It("should handle failed Job creation", func() {
				cr.Client = fake.NewClientBuilder().
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
					Name:      "diki-runner-abc",
					Namespace: "kube-system",
					Labels: map[string]string{
						"diki.gardener.cloud/compliancescan": "1",
					},
				},
			}

			fakeClientBuilder = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).
				WithObjects(complianceScan, dikiRunJob)
		})

		It("should set Completed when Job succeeds", func() {
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{
					Type:   batchv1.JobComplete,
					Status: corev1.ConditionTrue,
				},
			}

			fakeClient = fakeClientBuilder.Build()
			cr.Client = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal(dikiv1alpha1.ConditionTypeCompleted),
					"Status": Equal(dikiv1alpha1.ConditionTrue),
					"Reason": Equal(compliancescan.ConditionReasonCompleted),
				}),
			))
		})

		It("should set Failed when Job fails", func() {
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{
					Type:    batchv1.JobFailed,
					Status:  corev1.ConditionTrue,
					Message: "BackoffLimitExceeded",
				},
			}

			fakeClient = fakeClientBuilder.Build()
			cr.Client = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanFailed))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": ContainSubstring("BackoffLimitExceeded"),
				}),
			))
		})

		It("should requeue when Job is still running", func() {
			fakeClient = fakeClientBuilder.Build()
			cr.Client = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{RequeueAfter: compliancescan.ReconciliationRequeueInterval}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanRunning))
		})

		It("should set Failed when no Job is found", func() {
			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).
				WithObjects(complianceScan).
				Build()
			cr.Client = fakeClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanFailed))
			Expect(complianceScan.Status.Conditions).To(ContainElement(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(dikiv1alpha1.ConditionTypeFailed),
					"Status":  Equal(dikiv1alpha1.ConditionTrue),
					"Message": Equal("ComplianceScan failed with error: job not found for ComplianceScan compliancescan"),
				}),
			))
		})

		It("should handle status patch failure on completion", func() {
			dikiRunJob.Status.Conditions = []batchv1.JobCondition{
				{
					Type:   batchv1.JobComplete,
					Status: corev1.ConditionTrue,
				},
			}

			interceptingClient := fakeClientBuilder.
				WithInterceptorFuncs(interceptor.Funcs{
					SubResourcePatch: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
						cs := obj.(*dikiv1alpha1.ComplianceScan)
						if cs.Status.Phase == dikiv1alpha1.ComplianceScanCompleted {
							return errors.New("patch-failed")
						}
						return client.SubResource(subResourceName).Patch(ctx, obj, patch, opts...)
					},
				}).Build()
			cr.Client = interceptingClient

			res, err := cr.Reconcile(ctx, request)
			Expect(err).To(MatchError("patch-failed"))
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
				client.MatchingLabels{"diki.gardener.cloud/compliancescan": "1"},
			)).To(Succeed())
			Expect(len(configMapList.Items)).To(Equal(1))

			configMap := configMapList.Items[0]

			Expect(configMap.Data).To(HaveKey("config.yaml"))
			Expect(configMap.Data["config.yaml"]).To(Equal(configFor()))
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
				client.MatchingLabels{"diki.gardener.cloud/compliancescan": "1"},
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
				client.MatchingLabels{"diki.gardener.cloud/compliancescan": "1"},
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

})
