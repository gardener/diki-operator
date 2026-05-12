// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan_test

import (
	"context"
	"net/http"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/gardener/diki-operator/internal/webhook/scheduledcompliancescan"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("handler", func() {
	var (
		ctx = context.TODO()

		scheme        *runtime.Scheme
		decoder       admission.Decoder
		handler       admission.Handler
		request       admission.Request
		encoder       runtime.Encoder
		fakeClient    client.Client
		scheduledScan *v1alpha1.ScheduledComplianceScan

		responseAllowed = admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: true,
				Result: &metav1.Status{
					Code: int32(http.StatusOK),
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
		handler = &scheduledcompliancescan.Handler{
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
					Resource: "ScheduledComplianceScan",
					Version:  "v1alpha1",
				},
				Operation: admissionv1.Create,
			},
		}

		scheduledScan = &v1alpha1.ScheduledComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scheduled-scan",
			},
			Spec: v1alpha1.ScheduledComplianceScanSpec{
				Schedule:                    "0 0 * * 0",
				SuccessfulScansHistoryLimit: ptr.To[int32](3),
				FailedScansHistoryLimit:     ptr.To[int32](1),
				ScanTemplate: v1alpha1.ScheduledComplianceScanTemplate{
					Spec: v1alpha1.ComplianceScanSpec{
						Rulesets: []v1alpha1.RulesetConfig{
							{
								ID:      "ruleset-one",
								Version: "v0.0.0",
							},
						},
					},
				},
			},
		}
	})

	Describe("#Handle", func() {
		Context("test creating the ScheduledComplianceScan resource", func() {
			It("should allow creating the ScheduledComplianceScan", func() {
				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should deny creating with an invalid cron schedule", func() {
				scheduledScan.Spec.Schedule = "not-a-cron"
				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj

				resp := handler.Handle(ctx, request)
				Expect(resp.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(ContainSubstring("spec.schedule"))
			})

			It("should deny creating with a negative successfulScansHistoryLimit", func() {
				scheduledScan.Spec.SuccessfulScansHistoryLimit = ptr.To[int32](-1)
				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj

				resp := handler.Handle(ctx, request)
				Expect(resp.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(ContainSubstring("spec.successfulScansHistoryLimit"))
			})

			It("should deny creating with a negative failedScansHistoryLimit", func() {
				scheduledScan.Spec.FailedScansHistoryLimit = ptr.To[int32](-1)
				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj

				resp := handler.Handle(ctx, request)
				Expect(resp.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(ContainSubstring("spec.failedScansHistoryLimit"))
			})
		})

		Context("test updating the ScheduledComplianceScan resource", func() {
			It("should deny updating the scanTemplate", func() {
				oldScheduledScan := scheduledScan.DeepCopy()
				oldScheduledScanObj, err := runtime.Encode(encoder, oldScheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldScheduledScanObj

				scheduledScan.Spec.ScanTemplate.Spec.Rulesets = append(scheduledScan.Spec.ScanTemplate.Spec.Rulesets, v1alpha1.RulesetConfig{
					ID:      "ruleset-two",
					Version: "v0.1.0",
				})

				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj
				request.Operation = admissionv1.Update

				resp := handler.Handle(ctx, request)
				Expect(resp.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(ContainSubstring("spec.scanTemplate"))
			})

			It("should deny updating with an invalid cron schedule", func() {
				oldScheduledScan := scheduledScan.DeepCopy()
				oldScheduledScanObj, err := runtime.Encode(encoder, oldScheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldScheduledScanObj

				scheduledScan.Spec.Schedule = "bad-cron"
				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj
				request.Operation = admissionv1.Update

				resp := handler.Handle(ctx, request)
				Expect(resp.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(ContainSubstring("spec.schedule"))
			})

			It("should deny updating with a negative history limit", func() {
				oldScheduledScan := scheduledScan.DeepCopy()
				oldScheduledScanObj, err := runtime.Encode(encoder, oldScheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldScheduledScanObj

				scheduledScan.Spec.FailedScansHistoryLimit = ptr.To[int32](-5)
				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj
				request.Operation = admissionv1.Update

				resp := handler.Handle(ctx, request)
				Expect(resp.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(ContainSubstring("spec.failedScansHistoryLimit"))
			})

			It("should allow updating fields other than scanTemplate", func() {
				oldScheduledScan := scheduledScan.DeepCopy()
				oldScheduledScanObj, err := runtime.Encode(encoder, oldScheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldScheduledScanObj

				scheduledScan.Spec.Schedule = "*/5 * * * *"
				scheduledScan.Spec.SuccessfulScansHistoryLimit = ptr.To[int32](5)

				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj
				request.Operation = admissionv1.Update

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})

			It("should allow updating the metadata", func() {
				oldScheduledScan := scheduledScan.DeepCopy()
				oldScheduledScanObj, err := runtime.Encode(encoder, oldScheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.OldObject.Raw = oldScheduledScanObj

				scheduledScan.Labels = map[string]string{"foo": "bar"}

				scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
				Expect(err).ToNot(HaveOccurred())
				request.Object.Raw = scheduledScanObj
				request.Operation = admissionv1.Update

				Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
			})
		})

		It("should allow requests for operations other than Create and Update", func() {
			scheduledScanObj, err := runtime.Encode(encoder, scheduledScan)
			Expect(err).ToNot(HaveOccurred())
			request.Object.Raw = scheduledScanObj

			request.Operation = admissionv1.Delete
			Expect(handler.Handle(ctx, request)).To(Equal(responseAllowed))
		})
	})
})
