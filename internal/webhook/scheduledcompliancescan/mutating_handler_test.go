// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan_test

import (
	"context"
	"encoding/json"

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

		expected := scan.DeepCopy()
		expected.Spec.Schedule = "0 0 * * 0"
		expected.Spec.SuccessfulScansHistoryLimit = ptr.To[int32](3)
		expected.Spec.FailedScansHistoryLimit = ptr.To[int32](1)

		resp := handle(ctx, &request, handler, scan)
		Expect(resp.Allowed).To(BeTrue())
		Expect(resp.Patches).To(ConsistOf(patchResponse(scan, expected).Patches))
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

		resp := handle(ctx, &request, handler, scan)
		Expect(resp.Allowed).To(BeTrue())
		Expect(resp.Patches).To(BeEmpty())
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

		expected := scan.DeepCopy()
		expected.Spec.SuccessfulScansHistoryLimit = ptr.To[int32](3)
		expected.Spec.FailedScansHistoryLimit = ptr.To[int32](1)

		resp := handle(ctx, &request, handler, scan)
		Expect(resp.Allowed).To(BeTrue())
		Expect(resp.Patches).To(ConsistOf(patchResponse(scan, expected).Patches))
	})
})

func handle(ctx context.Context, request *admission.Request, handler admission.Handler, scan *v1alpha1.ScheduledComplianceScan) admission.Response {
	raw, err := json.Marshal(scan)
	Expect(err).ToNot(HaveOccurred())
	request.Object.Raw = raw

	return handler.Handle(ctx, *request)
}

func patchResponse(original, mutated *v1alpha1.ScheduledComplianceScan) admission.Response {
	originalJSON, err := json.Marshal(original)
	Expect(err).ToNot(HaveOccurred())
	mutatedJSON, err := json.Marshal(mutated)
	Expect(err).ToNot(HaveOccurred())

	return admission.PatchResponseFromRaw(originalJSON, mutatedJSON)
}
