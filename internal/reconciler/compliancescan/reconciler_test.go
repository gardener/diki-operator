// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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

		complianceScan *dikiv1alpha1.ComplianceScan
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
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

		request = reconcile.Request{NamespacedName: types.NamespacedName{
			Name: complianceScan.Name,
		}}
	})

	It("should successfully complete a compliance scan", func() {
		Expect(fakeClient.Create(ctx, complianceScan)).To(Succeed())

		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(reconcile.Result{}))

		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
		Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))
		Expect(complianceScan.Status.Conditions).To(ConsistOf(
			MatchFields(IgnoreExtras, Fields{
				"Type":    Equal(dikiv1alpha1.ConditionTypeCompleted),
				"Status":  Equal(dikiv1alpha1.ConditionTrue),
				"Reason":  Equal(compliancescan.ConditionReasonCompleted),
				"Message": Equal("ComplianceScan has completed successfully"),
			}),
		))
	})

	It("should handle failed compliance scan reconcile", func() {
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

	var _ = Describe("diki config ConfigMap", func() {
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
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))

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
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))

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
			Expect(res).To(Equal(reconcile.Result{}))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: complianceScan.Name}, complianceScan)).To(Succeed())
			Expect(complianceScan.Status.Phase).To(Equal(dikiv1alpha1.ComplianceScanCompleted))

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
