// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package reconciler_test

import (
	"context"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	testclock "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	scheduledcompliancescan "github.com/gardener/diki-operator/internal/reconciler/scheduledcompliancescan"
	dikiinstall "github.com/gardener/diki-operator/pkg/apis/diki/install"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("Controller", func() {
	var (
		ctx = logf.IntoContext(context.Background(), logzap.New(logzap.WriteTo(GinkgoWriter)))

		cr         *scheduledcompliancescan.Reconciler
		fakeClient client.Client
		fakeClock  *testclock.FakeClock

		request reconcile.Request

		scheduledScan *dikiv1alpha1.ScheduledComplianceScan
		scheme        *runtime.Scheme

		baseTime = time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(dikiinstall.AddToScheme(scheme)).To(Succeed())

		fakeClock = testclock.NewFakeClock(baseTime)

		scheduledScan = &dikiv1alpha1.ScheduledComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-scheduled-scan",
				UID:  types.UID("scs-uid-1"),
			},
			Spec: dikiv1alpha1.ScheduledComplianceScanSpec{
				Schedule:                    "0 0 * * 0",
				SuccessfulScansHistoryLimit: ptr.To(int32(3)),
				FailedScansHistoryLimit:     ptr.To(int32(1)),
				ScanTemplate: dikiv1alpha1.ScheduledComplianceScanTemplate{
					Spec: dikiv1alpha1.ComplianceScanSpec{
						Rulesets: []dikiv1alpha1.RulesetConfig{
							{
								ID:      "test-ruleset",
								Version: "v1",
							},
						},
					},
				},
			},
		}

		request = reconcile.Request{NamespacedName: types.NamespacedName{
			Name: scheduledScan.Name,
		}}

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&dikiv1alpha1.ScheduledComplianceScan{}, &dikiv1alpha1.ComplianceScan{}).
			WithObjects(scheduledScan).
			Build()
	})

	JustBeforeEach(func() {
		cr = &scheduledcompliancescan.Reconciler{
			Client: fakeClient,
			Clock:  fakeClock,
		}
	})

	It("should return no error when the ScheduledComplianceScan does not exist", func() {
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&dikiv1alpha1.ScheduledComplianceScan{}).
			Build()
		cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(reconcile.Result{}))
	})

	It("should create a ComplianceScan immediately on first run", func() {
		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(BeNumerically(">", 0))

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(HaveLen(1))

		child := childScans.Items[0]
		Expect(child.Spec.Rulesets).To(HaveLen(1))
		Expect(child.Spec.Rulesets[0].ID).To(Equal("test-ruleset"))
		Expect(child.Spec.Rulesets[0].Version).To(Equal("v1"))

		Expect(child.OwnerReferences).To(HaveLen(1))
		Expect(child.OwnerReferences[0].Name).To(Equal(scheduledScan.Name))
		Expect(child.OwnerReferences[0].UID).To(Equal(scheduledScan.UID))
		Expect(*child.OwnerReferences[0].Controller).To(BeTrue())

		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
		Expect(scheduledScan.Status.Active).NotTo(BeNil())
		Expect(scheduledScan.Status.Active.Name).To(Equal(child.Name))
		Expect(scheduledScan.Status.LastScheduleTime).NotTo(BeNil())
		Expect(scheduledScan.Status.LastScheduleTime.Time).To(BeTemporally("~", baseTime, time.Second))
	})

	It("should truncate the child scan name to respect the DNS label limit", func() {
		longName := strings.Repeat("a", 60)
		scheduledScan.Name = longName
		scheduledScan.UID = types.UID("scs-uid-long")
		request.Name = longName

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&dikiv1alpha1.ScheduledComplianceScan{}).
			WithObjects(scheduledScan).
			Build()
		cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

		_, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(HaveLen(1))
		Expect(len(childScans.Items[0].Name)).To(BeNumerically("<=", 63))
	})

	It("should not create a new ComplianceScan while one is active and requeue", func() {
		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: baseTime.Add(-1 * time.Hour)}
		scheduledScan.Status.Active = &corev1.ObjectReference{Name: "test-scheduled-scan-active"}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

		Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-scheduled-scan-active",
				Labels: map[string]string{
					"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
					"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
				},
			},
			Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanRunning},
		})).To(Succeed())

		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(1 * time.Minute))

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(HaveLen(1))
	})

	It("should clear the active reference and set lastCompletionTime when scan completes", func() {
		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: baseTime.Add(-1 * time.Hour)}
		scheduledScan.Status.Active = &corev1.ObjectReference{Name: "test-scheduled-scan-completed"}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

		Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-scheduled-scan-completed",
				Labels: map[string]string{
					"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
					"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
				},
			},
			Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanCompleted},
		})).To(Succeed())

		_, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
		Expect(scheduledScan.Status.Active).To(BeNil())
		Expect(scheduledScan.Status.LastCompletionTime).NotTo(BeNil())
		Expect(scheduledScan.Status.LastCompletionTime.Time).To(BeTemporally("~", baseTime, time.Second))
	})

	It("should clear the active reference and set lastCompletionTime when scan fails", func() {
		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: baseTime.Add(-1 * time.Hour)}
		scheduledScan.Status.Active = &corev1.ObjectReference{Name: "test-scheduled-scan-failed"}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

		Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-scheduled-scan-failed",
				Labels: map[string]string{
					"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
					"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
				},
			},
			Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanFailed},
		})).To(Succeed())

		_, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
		Expect(scheduledScan.Status.Active).To(BeNil())
		Expect(scheduledScan.Status.LastCompletionTime).NotTo(BeNil())
	})

	It("should recover the active reference for an orphaned ComplianceScan", func() {
		orphanCreationTime := baseTime.Add(-30 * time.Minute)

		Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-scheduled-scan-orphan",
				CreationTimestamp: metav1.Time{Time: orphanCreationTime},
				Labels: map[string]string{
					"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
					"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
				},
			},
			Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanRunning},
		})).To(Succeed())

		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(1 * time.Minute))

		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
		Expect(scheduledScan.Status.Active).NotTo(BeNil())
		Expect(scheduledScan.Status.Active.Name).To(Equal("test-scheduled-scan-orphan"))
		Expect(scheduledScan.Status.LastScheduleTime).NotTo(BeNil())

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(HaveLen(1))
	})

	It("should create a new ComplianceScan when the schedule is due", func() {
		lastSunday := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
		thisSunday := time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC)

		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: lastSunday}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())
		fakeClock = testclock.NewFakeClock(thisSunday)
		cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(BeNumerically(">", 0))

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(HaveLen(1))
	})

	It("should prepend CRON_TZ when timeZone is set and schedule accordingly", func() {
		// Schedule fires at 03:00 Asia/Kolkata (UTC+5:30, no DST).
		// Clock is set to 21:35 UTC the previous day = 03:05, so the schedule is due.
		// Without CRON_TZ the parser would use UTC and 03:00 UTC would not yet be due.
		scheduledScan.Spec.Schedule = "0 3 * * *"
		scheduledScan.Spec.TimeZone = ptr.To("Asia/Kolkata")
		Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())

		// LastScheduleTime is the previous day's 03:00 = 2026-03-29 21:30 UTC.
		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: time.Date(2026, 3, 29, 21, 30, 0, 0, time.UTC)}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

		fakeClock = testclock.NewFakeClock(time.Date(2026, 3, 30, 21, 35, 0, 0, time.UTC))
		cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

		_, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(HaveLen(1))
	})

	It("should not prepend CRON_TZ when timeZone is nil and use parser default", func() {
		// With nil timeZone no CRON_TZ= prefix is added; the parser uses local time.
		// LastScheduleTime is set to 03:00 UTC today; clock is 03:30 UTC same day.
		// Next 03:00 in any timezone is at least 20+ hours away, so no scan should fire.
		// If CRON_TZ= were mistakenly prepended with any ahead-of-UTC timezone,
		// the next firing would still be in the future since we just passed 03:00 UTC.
		scheduledScan.Spec.Schedule = "0 3 * * *"
		Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())

		today0300UTC := time.Date(2026, 3, 30, 3, 0, 0, 0, time.UTC)
		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: today0300UTC}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

		fakeClock = testclock.NewFakeClock(time.Date(2026, 3, 30, 3, 30, 0, 0, time.UTC))
		cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

		_, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(BeEmpty())
	})

	It("should return no error and create no scan when the cron schedule is invalid", func() {
		scheduledScan.Spec.Schedule = "not-a-cron"
		Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())

		_, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(BeEmpty())
	})

	It("should not create a ComplianceScan when the schedule is not yet due and requeue correctly", func() {
		lastSunday := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
		wednesday := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)

		scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: lastSunday}
		Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())
		fakeClock = testclock.NewFakeClock(wednesday)
		cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

		res, err := cr.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		childScans := &dikiv1alpha1.ComplianceScanList{}
		Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
			"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
		})).To(Succeed())
		Expect(childScans.Items).To(BeEmpty())

		// Next Sunday is 2026-03-29 00:00:00 UTC; Wednesday is 2026-03-25 12:00:00 UTC.
		Expect(res.RequeueAfter).To(BeNumerically(">", 3*24*time.Hour))
		Expect(res.RequeueAfter).To(BeNumerically("<", 4*24*time.Hour))
	})

	Context("operation annotation", func() {
		var wednesday = time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
		var lastSunday = time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)

		It("should create a ComplianceScan immediately and remove the annotation when operation=trigger and no active scan", func() {
			scheduledScan.Annotations = map[string]string{
				"scheduledcompliancescan.diki.gardener.cloud/operation": "trigger",
			}
			Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: lastSunday}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())
			fakeClock = testclock.NewFakeClock(wednesday)
			cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

			_, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(1))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
			Expect(scheduledScan.Annotations).NotTo(HaveKey("scheduledcompliancescan.diki.gardener.cloud/operation"))
			Expect(scheduledScan.Status.Active).NotTo(BeNil())
		})

		It("should wait for the active scan to finish and then create a new ComplianceScan when operation=trigger", func() {
			scheduledScan.Annotations = map[string]string{
				"scheduledcompliancescan.diki.gardener.cloud/operation": "trigger",
			}
			Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: lastSunday}
			scheduledScan.Status.Active = &corev1.ObjectReference{Name: "test-scheduled-scan-running"}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())
			fakeClock = testclock.NewFakeClock(wednesday)
			cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

			activeScan := &dikiv1alpha1.ComplianceScan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-scheduled-scan-running",
					Labels: map[string]string{
						"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
						"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
					},
				},
				Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanRunning},
			}
			Expect(fakeClient.Create(ctx, activeScan)).To(Succeed())
			Expect(fakeClient.Status().Update(ctx, activeScan)).To(Succeed())

			// Cycle 1: active scan present, annotation must be preserved and no new scan created.
			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(1 * time.Minute))

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(1))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
			Expect(scheduledScan.Annotations).To(HaveKeyWithValue("scheduledcompliancescan.diki.gardener.cloud/operation", "trigger"))

			// Between cycles: mark the active scan as completed.
			activeScan.Status.Phase = dikiv1alpha1.ComplianceScanCompleted
			Expect(fakeClient.Status().Update(ctx, activeScan)).To(Succeed())

			// Cycle 2: active scan gone, annotation consumed, new scan created.
			_, err = cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(2))

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
			Expect(scheduledScan.Annotations).NotTo(HaveKey("scheduledcompliancescan.diki.gardener.cloud/operation"))
			Expect(scheduledScan.Status.Active).NotTo(BeNil())
		})

		It("should return an error when the annotation patch fails", func() {
			scheduledScan.Annotations = map[string]string{
				"scheduledcompliancescan.diki.gardener.cloud/operation": "trigger",
			}
			Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: lastSunday}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())
			fakeClock = testclock.NewFakeClock(wednesday)

			failingClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&dikiv1alpha1.ScheduledComplianceScan{}, &dikiv1alpha1.ComplianceScan{}).
				WithObjects(scheduledScan).
				WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(_ context.Context, _ client.WithWatch, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
						return errors.New("patch-failed")
					},
				}).Build()
			cr = &scheduledcompliancescan.Reconciler{Client: failingClient, Clock: fakeClock}

			_, err := cr.Reconcile(ctx, request)
			Expect(err).To(MatchError(ContainSubstring("failed to patch ScheduledComplianceScan")))

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(failingClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(1))
		})

		It("should not create a ComplianceScan and not remove the annotation when operation is not trigger", func() {
			scheduledScan.Annotations = map[string]string{
				"scheduledcompliancescan.diki.gardener.cloud/operation": "not-trigger",
			}
			Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: lastSunday}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())
			fakeClock = testclock.NewFakeClock(wednesday)
			cr = &scheduledcompliancescan.Reconciler{Client: fakeClient, Clock: fakeClock}

			res, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(BeNumerically(">", 0))

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(BeEmpty())

			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: scheduledScan.Name}, scheduledScan)).To(Succeed())
			Expect(scheduledScan.Annotations).To(HaveKeyWithValue("scheduledcompliancescan.diki.gardener.cloud/operation", "not-trigger"))
		})
	})

	Context("history cleanup", func() {
		It("should delete the oldest successful scans exceeding the limit", func() {
			scheduledScan.Spec.SuccessfulScansHistoryLimit = ptr.To(int32(2))
			Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: baseTime}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

			for i := range 4 {
				Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
					ObjectMeta: metav1.ObjectMeta{
						Name:              scheduledScan.Name + "-s-" + string(rune('a'+i)),
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Duration(i) * time.Hour)},
						Labels: map[string]string{
							"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
							"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
						},
					},
					Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanCompleted},
				})).To(Succeed())
			}

			_, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(2))

			names := make([]string, len(childScans.Items))
			for i, s := range childScans.Items {
				names[i] = s.Name
			}
			Expect(names).To(ContainElements(
				scheduledScan.Name+"-s-c",
				scheduledScan.Name+"-s-d",
			))
		})

		It("should delete the oldest failed scans exceeding the limit", func() {
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: baseTime}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

			for i := range 3 {
				Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
					ObjectMeta: metav1.ObjectMeta{
						Name:              scheduledScan.Name + "-f-" + string(rune('a'+i)),
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Duration(i) * time.Hour)},
						Labels: map[string]string{
							"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
							"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
						},
					},
					Status: dikiv1alpha1.ComplianceScanStatus{Phase: dikiv1alpha1.ComplianceScanFailed},
				})).To(Succeed())
			}

			_, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(1))
			Expect(childScans.Items[0].Name).To(Equal(scheduledScan.Name + "-f-c"))
		})

		It("should independently clean up successful and failed scans per their limits", func() {
			scheduledScan.Spec.SuccessfulScansHistoryLimit = ptr.To(int32(1))
			scheduledScan.Spec.FailedScansHistoryLimit = ptr.To(int32(1))
			Expect(fakeClient.Update(ctx, scheduledScan)).To(Succeed())
			scheduledScan.Status.LastScheduleTime = &metav1.Time{Time: baseTime}
			Expect(fakeClient.Status().Update(ctx, scheduledScan)).To(Succeed())

			for _, s := range []struct {
				name   string
				phase  dikiv1alpha1.ComplianceScanPhase
				offset time.Duration
			}{
				{"s-old", dikiv1alpha1.ComplianceScanCompleted, 0},
				{"s-new", dikiv1alpha1.ComplianceScanCompleted, 1 * time.Hour},
				{"f-old", dikiv1alpha1.ComplianceScanFailed, 2 * time.Hour},
				{"f-new", dikiv1alpha1.ComplianceScanFailed, 3 * time.Hour},
			} {
				Expect(fakeClient.Create(ctx, &dikiv1alpha1.ComplianceScan{
					ObjectMeta: metav1.ObjectMeta{
						Name:              scheduledScan.Name + "-" + s.name,
						CreationTimestamp: metav1.Time{Time: baseTime.Add(s.offset)},
						Labels: map[string]string{
							"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
							"scheduledcompliancescan.diki.gardener.cloud/uid":  string(scheduledScan.UID),
						},
					},
					Status: dikiv1alpha1.ComplianceScanStatus{Phase: s.phase},
				})).To(Succeed())
			}

			_, err := cr.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			childScans := &dikiv1alpha1.ComplianceScanList{}
			Expect(fakeClient.List(ctx, childScans, client.MatchingLabels{
				"scheduledcompliancescan.diki.gardener.cloud/name": scheduledScan.Name,
			})).To(Succeed())
			Expect(childScans.Items).To(HaveLen(2))

			names := make([]string, len(childScans.Items))
			for i, s := range childScans.Items {
				names[i] = s.Name
			}
			Expect(names).To(ContainElements(
				scheduledScan.Name+"-s-new",
				scheduledScan.Name+"-f-new",
			))
		})
	})
})
