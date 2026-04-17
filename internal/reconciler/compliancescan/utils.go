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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
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

func (r *Reconciler) findDikiRunJob(ctx context.Context, complianceScanUID types.UID) (*batchv1.Job, error) {
	jobList := &batchv1.JobList{}
	if err := r.Client.List(ctx, jobList, client.InNamespace(r.Config.DikiRunner.Namespace), client.MatchingLabels{
		ComplianceScanLabel: string(complianceScanUID),
	}); err != nil {
		return nil, fmt.Errorf("failed to list diki runner jobs: %w", err)
	}

	if len(jobList.Items) == 0 {
		return nil, nil
	}

	return &jobList.Items[0], nil
}

func (r *Reconciler) getOwnerReference(job *batchv1.Job) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         batchv1.SchemeGroupVersion.String(),
			Kind:               "Job",
			Name:               job.Name,
			UID:                job.UID,
			Controller:         ptr.To(true),
			BlockOwnerDeletion: ptr.To(true),
		},
	}
}

func (r *Reconciler) upscaleDikiRunJob(job *batchv1.Job) {
	job.Spec.Parallelism = ptr.To(int32(1))
}
