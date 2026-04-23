// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler_test

import (
	"context"
	"strings"
	"time"

	"github.com/gardener/gardener/pkg/client/kubernetes"
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
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
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
