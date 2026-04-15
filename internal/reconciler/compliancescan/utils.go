// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/internal/constants"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	v1alpha1helper "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1/helper"
)

func (r *Reconciler) handleFailedScan(ctx context.Context, complianceScan *v1alpha1.ComplianceScan, log logr.Logger, err error) error {
	patch := client.MergeFrom(complianceScan.DeepCopy())
	complianceScan.Status.Phase = v1alpha1.ComplianceScanFailed
	complianceScan.Status.Conditions = v1alpha1helper.UpdateConditions(
		complianceScan.Status.Conditions,
		v1alpha1.ConditionTypeFailed,
		v1alpha1.ConditionTrue,
		ConditionReasonFailed,
		fmt.Sprintf("ComplianceScan failed with error: %s", err.Error()),
		time.Now(),
	)
	complianceScan.Status.Conditions = slices.DeleteFunc(complianceScan.Status.Conditions, func(c v1alpha1.Condition) bool {
		return c.Type == v1alpha1.ConditionTypeCompleted
	})

	if err2 := r.Client.Status().Patch(ctx, complianceScan, patch); err2 != nil {
		return fmt.Errorf("failed to update ComplianceScan status to Failed: %w, original error: %w", err2, err)
	}

	log.Info("Updated ComplianceScan phase to Failed", "error", err.Error())

	return nil
}

func (r *Reconciler) getLabels(complianceScan *v1alpha1.ComplianceScan) map[string]string {
	labels := map[string]string{
		constants.LabelAppName:      constants.LabelValueDiki,
		constants.LabelAppManagedBy: constants.LabelValueDikiOperator,
	}

	maps.Copy(labels, r.Config.DikiRunner.Labels)
	labels[ComplianceScanLabel] = string(complianceScan.UID)

	return labels
}

// func (r *Reconciler) getOwnerReference(job *batchv1.Job) []metav1.OwnerReference {
// 	return []metav1.OwnerReference{
// 		{
// 			APIVersion:         batchv1.SchemeGroupVersion.String(),
// 			Kind:               "Job",
// 			Name:               job.Name,
// 			UID:                job.UID,
// 			Controller:         ptr.To(true),
// 			BlockOwnerDeletion: ptr.To(true),
// 		},
// 	}
// }
