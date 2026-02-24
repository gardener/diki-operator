// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"context"
	"fmt"
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

func (r *Reconciler) handleFailedRun(ctx context.Context, complianceRun *v1alpha1.ComplianceRun, err error) error {
	log := logf.FromContext(ctx)
	patch := client.MergeFrom(complianceRun.DeepCopy())
	complianceRun.Status.Phase = v1alpha1.ComplianceRunFailed
	// TODO(AleksandarSavchev): Update conditions here.

	if err2 := r.Client.Status().Patch(ctx, complianceRun, patch); err2 != nil {
		return fmt.Errorf("failed to update ComplianceRun status to Failed: %w", err2)
	}

	log.Info("Updated ComplianceRun phase to Failed", "name", complianceRun.Name, "error", err.Error())

	return nil
}

func (r *Reconciler) getLabels() map[string]string {
	labels := map[string]string{
		"origin": "diki-operator",
	}

	maps.Copy(labels, r.Config.DikiRunner.Labels)
	return labels
}

func (r *Reconciler) getOwnerReference(complianceRun *v1alpha1.ComplianceRun) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         v1alpha1.SchemeGroupVersion.String(),
			Kind:               "ComplianceRun",
			Name:               complianceRun.Name,
			UID:                complianceRun.UID,
			Controller:         ptr.To(true),
			BlockOwnerDeletion: ptr.To(true),
		},
	}
}
