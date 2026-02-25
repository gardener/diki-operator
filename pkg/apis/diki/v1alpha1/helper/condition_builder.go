// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"cmp"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// ConditionBuilder builds a Condition.
type ConditionBuilder struct {
	old           v1alpha1.Condition
	status        v1alpha1.ConditionStatus
	conditionType v1alpha1.ConditionType
	reason        string
	message       string
	time          time.Time
}

// NewConditionBuilder returns a ConditionBuilder for a specific condition.
func NewConditionBuilder(conditionType v1alpha1.ConditionType) *ConditionBuilder {
	return &ConditionBuilder{
		conditionType: conditionType,
		time:          time.Now(),
	}
}

// WithOldCondition sets the old condition. It can be used to provide default values.
// The old's condition type is overridden to the one specified in the builder.
func (cb *ConditionBuilder) WithOldCondition(old v1alpha1.Condition) *ConditionBuilder {
	old.Type = cb.conditionType
	cb.old = old

	return cb
}

// WithStatus sets the status of the condition.
func (cb *ConditionBuilder) WithStatus(status v1alpha1.ConditionStatus) *ConditionBuilder {
	cb.status = status
	return cb
}

// WithReason sets the reason of the condition.
func (cb *ConditionBuilder) WithReason(reason string) *ConditionBuilder {
	cb.reason = reason
	return cb
}

// WithMessage sets the message of the condition.
func (cb *ConditionBuilder) WithMessage(message string) *ConditionBuilder {
	cb.message = message
	return cb
}

// WithTime sets the time for the condition.
func (cb *ConditionBuilder) WithTime(time time.Time) *ConditionBuilder {
	cb.time = time

	return cb
}

// Build creates the condition and returns if there are modifications with the OldCondition.
// If OldCondition is provided:
// - Any changes to status set the `LastTransitionTime`
// - Any updates to the message or reason cause set `LastUpdateTime` to the current time.
func (cb *ConditionBuilder) Build() (c v1alpha1.Condition, updated bool) {
	var (
		time     = metav1.Time{Time: cb.time}
		zeroTime = metav1.Time{}
	)

	c = *cb.old.DeepCopy()

	if c.LastTransitionTime == zeroTime {
		c.LastTransitionTime = time
	}
	if c.LastUpdateTime == zeroTime {
		c.LastUpdateTime = time
	}

	c.Type = cb.conditionType
	c.Status = cmp.Or(cb.status, c.Status, v1alpha1.ConditionUnknown)
	c.Reason = cmp.Or(cb.reason, c.Reason, "Unspecified")
	c.Message = cmp.Or(cb.message, c.Message, "No message given.")

	if c.Status != cb.old.Status {
		c.LastTransitionTime = time
	}
	if c.Reason != cb.old.Reason ||
		c.Message != cb.old.Message {
		c.LastUpdateTime = time
	}

	return c, !apiequality.Semantic.DeepEqual(c, cb.old)
}
