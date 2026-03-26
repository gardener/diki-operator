// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package compliancescan_test

import (
	"context"
	"errors"
	"net/http"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/gardener/diki-operator/internal/webhook/compliancescan"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("handler", func() {
	var (
		ctx = context.TODO()

		scheme            *runtime.Scheme
		decoder           admission.Decoder
		handler           admission.Handler
		request           admission.Request
		encoder           runtime.Encoder
		fakeClient        client.Client
		oldComplianceScan *v1alpha1.ComplianceScan
		complianceScan    *v1alpha1.ComplianceScan
		namespace         *v1.Namespace

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
		scheme = runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(authenticationv1.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		ctx = context.TODO()
		decoder = admission.NewDecoder(scheme)
		handler = &compliancescan.Handler{
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

		complianceScan = &v1alpha1.ComplianceScan{
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

		oldComplianceScan = complianceScan.DeepCopy()
	})

	Describe("#Handle", func() {
		Context("test updating the ComplianceScan resource", func() {

			It("should allow updating the ComplianceScan's metadata", func() {
				oldComplianceScanObj, err := runtime.Encode(encoder, oldComplianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldComplianceScanObj

				complianceScan.Labels = map[string]string{
					"foo": "bar",
				}

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				request.Operation = admissionv1.Update
				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should allow updating the ComplianceScan's status field", func() {
				oldComplianceScan.Status = v1alpha1.ComplianceScanStatus{
					Phase: v1alpha1.ComplianceScanRunning,
				}
				oldComplianceScanObj, err := runtime.Encode(encoder, oldComplianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldComplianceScanObj

				complianceScan.Status = v1alpha1.ComplianceScanStatus{
					Phase: v1alpha1.ComplianceScanCompleted,
				}
				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				request.Operation = admissionv1.Update
				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should deny updating the ComplianceScan's spec field", func() {
				oldComplianceScan.Spec.Rulesets = append(oldComplianceScan.Spec.Rulesets, v1alpha1.RulesetConfig{
					ID:      "ruleset-two",
					Version: "v0.1.0",
				})
				oldComplianceScanObj, err := runtime.Encode(encoder, oldComplianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldComplianceScanObj

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj
				request.Operation = admissionv1.Update

				responseForbidden.Result.Message = "updating the ComplianceScan spec is not permitted"
				Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
			})
		})

		Context("test creating the ComplianceScan resource", func() {
			It("should allow creating the ComplianceScan if it has no options configured", func() {
				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should allow creating a ComplianceScan containing a rule option pointing to an existing configMap", func() {
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
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
					Data: map[string]string{
						"ruleset-one-rules": "options",
					},
				}
				Expect(fakeClient.Create(ctx, namespace)).To(Succeed())
				Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should allow creating a ComplianceScan containing a ruleset option pointing to an existing configMap", func() {
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
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
					Data: map[string]string{
						"ruleset-one": "options",
					},
				}
				Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should allow creating a ComplianceScan containing both options that point to existing configMaps", func() {
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
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
					Data: map[string]string{
						"ruleset-one-rules": "options",
					},
				}
				Expect(fakeClient.Create(ctx, ruleOptionsConfigMap)).To(Succeed())

				rulesetOptionsConfigMap := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ruleset-options-configmap",
						Namespace: "kube-system",
					},
					Data: map[string]string{
						"ruleset-one": "options",
					},
				}
				Expect(fakeClient.Create(ctx, rulesetOptionsConfigMap)).To(Succeed())

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should forbid creating a ComplianceScan that has a rule option pointing to a non-existent configMap", func() {
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
					Rules: &v1alpha1.Options{
						ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
							Name:      "rule-options-configmap",
							Namespace: "kube-system",
						},
					},
				}

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				responseForbidden.Result.Message = "spec.rulesets[0].options.rules: Not found: \"the referenced configMap does not exist\""

				Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
			})

			It("should forbid creating a ComplianceScan that has a ruleset option pointing to a non-existent configMap", func() {
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
					Ruleset: &v1alpha1.Options{
						ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
							Name:      "ruleset-options-configmap",
							Namespace: "kube-system",
						},
					},
				}

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				responseForbidden.Result.Message = "spec.rulesets[0].options.ruleset: Not found: \"the referenced configMap does not exist\""

				Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
			})

			It("should forbid creating a ComplianceScan that contains at least one invalid reference", func() {
				complianceScan.Spec.Rulesets = append(complianceScan.Spec.Rulesets, v1alpha1.RulesetConfig{
					ID:      "ruleset-two",
					Version: "v0.0.0",
				})
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
					Ruleset: &v1alpha1.Options{
						ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
							Name:      "configmap",
							Namespace: "kube-system",
						},
					},
				}
				complianceScan.Spec.Rulesets[1].Options = &v1alpha1.RulesetOptions{
					Rules: &v1alpha1.Options{
						ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
							Name:      "non-existent-configmap",
							Namespace: "kube-system",
						},
					},
				}

				configMap := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap",
						Namespace: "kube-system",
					},
					Data: map[string]string{
						"ruleset-one": "options",
					},
				}
				Expect(fakeClient.Create(ctx, namespace)).To(Succeed())
				Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				responseForbidden.Result.Message = "spec.rulesets[1].options.rules: Not found: \"the referenced configMap does not exist\""

				Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
			})

			Context("test configMap references with data keys", func() {
				It("should allow creating a ComplianceScan containing a rule option with an existing configMap key", func() {
					complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
						Rules: &v1alpha1.Options{
							ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
								Name:      "configmap",
								Namespace: "kube-system",
								Key:       ptr.To("key"),
							},
						},
					}

					configMap := &v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "configmap",
							Namespace: "kube-system",
						},
						Data: map[string]string{
							"key": "some-value",
						},
					}

					Expect(fakeClient.Create(ctx, namespace)).To(Succeed())
					Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

					complianceScanObj, err := runtime.Encode(encoder, complianceScan)
					Expect(err).ToNot(HaveOccurred())
					request.Object.Raw = complianceScanObj

					Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
				})

				It("should allow creating a ComplianceScan containing a ruleset option with an existing configMap key", func() {
					complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
						Ruleset: &v1alpha1.Options{
							ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
								Name:      "configmap",
								Namespace: "kube-system",
								Key:       ptr.To("key"),
							},
						},
					}

					configMap := &v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "configmap",
							Namespace: "kube-system",
						},
						Data: map[string]string{
							"key": "some-value",
						},
					}
					Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

					complianceScanObj, err := runtime.Encode(encoder, complianceScan)
					Expect(err).ToNot(HaveOccurred())
					request.Object.Raw = complianceScanObj

					Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
				})

				It("should forbid creating a ComplianceScan containing a rule option reference with an non-existent configMap key", func() {
					complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
						Rules: &v1alpha1.Options{
							ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
								Name:      "configmap",
								Namespace: "kube-system",
								Key:       ptr.To("no-key"),
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
					Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

					complianceScanObj, err := runtime.Encode(encoder, complianceScan)
					Expect(err).ToNot(HaveOccurred())
					request.Object.Raw = complianceScanObj

					responseForbidden.Result.Message = "spec.rulesets[0].options.rules.key: Not found: \"the referenced key within the configMap does not exist\""
					Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
				})

				It("should forbid creating a ComplianceScan containing a ruleset option reference with an non-existent configMap key", func() {
					complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
						Ruleset: &v1alpha1.Options{
							ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
								Name:      "configmap",
								Namespace: "kube-system",
								Key:       ptr.To("no-key"),
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
					Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

					complianceScanObj, err := runtime.Encode(encoder, complianceScan)
					Expect(err).ToNot(HaveOccurred())
					request.Object.Raw = complianceScanObj

					responseForbidden.Result.Message = "spec.rulesets[0].options.ruleset.key: Not found: \"the referenced key within the configMap does not exist\""
					Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
				})
			})

			// TODO (georgibaltiev): Remove these testcases once the mutating admission webhook has been added.
			Context("test configMap references with default keys", func() {
				It("should forbid creating a ComplianceScan when the configMap does not contain the default rule options key", func() {
					complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
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
						Data: map[string]string{
							"some-other-key": "options",
						},
					}
					Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

					complianceScanObj, err := runtime.Encode(encoder, complianceScan)
					Expect(err).ToNot(HaveOccurred())
					request.Object.Raw = complianceScanObj

					responseForbidden.Result.Message = "spec.rulesets[0].options.rules.key: Not found: \"the referenced key within the configMap does not exist\""
					Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
				})

				It("should forbid creating a ComplianceScan when the configMap does not contain the default ruleset options key", func() {
					complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
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
						Data: map[string]string{
							"some-other-key": "options",
						},
					}
					Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

					complianceScanObj, err := runtime.Encode(encoder, complianceScan)
					Expect(err).ToNot(HaveOccurred())
					request.Object.Raw = complianceScanObj

					responseForbidden.Result.Message = "spec.rulesets[0].options.ruleset.key: Not found: \"the referenced key within the configMap does not exist\""
					Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
				})
			})

			It("should concatenate multiple errors", func() {
				complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
					Ruleset: &v1alpha1.Options{
						ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
							Name:      "configmap",
							Namespace: "kube-system",
							Key:       ptr.To("no-key"),
						},
					},
				}
				complianceScan.Spec.Rulesets = append(complianceScan.Spec.Rulesets, v1alpha1.RulesetConfig{
					ID:      "ruleset-two",
					Version: "v0.0.0",
					Options: &v1alpha1.RulesetOptions{
						Rules: &v1alpha1.Options{
							ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
								Name:      "non-existent-configmap",
								Namespace: "kube-system",
							},
						},
					},
				})

				configMap := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap",
						Namespace: "kube-system",
					},
					Data: map[string]string{},
				}
				Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

				complianceScanObj, err := runtime.Encode(encoder, complianceScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = complianceScanObj

				responseForbidden.Result.Message = "[spec.rulesets[0].options.ruleset.key: Not found: \"the referenced key within the configMap does not exist\", spec.rulesets[1].options.rules: Not found: \"the referenced configMap does not exist\"]"
				Expect(handler.Handle(ctx, request)).To(Equal(responseForbidden))
			})
		})

		It("should return an internal error when the configMap get fails with a non-NotFound error", func() {
			fakeErr := errors.New("internal server error")
			interceptedClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
						return fakeErr
					},
				}).Build()

			handler = &compliancescan.Handler{
				Decoder: decoder,
				Client:  interceptedClient,
			}

			complianceScan.Spec.Rulesets[0].Options = &v1alpha1.RulesetOptions{
				Rules: &v1alpha1.Options{
					ConfigMapRef: &v1alpha1.OptionsConfigMapRef{
						Name:      "configmap",
						Namespace: "kube-system",
					},
				},
			}

			complianceScanObj, err := runtime.Encode(encoder, complianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = complianceScanObj

			response := handler.Handle(ctx, request)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("Internal error"))
		})

		It("should allow requests for operations other than Create and Update", func() {
			complianceScanObj, err := runtime.Encode(encoder, complianceScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = complianceScanObj

			request.Operation = admissionv1.Delete
			Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
		})
	})
})
