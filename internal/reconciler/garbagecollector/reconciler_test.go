// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler_test

import (
	"context"
	"errors"
	"time"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	garbagecollector "github.com/gardener/diki-operator/internal/reconciler/garbagecollector"
	dikiinstall "github.com/gardener/diki-operator/pkg/apis/diki/install"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

var _ = Describe("Controller", func() {
	var (
		ctx = logf.IntoContext(context.Background(), logzap.New(logzap.WriteTo(GinkgoWriter)))

		cr         *garbagecollector.Reconciler
		fakeClient client.Client
		scheme     *runtime.Scheme
		scan       *dikiv1alpha1.ComplianceScan

		jobNamespace = "kube-system"
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(kubernetes.AddGardenSchemeToScheme(scheme)).To(Succeed())
		Expect(dikiinstall.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&dikiv1alpha1.ComplianceScan{}).Build()

		cr = &garbagecollector.Reconciler{
			Client:       fakeClient,
			SourceClient: fakeClient,
			Config: garbagecollector.Config{
				Namespace:       jobNamespace,
				RequeueInterval: 1 * time.Minute,
			},
		}

		scan = &dikiv1alpha1.ComplianceScan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scan",
				UID:  types.UID("scan-uid"),
			},
		}
	})

	It("should requeue after interval when no Jobs exist", func() {
		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))
	})

	It("should not delete Job when ComplianceScan is Pending", func() {
		scan.Status.Phase = dikiv1alpha1.ComplianceScanPending
		Expect(fakeClient.Create(ctx, scan)).To(Succeed())
		Expect(fakeClient.Status().Update(ctx, scan)).To(Succeed())

		job := newDikiRunJob("diki-run-scan-uid", jobNamespace, "scan-uid")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

		Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
	})

	It("should not delete Job when ComplianceScan is still running", func() {
		scan.Status.Phase = dikiv1alpha1.ComplianceScanRunning
		Expect(fakeClient.Create(ctx, scan)).To(Succeed())
		Expect(fakeClient.Status().Update(ctx, scan)).To(Succeed())

		job := newDikiRunJob("diki-run-scan-uid", jobNamespace, "scan-uid")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

		Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
	})

	It("should delete Job when ComplianceScan is Completed", func() {
		scan.Status.Phase = dikiv1alpha1.ComplianceScanCompleted
		Expect(fakeClient.Create(ctx, scan)).To(Succeed())
		Expect(fakeClient.Status().Update(ctx, scan)).To(Succeed())

		job := newDikiRunJob("diki-run-scan-uid", jobNamespace, "scan-uid")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

		err = fakeClient.Get(ctx, client.ObjectKeyFromObject(job), job)
		Expect(err).To(HaveOccurred())
		Expect(client.IgnoreNotFound(err)).To(Succeed())
	})

	It("should delete Job when ComplianceScan is Failed", func() {
		scan.Status.Phase = dikiv1alpha1.ComplianceScanFailed
		Expect(fakeClient.Create(ctx, scan)).To(Succeed())
		Expect(fakeClient.Status().Update(ctx, scan)).To(Succeed())

		job := newDikiRunJob("diki-run-scan-uid", jobNamespace, "scan-uid")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

		err = fakeClient.Get(ctx, client.ObjectKeyFromObject(job), job)
		Expect(err).To(HaveOccurred())
		Expect(client.IgnoreNotFound(err)).To(Succeed())
	})

	It("should delete Job when ComplianceScan does not exist (orphaned)", func() {
		job := newDikiRunJob("diki-run-uid-orphan", jobNamespace, "uid-orphan")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

		err = fakeClient.Get(ctx, client.ObjectKeyFromObject(job), job)
		Expect(err).To(HaveOccurred())
		Expect(client.IgnoreNotFound(err)).To(Succeed())
	})

	It("should not delete Job that is missing the ComplianceScan UID label", func() {
		job := newDikiRunJob("unrelated-job", jobNamespace, "uid-unlabeled")
		delete(job.Labels, "compliancescan.diki.gardener.cloud/uid")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

		Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
	})

	Context("when source and target clusters are different", func() {
		var sourceClient client.Client

		BeforeEach(func() {
			sourceClient = fake.NewClientBuilder().WithScheme(scheme).Build()
			cr.SourceClient = sourceClient
		})

		It("should delete Job from source cluster when ComplianceScan on target cluster is Completed", func() {
			scan.Status.Phase = dikiv1alpha1.ComplianceScanCompleted
			Expect(fakeClient.Create(ctx, scan)).To(Succeed())
			Expect(fakeClient.Status().Update(ctx, scan)).To(Succeed())

			job := newDikiRunJob("diki-run-scan-uid", jobNamespace, "scan-uid")
			Expect(sourceClient.Create(ctx, job)).To(Succeed())

			res, err := cr.Reconcile(ctx, reconcile.Request{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

			err = sourceClient.Get(ctx, client.ObjectKeyFromObject(job), job)
			Expect(err).To(HaveOccurred())
			Expect(client.IgnoreNotFound(err)).To(Succeed())
		})

		It("should delete Job from source cluster when ComplianceScan does not exist on target cluster", func() {
			job := newDikiRunJob("diki-run-orphan-uid", jobNamespace, "orphan-uid")
			Expect(sourceClient.Create(ctx, job)).To(Succeed())

			res, err := cr.Reconcile(ctx, reconcile.Request{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(cr.Config.RequeueInterval))

			err = sourceClient.Get(ctx, client.ObjectKeyFromObject(job), job)
			Expect(err).To(HaveOccurred())
			Expect(client.IgnoreNotFound(err)).To(Succeed())
		})
	})

	It("should return error when listing Jobs fails", func() {
		cr.SourceClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithInterceptorFuncs(interceptor.Funcs{
				List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
					if _, ok := list.(*batchv1.JobList); ok {
						return errors.New("list-jobs-failed")
					}
					return c.List(ctx, list, opts...)
				},
			}).Build()

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).To(MatchError(ContainSubstring("list-jobs-failed")))
		Expect(res).To(Equal(reconcile.Result{}))
	})

	It("should return error when deleting a Job fails", func() {
		scan.Status.Phase = dikiv1alpha1.ComplianceScanCompleted
		Expect(fakeClient.Create(ctx, scan)).To(Succeed())
		Expect(fakeClient.Status().Update(ctx, scan)).To(Succeed())

		job := newDikiRunJob("diki-run-scan-uid", jobNamespace, "scan-uid")
		Expect(fakeClient.Create(ctx, job)).To(Succeed())

		cr.SourceClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(job).
			WithInterceptorFuncs(interceptor.Funcs{
				Delete: func(_ context.Context, _ client.WithWatch, _ client.Object, _ ...client.DeleteOption) error {
					return errors.New("delete-failed")
				},
			}).Build()

		res, err := cr.Reconcile(ctx, reconcile.Request{})
		Expect(err).To(MatchError(ContainSubstring("delete-failed")))
		Expect(res).To(Equal(reconcile.Result{}))
	})
})

func newDikiRunJob(name, namespace, complianceScanUID string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"compliancescan.diki.gardener.cloud/uid": complianceScanUID,
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "diki-scan", Image: "diki:latest"},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}
