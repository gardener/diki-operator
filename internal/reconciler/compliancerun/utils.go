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

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	v1alpha1helper "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1/helper"
)

func (r *Reconciler) handleFailedRun(ctx context.Context, complianceRun *v1alpha1.ComplianceRun, log logr.Logger, err error) error {
	patch := client.MergeFrom(complianceRun.DeepCopy())
	complianceRun.Status.Phase = v1alpha1.ComplianceRunFailed
	complianceRun.Status.Conditions = v1alpha1helper.UpdateConditions(
		complianceRun.Status.Conditions,
		v1alpha1.ConditionTypeFailed,
		v1alpha1.ConditionTrue,
		ConditionReasonFailed,
		fmt.Sprintf("ComplianceRun failed with error: %s", err.Error()),
		time.Now(),
	)
	complianceRun.Status.Conditions = slices.DeleteFunc(complianceRun.Status.Conditions, func(c v1alpha1.Condition) bool {
		return c.Type == v1alpha1.ConditionTypeCompleted
	})

	if err2 := r.Client.Status().Patch(ctx, complianceRun, patch); err2 != nil {
		return fmt.Errorf("failed to update ComplianceRun status to Failed: %w, original error: %w", err2, err)
	}

	log.Info("Updated ComplianceRun phase to Failed", "error", err.Error())

	return nil
}

func (r *Reconciler) getLabels(complianceRun *v1alpha1.ComplianceRun) map[string]string {
	labels := map[string]string{
		LabelAppName:      LabelValueDiki,
		LabelAppManagedBy: LabelValueDikiOperator,
	}

	maps.Copy(labels, r.Config.DikiRunner.Labels)
	labels[ComplianceRunLabel] = string(complianceRun.UID)

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
