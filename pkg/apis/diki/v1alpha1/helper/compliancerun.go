// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"time"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// UpdateComplianceRunConditions updates or adds a ComplianceRun condition.
func UpdateComplianceRunConditions(conditions []v1alpha1.Condition, cType v1alpha1.ConditionType, status v1alpha1.ConditionStatus, reason, message string, now time.Time) []v1alpha1.Condition {
	builder := NewConditionBuilder(cType).
		WithStatus(status).
		WithReason(reason).
		WithMessage(message).
		WithTime(now)

	for i, condition := range conditions {
		if condition.Type == cType {

			conditions[i], _ = builder.WithOldCondition(condition).Build()
			return conditions
		}
	}

	new, _ := builder.Build()
	return append(conditions, new)
}

// RemoveComplianceRunCondition removes a ComplianceRun condition of the given type.
func RemoveComplianceRunCondition(conditions []v1alpha1.Condition, cType v1alpha1.ConditionType) []v1alpha1.Condition {
	newConditions := []v1alpha1.Condition{}
	for _, condition := range conditions {
		if condition.Type != cType {
			newConditions = append(newConditions, condition)
		}
	}
	return newConditions
}
