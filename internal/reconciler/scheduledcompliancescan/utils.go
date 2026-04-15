// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/internal/constants"
	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

func (r *Reconciler) setActiveScan(ctx context.Context, scheduledScan *v1alpha1.ScheduledComplianceScan, scan *v1alpha1.ComplianceScan, scheduleTime time.Time) error {
	patch := client.MergeFrom(scheduledScan.DeepCopy())
	scheduledScan.Status.Active = &corev1.ObjectReference{
		APIVersion: v1alpha1.SchemeGroupVersion.String(),
		Kind:       "ComplianceScan",
		Name:       scan.Name,
		UID:        scan.UID,
	}
	scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: scheduleTime}
	if err := r.Client.Status().Patch(ctx, scheduledScan, patch); err != nil {
		return fmt.Errorf("failed to update ScheduledComplianceScan status: %w", err)
	}
	return nil
}

func (r *Reconciler) deployComplianceScan(ctx context.Context, parent *v1alpha1.ScheduledComplianceScan, now time.Time) (*v1alpha1.ComplianceScan, error) {
	complianceScan := &v1alpha1.ComplianceScan{
		ObjectMeta: metav1.ObjectMeta{
			Name: childScanName(parent.Name, now),
			Labels: map[string]string{
				LabelScheduledComplianceScan: parent.Name,
				constants.LabelAppName:       constants.LabelValueDiki,
				constants.LabelAppManagedBy:  constants.LabelValueDikiOperator,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         v1alpha1.SchemeGroupVersion.String(),
					Kind:               "ScheduledComplianceScan",
					Name:               parent.Name,
					UID:                parent.UID,
					Controller:         ptr.To(true),
					BlockOwnerDeletion: ptr.To(true),
				},
			},
		},
		Spec: *parent.Spec.ScanTemplate.Spec.DeepCopy(),
	}

	if err := r.Client.Create(ctx, complianceScan); err != nil {
		return nil, fmt.Errorf("failed to create ComplianceScan: %w", err)
	}

	return complianceScan, nil
}

func sortByCreationTimestamp(scans []v1alpha1.ComplianceScan) {
	slices.SortFunc(scans, func(a, b v1alpha1.ComplianceScan) int {
		return a.CreationTimestamp.Compare(b.CreationTimestamp.Time)
	})
}

func (r *Reconciler) cleanupOldScans(ctx context.Context, log logr.Logger, scans []v1alpha1.ComplianceScan, limit int) {
	sortByCreationTimestamp(scans)
	excess := len(scans) - limit
	for i := 0; i < excess; i++ {
		if err := r.Client.Delete(ctx, &scans[i]); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to delete old ComplianceScan", "name", scans[i].Name)
		} else {
			log.Info("Deleted old ComplianceScan", "name", scans[i].Name)
		}
	}
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// childScanName generates a name for a child ComplianceScan by combining
// the parent name with a unix timestamp, truncating the parent name if
// necessary to stay within the DNS label length limit of 63 characters.
func childScanName(parentName string, t time.Time) string {
	suffix := "-" + strconv.FormatInt(t.Unix(), 10)
	maxParentLen := validation.DNS1035LabelMaxLength - len(suffix)
	if len(parentName) > maxParentLen {
		parentName = parentName[:maxParentLen]
	}
	return parentName + suffix
}
