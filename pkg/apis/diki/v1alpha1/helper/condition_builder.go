// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// ConditionBuilder build a Condition.
type ConditionBuilder interface {
	WithOldCondition(old v1alpha1.Condition) ConditionBuilder
	WithStatus(status v1alpha1.ConditionStatus) ConditionBuilder
	WithReason(reason string) ConditionBuilder
	WithMessage(message string) ConditionBuilder
	WithTime(now time.Time) ConditionBuilder
	Build() (new v1alpha1.Condition, updated bool)
}

// defaultConditionBuilder build a Condition.
type defaultConditionBuilder struct {
	old           v1alpha1.Condition
	status        v1alpha1.ConditionStatus
	conditionType v1alpha1.ConditionType
	reason        string
	message       string
	time          time.Time
}

// NewConditionBuilder returns a ConditionBuilder for a specific condition.
func NewConditionBuilder(conditionType v1alpha1.ConditionType) ConditionBuilder {
	return &defaultConditionBuilder{
		conditionType: conditionType,
		time:          time.Now(),
	}
}

// WithOldCondition sets the old condition. It can be used to provide default values.
// The old's condition type is overridden to the one specified in the builder.
func (b *defaultConditionBuilder) WithOldCondition(old v1alpha1.Condition) ConditionBuilder {
	old.Type = b.conditionType
	b.old = old

	return b
}

// WithStatus sets the status of the condition.
func (b *defaultConditionBuilder) WithStatus(status v1alpha1.ConditionStatus) ConditionBuilder {
	b.status = status
	return b
}

// WithReason sets the reason of the condition.
func (b *defaultConditionBuilder) WithReason(reason string) ConditionBuilder {
	b.reason = reason
	return b
}

// WithMessage sets the message of the condition.
func (b *defaultConditionBuilder) WithMessage(message string) ConditionBuilder {
	b.message = message
	return b
}

// WithClock sets the time for the condition.
func (b *defaultConditionBuilder) WithTime(now time.Time) ConditionBuilder {
	b.time = now

	return b
}

// Build creates the condition and returns if there are modifications with the OldCondition.
// If OldCondition is provided:
// - Any changes to status set the `LastTransitionTime`
// - Any updates to the message or reason cause set `LastUpdateTime` to the current time.
func (b *defaultConditionBuilder) Build() (c v1alpha1.Condition, updated bool) {
	var (
		now       = metav1.Time{Time: b.time}
		emptyTime = metav1.Time{}
	)

	c = *b.old.DeepCopy()

	if c.LastTransitionTime == emptyTime {
		c.LastTransitionTime = now
	}

	if c.LastUpdateTime == emptyTime {
		c.LastUpdateTime = now
	}

	c.Type = b.conditionType

	if len(b.status) != 0 {
		c.Status = b.status
	} else if len(c.Status) == 0 {
		c.Status = v1alpha1.ConditionUnknown
	}

	if len(b.reason) > 0 {
		c.Reason = b.reason
	} else if len(c.Reason) == 0 {
		c.Reason = "Unspecified"
	}

	if len(b.message) > 0 {
		c.Message = b.message
	} else if len(c.Message) == 0 {
		c.Message = "No message given."
	}

	if c.Status != b.old.Status {
		c.LastTransitionTime = now
	}

	if c.Reason != b.old.Reason ||
		c.Message != b.old.Message {
		c.LastUpdateTime = now
	}

	return c, !apiequality.Semantic.DeepEqual(c, b.old)
}
