// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	. "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1/helper"
)

var _ = Describe("Builder", func() {
	const (
		conditionType      = v1alpha1.ConditionType("Test")
		unknownStatus      = v1alpha1.ConditionStatus("Unknown")
		fooStatus          = v1alpha1.ConditionStatus("Foo")
		bazReason          = "Baz"
		fubarMessage       = "FuBar"
		unspecifiedMessage = "No message given."
		unspecifiedReason  = "Unspecified"
	)

	var (
		defaultTime metav1.Time
	)

	BeforeEach(func() {
		defaultTime = metav1.NewTime(time.Unix(2, 2))
	})

	Describe("#Build", func() {
		var (
			result  v1alpha1.Condition
			updated bool
			bldr    ConditionBuilder
		)

		JustBeforeEach(func() {
			bldr = NewConditionBuilder(conditionType)
		})

		Context("empty condition", func() {
			JustBeforeEach(func() {
				result, updated = bldr.WithTime(defaultTime.Time).Build()
			})

			It("should mark the result as updated", func() {
				Expect(updated).To(BeTrue())
			})

			It("should return correct result", func() {
				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             unknownStatus,
					LastTransitionTime: defaultTime,
					LastUpdateTime:     defaultTime,
					Reason:             unspecifiedReason,
					Message:            unspecifiedMessage,
				}))
			})
		})

		Context("#WithStatus", func() {
			JustBeforeEach(func() {
				result, updated = bldr.
					WithTime(defaultTime.Time).
					WithStatus(fooStatus).
					Build()
			})

			It("should mark the result as updated", func() {
				Expect(updated).To(BeTrue())
			})

			It("should return correct result", func() {
				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             fooStatus,
					LastTransitionTime: defaultTime,
					LastUpdateTime:     defaultTime,
					Reason:             unspecifiedReason,
					Message:            unspecifiedMessage,
				}))
			})
		})

		Context("#WithReason", func() {
			DescribeTable("New condition", func(reason *string, expectedReason string) {
				if reason != nil {
					bldr.WithReason(*reason)
				}

				result, updated = bldr.
					WithTime(defaultTime.Time).
					Build()

				Expect(updated).To(BeTrue())

				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             unknownStatus,
					LastTransitionTime: defaultTime,
					LastUpdateTime:     defaultTime,
					Reason:             expectedReason,
					Message:            unspecifiedMessage,
				}))
			},
				Entry("reason is not set", nil, unspecifiedReason),
				Entry("empty reason is set", ptr.To(""), unspecifiedReason),
				Entry("reason is set", ptr.To(bazReason), bazReason),
			)

			DescribeTable("With old condition", func(reason *string, previousReason, expectedReason string) {
				lastUpdateTime := metav1.NewTime(time.Unix(11, 0))

				if reason != nil {
					bldr.WithReason(*reason)
				}

				result, updated = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     lastUpdateTime,
						Reason:             previousReason,
						Message:            fubarMessage,
					}).
					Build()

				if reason != nil && *reason != previousReason || previousReason == "" {
					Expect(updated).To(BeTrue())
					lastUpdateTime = defaultTime
				}

				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             fooStatus,
					LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
					LastUpdateTime:     lastUpdateTime,
					Reason:             expectedReason,
					Message:            fubarMessage,
				}))
			},
				Entry("reason is not set", nil, bazReason, bazReason),
				Entry("reason was previously empty", nil, "", unspecifiedReason),
				Entry("empty reason is set", ptr.To(""), "", unspecifiedReason),
				Entry("message is the same", ptr.To("ReasonA"), "ReasonA", "ReasonA"),
				Entry("message changed", ptr.To("ReasonA"), bazReason, "ReasonA"),
			)
		})

		Context("#WithMessage", func() {
			DescribeTable("New condition", func(message *string, expectedMessage string) {
				if message != nil {
					bldr.WithMessage(*message)
				}

				result, updated = bldr.
					WithTime(defaultTime.Time).
					Build()

				Expect(updated).To(BeTrue())

				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             unknownStatus,
					LastTransitionTime: defaultTime,
					LastUpdateTime:     defaultTime,
					Reason:             unspecifiedReason,
					Message:            expectedMessage,
				}))
			},
				Entry("message is not set", nil, unspecifiedMessage),
				Entry("empty message is set", ptr.To(""), unspecifiedMessage),
				Entry("message is set", ptr.To(fubarMessage), fubarMessage),
			)

			DescribeTable("With old condition", func(message *string, previousMessage, expectedMessage string) {
				lastUpdateTime := metav1.NewTime(time.Unix(11, 0))

				if message != nil {
					bldr.WithMessage(*message)
				}

				result, updated = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     lastUpdateTime,
						Reason:             bazReason,
						Message:            previousMessage,
					}).
					Build()

				if message != nil && *message != previousMessage || previousMessage == "" {
					Expect(updated).To(BeTrue())
					lastUpdateTime = defaultTime
				}

				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             fooStatus,
					LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
					LastUpdateTime:     lastUpdateTime,
					Reason:             bazReason,
					Message:            expectedMessage,
				}))
			},
				Entry("message is not set", nil, fubarMessage, fubarMessage),
				Entry("message was previously empty", nil, "", unspecifiedMessage),
				Entry("empty message is set", ptr.To(""), "", unspecifiedMessage),
				Entry("message is the same", ptr.To("another message"), "another message", "another message"),
				Entry("message changed", ptr.To("another message"), fubarMessage, "another message"),
			)
		})

		Context("#WithOldCondition", func() {
			JustBeforeEach(func() {
				result, updated = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					Build()
			})

			It("should mark the result as not updated", func() {
				Expect(updated).To(BeFalse())
			})

			It("should return correct result", func() {
				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             fooStatus,
					LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
					LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
					Reason:             bazReason,
					Message:            fubarMessage,
				}))
			})
		})

		Context("Full override", func() {
			JustBeforeEach(func() {
				result, updated = bldr.
					WithTime(defaultTime.Time).
					WithStatus("SomeNewStatus").
					WithMessage("Some message").
					WithReason("SomeNewReason").
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					Build()
			})

			It("should mark the result as updated", func() {
				Expect(updated).To(BeTrue())
			})

			It("should return correct result", func() {
				Expect(result).To(Equal(v1alpha1.Condition{
					Type:               conditionType,
					Status:             "SomeNewStatus",
					LastTransitionTime: defaultTime,
					LastUpdateTime:     defaultTime,
					Reason:             "SomeNewReason",
					Message:            "Some message",
				}))
			})
		})

		Context("LastTransitionTime", func() {
			It("should update last transition time when status is updated", func() {
				result, _ = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					WithStatus("SomeNewStatus").
					Build()

				Expect(result.LastTransitionTime).To(Equal(defaultTime))
			})

			It("should not update last transition time when status is not updated", func() {
				result, _ = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					Build()

				Expect(result.LastTransitionTime).To(Equal(metav1.NewTime(time.Unix(10, 0))))
			})
		})

		Context("LastUpdateTime", func() {

			It("should update LastUpdateTime when message is updated", func() {
				result, _ = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					WithMessage("Some message").
					Build()

				Expect(result.LastUpdateTime).To(Equal(defaultTime))
			})

			It("should update LastUpdateTime when reason is updated", func() {
				result, _ = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					WithReason("SomeNewReason").
					Build()

				Expect(result.LastUpdateTime).To(Equal(defaultTime))
			})

			It("should not update LastUpdateTime when codes, message and reason are not updated", func() {
				result, _ = bldr.
					WithTime(defaultTime.Time).
					WithOldCondition(v1alpha1.Condition{
						Type:               conditionType,
						Status:             fooStatus,
						LastTransitionTime: metav1.NewTime(time.Unix(10, 0)),
						LastUpdateTime:     metav1.NewTime(time.Unix(11, 0)),
						Reason:             bazReason,
						Message:            fubarMessage,
					}).
					Build()

				Expect(result.LastUpdateTime).To(Equal(metav1.NewTime(time.Unix(11, 0))))
			})
		})
	})
})
