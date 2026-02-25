// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"time"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// UpdateConditions updates or adds a ComplianceRun condition.
func UpdateConditions(conditions []v1alpha1.Condition, cType v1alpha1.ConditionType, status v1alpha1.ConditionStatus, reason, message string, time time.Time) []v1alpha1.Condition {
	builder := NewConditionBuilder(cType).
		WithStatus(status).
		WithReason(reason).
		WithMessage(message).
		WithTime(time)

	for i, condition := range conditions {
		if condition.Type == cType {
			conditions[i], _ = builder.WithOldCondition(condition).Build()
			return conditions
		}
	}

	c, _ := builder.Build()
	return append(conditions, c)
}
