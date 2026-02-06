// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	. "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1/helper"
)

var _ = Describe("Builder", func() {
	const (
		conditionType = v1alpha1.ConditionType("Test")
		fooStatus     = v1alpha1.ConditionStatus("Foo")
		bazReason     = "Baz"
		fubarMessage  = "FuBar"
	)

	var (
		defaultTime metav1.Time
	)

	BeforeEach(func() {
		defaultTime = metav1.NewTime(time.Unix(2, 2))
	})

	Describe("#RemoveComplianceRunCondition", func() {
		It("should remove the condition with the specified type", func() {
			conditions := []v1alpha1.Condition{
				{
					Type:               v1alpha1.ConditionType("bar"),
					Status:             v1alpha1.ConditionStatus("False"),
					Reason:             "ReasonB",
					Message:            "MessageB",
					LastTransitionTime: defaultTime,
				},
				{
					Type:               v1alpha1.ConditionType("foo"),
					Status:             v1alpha1.ConditionStatus("True"),
					Reason:             "ReasonA",
					Message:            "MessageA",
					LastTransitionTime: defaultTime,
				},
				{
					Type:               v1alpha1.ConditionType("baz"),
					Status:             v1alpha1.ConditionStatus("Unknown"),
					Reason:             "ReasonC",
					Message:            "MessageC",
					LastTransitionTime: defaultTime,
				},
			}

			result := RemoveComplianceRunCondition(conditions, v1alpha1.ConditionType("bar"))

			Expect(result).To(HaveLen(2))
			Expect(result[0].Type).To(Equal(v1alpha1.ConditionType("foo")))
			Expect(result[1].Type).To(Equal(v1alpha1.ConditionType("baz")))
		})

		It("should return the same conditions if the type is not found", func() {
			conditions := []v1alpha1.Condition{
				{
					Type:               v1alpha1.ConditionType("foo"),
					Status:             v1alpha1.ConditionStatus("True"),
					Reason:             "ReasonA",
					Message:            "MessageA",
					LastTransitionTime: defaultTime,
				},
			}

			result := RemoveComplianceRunCondition(conditions, v1alpha1.ConditionType("bar"))

			Expect(result).To(HaveLen(1))
			Expect(result[0].Type).To(Equal(v1alpha1.ConditionType("foo")))
		})

		It("should return an empty slice when removing the only condition", func() {
			conditions := []v1alpha1.Condition{
				{
					Type:               v1alpha1.ConditionType("foo"),
					Status:             v1alpha1.ConditionStatus("True"),
					Reason:             "ReasonA",
					Message:            "MessageA",
					LastTransitionTime: defaultTime,
				},
			}

			result := RemoveComplianceRunCondition(conditions, v1alpha1.ConditionType("foo"))

			Expect(result).To(BeEmpty())
		})
	})

	Describe("#UpdateComplianceRunConditions", func() {
		It("should add a new condition to an empty slice", func() {
			conditions := []v1alpha1.Condition{}

			result := UpdateComplianceRunConditions(conditions, conditionType, fooStatus, bazReason, fubarMessage, defaultTime.Time)

			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement(v1alpha1.Condition{
				Type:               conditionType,
				Status:             fooStatus,
				Reason:             bazReason,
				Message:            fubarMessage,
				LastTransitionTime: defaultTime,
				LastUpdateTime:     defaultTime,
			}))
		})

		It("should add a new condition when type doesn't exist", func() {
			conditions := []v1alpha1.Condition{{Type: v1alpha1.ConditionType("Existing")}}

			result := UpdateComplianceRunConditions(conditions, conditionType, fooStatus, bazReason, fubarMessage, defaultTime.Time)

			Expect(result).To(HaveLen(2))
			Expect(result).To(ContainElement(v1alpha1.Condition{
				Type:               conditionType,
				Status:             fooStatus,
				Reason:             bazReason,
				Message:            fubarMessage,
				LastTransitionTime: defaultTime,
				LastUpdateTime:     defaultTime,
			}))
		})

		It("should update an existing condition with status change", func() {
			oldTime := metav1.NewTime(time.Unix(5, 5))
			conditions := []v1alpha1.Condition{
				{
					Type:               conditionType,
					Status:             v1alpha1.ConditionStatus("OldStatus"),
					Reason:             "OldReason",
					Message:            "OldMessage",
					LastTransitionTime: oldTime,
				},
			}

			result := UpdateComplianceRunConditions(conditions, conditionType, fooStatus, bazReason, fubarMessage, defaultTime.Time)

			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement(v1alpha1.Condition{
				Type:               conditionType,
				Status:             fooStatus,
				Reason:             bazReason,
				Message:            fubarMessage,
				LastTransitionTime: defaultTime,
				LastUpdateTime:     defaultTime,
			}))
		})

		It("should preserve LastTransitionTime when status doesn't change", func() {
			oldTime := metav1.NewTime(time.Unix(5, 5))
			conditions := []v1alpha1.Condition{
				{
					Type:               conditionType,
					Status:             fooStatus,
					Reason:             "OldReason",
					Message:            "OldMessage",
					LastTransitionTime: oldTime,
				},
			}

			result := UpdateComplianceRunConditions(conditions, conditionType, fooStatus, bazReason, fubarMessage, defaultTime.Time)

			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement(v1alpha1.Condition{
				Type:               conditionType,
				Status:             fooStatus,
				Reason:             bazReason,
				Message:            fubarMessage,
				LastTransitionTime: oldTime,
				LastUpdateTime:     defaultTime,
			}))
		})

		It("should update the correct condition when multiple exist", func() {
			oldTime := metav1.NewTime(time.Unix(5, 5))
			conditions := []v1alpha1.Condition{
				{
					Type:               v1alpha1.ConditionType("TypeA"),
					Status:             v1alpha1.ConditionStatus("True"),
					Reason:             "ReasonA",
					Message:            "MessageA",
					LastTransitionTime: defaultTime,
				},
				{
					Type:               conditionType,
					Status:             v1alpha1.ConditionStatus("OldStatus"),
					Reason:             "OldReason",
					Message:            "OldMessage",
					LastTransitionTime: oldTime,
				},
				{
					Type:               v1alpha1.ConditionType("TypeC"),
					Status:             v1alpha1.ConditionStatus("Unknown"),
					Reason:             "ReasonC",
					Message:            "MessageC",
					LastTransitionTime: defaultTime,
				},
			}

			result := UpdateComplianceRunConditions(conditions, conditionType, fooStatus, bazReason, fubarMessage, defaultTime.Time)

			// TODO: Use consist of
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElement(v1alpha1.Condition{
				Type:               conditionType,
				Status:             fooStatus,
				Reason:             bazReason,
				Message:            fubarMessage,
				LastTransitionTime: defaultTime,
				LastUpdateTime:     defaultTime,
			}))
		})
	})
})
