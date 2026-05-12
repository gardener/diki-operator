// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan_test

import (
	"context"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/gardener/diki-operator/internal/webhook/scheduledcompliancescan"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("MutatingHandler", func() {
	var (
		ctx = context.TODO()

		scheme  *runtime.Scheme
		handler admission.Handler
		request admission.Request
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(authenticationv1.AddToScheme(scheme)).To(Succeed())

		handler = &scheduledcompliancescan.MutatingHandler{
			Decoder: admission.NewDecoder(scheme),
		}

		request = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
			},
		}
	})

	It("should set all defaults when fields are omitted", func() {
		scan := &v1alpha1.ScheduledComplianceScan{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1alpha1.ScheduledComplianceScanSpec{
				ScanTemplate: v1alpha1.ScheduledComplianceScanTemplate{
					Spec: v1alpha1.ComplianceScanSpec{
						Rulesets: []v1alpha1.RulesetConfig{{ID: "r1", Version: "v1"}},
					},
				},
			},
		}

		result := handleAndPatch(ctx, &request, handler, scan)
		Expect(result.Spec.Schedule).To(Equal("0 0 * * 0"))
		Expect(result.Spec.SuccessfulScansHistoryLimit).ToNot(BeNil())
		Expect(*result.Spec.SuccessfulScansHistoryLimit).To(Equal(int32(3)))
		Expect(result.Spec.FailedScansHistoryLimit).ToNot(BeNil())
		Expect(*result.Spec.FailedScansHistoryLimit).To(Equal(int32(1)))
	})

	It("should not override values that are already set", func() {
		scan := &v1alpha1.ScheduledComplianceScan{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1alpha1.ScheduledComplianceScanSpec{
				Schedule:                    "*/5 * * * *",
				SuccessfulScansHistoryLimit: ptr.To[int32](10),
				FailedScansHistoryLimit:     ptr.To[int32](5),
				ScanTemplate: v1alpha1.ScheduledComplianceScanTemplate{
					Spec: v1alpha1.ComplianceScanSpec{
						Rulesets: []v1alpha1.RulesetConfig{{ID: "r1", Version: "v1"}},
					},
				},
			},
		}

		result := handleAndPatch(ctx, &request, handler, scan)
		Expect(result.Spec.Schedule).To(Equal("*/5 * * * *"))
		Expect(*result.Spec.SuccessfulScansHistoryLimit).To(Equal(int32(10)))
		Expect(*result.Spec.FailedScansHistoryLimit).To(Equal(int32(5)))
	})

	It("should only default fields that are missing", func() {
		scan := &v1alpha1.ScheduledComplianceScan{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1alpha1.ScheduledComplianceScanSpec{
				Schedule: "*/10 * * * *",
				ScanTemplate: v1alpha1.ScheduledComplianceScanTemplate{
					Spec: v1alpha1.ComplianceScanSpec{
						Rulesets: []v1alpha1.RulesetConfig{{ID: "r1", Version: "v1"}},
					},
				},
			},
		}

		result := handleAndPatch(ctx, &request, handler, scan)
		Expect(result.Spec.Schedule).To(Equal("*/10 * * * *"))
		Expect(*result.Spec.SuccessfulScansHistoryLimit).To(Equal(int32(3)))
		Expect(*result.Spec.FailedScansHistoryLimit).To(Equal(int32(1)))
	})
})

func handleAndPatch(ctx context.Context, request *admission.Request, handler admission.Handler, scan *v1alpha1.ScheduledComplianceScan) *v1alpha1.ScheduledComplianceScan {
	raw, err := json.Marshal(scan)
	Expect(err).ToNot(HaveOccurred())
	request.Object.Raw = raw

	resp := handler.Handle(ctx, *request)
	Expect(resp.Allowed).To(BeTrue())

	// Complete serializes Patches into the Patch byte field.
	Expect(resp.Complete(*request)).To(Succeed())

	patched := raw
	if resp.Patch != nil {
		patch, err := jsonpatch.DecodePatch(resp.Patch)
		Expect(err).ToNot(HaveOccurred())
		patched, err = patch.Apply(raw)
		Expect(err).ToNot(HaveOccurred())
	}

	result := &v1alpha1.ScheduledComplianceScan{}
	Expect(json.Unmarshal(patched, result)).To(Succeed())
	return result
}
