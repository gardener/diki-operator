// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package compliancescan_test

import (
	"context"
	"net/http"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/gardener/diki-operator/internal/webhook/compliancescan"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("handler", func() {
	var (
		ctx = context.TODO()

		decoder        admission.Decoder
		log            logr.Logger
		handler        admission.Handler
		request        admission.Request
		encoder        runtime.Encoder
		fakeClient     client.Client
		ComplianceScan *v1alpha1.ComplianceScan
		namespace      *v1.Namespace

		responseAllowed = admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: true,
				Result: &metav1.Status{
					Code: int32(http.StatusOK),
				},
			},
		}

		responseForbidden = admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:   http.StatusForbidden,
					Reason: metav1.StatusReasonForbidden,
				},
			},
		}
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(authenticationv1.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		ctx = context.TODO()
		log = logger.MustNewZapLogger(logger.DebugLevel, logger.FormatJSON, logzap.WriteTo(GinkgoWriter))

		decoder = admission.NewDecoder(scheme)
		handler = &compliancescan.Handler{
			Logger:  log,
			Decoder: decoder,
			Client:  fakeClient,
		}

		encoder = &json.Serializer{}
		request = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				UserInfo: authenticationv1.UserInfo{
					Username: "diki-operator",
				},
				Resource: metav1.GroupVersionResource{
					Group:    "diki.gardener.cloud",
					Resource: "ComplianceScan",
					Version:  "v1alpha1",
				},
				Operation: admissionv1.Create,
			},
		}

		namespace = &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
		}

		ComplianceScan = &v1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "compliance-scan",
				Namespace: "diki",
			},
			Spec: v1alpha1.ComplianceScanSpec{
				Rulesets: []v1alpha1.RulesetConfig{
					{
						ID:      "ruleset-one",
						Version: "v0.0.0",
					},
				},
			},
		}
	})

	Describe("#Handle", func() {
		It("should deny updating a ComplianceScan", func() {
			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj
			request.Operation = admissionv1.Update

			responseForbidden.Result.Message = "updating ComplianceScan resources is not permitted"

			Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
		})

		It("should allow creating a ComplianceScan that has no options configured", func() {
			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj

			Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
		})

		It("should allow creating a ComplianceScan that has a ruleOption pointing to an existing configMap", func() {
			ComplianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
				Rules: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "configmap",
						Namespace: "kube-system",
					},
				},
			}

			configMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "configmap",
					Namespace: "kube-system",
				},
				Data: map[string]string{},
			}
			Expect(fakeClient.Create(ctx, namespace)).To(Succeed())
			Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj

			Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
		})

		It("should allow creating a ComplianceScan that has a rulesetOption pointing to an existing configMap", func() {
			ComplianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
				Ruleset: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "configmap",
						Namespace: "kube-system",
					},
				},
			}

			configMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "configmap",
					Namespace: "kube-system",
				},
			}
			Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj

			Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
		})

		It("should allow creating a ComplianceScan that has both options pointing to existing configMaps", func() {
			ComplianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
				Ruleset: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "ruleset-options-configmap",
						Namespace: "kube-system",
					},
				},
				Rules: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "rule-options-configmap",
						Namespace: "kube-system",
					},
				},
			}

			ruleOptionsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rule-options-configmap",
					Namespace: "kube-system",
				},
				Data: map[string]string{},
			}
			Expect(fakeClient.Create(ctx, ruleOptionsConfigMap)).To(Succeed())

			rulesetOptionsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ruleset-options-configmap",
					Namespace: "kube-system",
				},
				Data: map[string]string{},
			}
			Expect(fakeClient.Create(ctx, rulesetOptionsConfigMap)).To(Succeed())

			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj

			Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
		})

		It("should forbid creating a ComplianceScan that has a rule option pointing to a non-existent configMap", func() {
			ComplianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
				Rules: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "rule-options-configmap",
						Namespace: "kube-system",
					},
				},
			}

			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj

			responseForbidden.Result.Message = "spec.rulesets[0].options.rules: the referenced configMap does not exist"

			Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
		})

		It("should forbid creating a ComplianceScan that has a ruleset option pointing to a non-existent configMap", func() {
			ComplianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
				Ruleset: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "ruleset-options-configmap",
						Namespace: "kube-system",
					},
				},
			}

			ComplianceScanObj, err := runtime.Encode(encoder, ComplianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = ComplianceScanObj

			responseForbidden.Result.Message = "spec.rulesets[0].options.ruleset: the referenced configMap does not exist"

			Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
		})
	})
})
