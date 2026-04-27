// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package scheduledcompliancescan

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// MutatingHandler is an admission webhook handler that sets defaults on ScheduledComplianceScan resources.
type MutatingHandler struct {
	Decoder admission.Decoder
}

// Handle sets default values on ScheduledComplianceScan resources.
func (h *MutatingHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	scheduledScan := &dikiv1alpha1.ScheduledComplianceScan{}
	if err := json.Unmarshal(req.Object.Raw, scheduledScan); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	needsMutation := false

	if scheduledScan.Spec.Schedule == "" {
		scheduledScan.Spec.Schedule = "0 0 * * 0"
		needsMutation = true
	}
	if scheduledScan.Spec.SuccessfulScansHistoryLimit == nil {
		scheduledScan.Spec.SuccessfulScansHistoryLimit = ptr.To(int32(3))
		needsMutation = true
	}
	if scheduledScan.Spec.FailedScansHistoryLimit == nil {
		scheduledScan.Spec.FailedScansHistoryLimit = ptr.To(int32(1))
		needsMutation = true
	}

	if !needsMutation {
		return admission.Allowed("")
	}

	marshaledScan, err := json.Marshal(scheduledScan)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledScan)
}
