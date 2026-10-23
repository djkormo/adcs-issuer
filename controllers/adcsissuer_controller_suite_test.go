package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/djkormo/adcs-issuer/issuers"
)

// newTestManager creates a manager against the envtest API server with the
// metrics server disabled, so that multiple managers can be created across
// specs without port conflicts.
func newTestManager() (ctrl.Manager, error) {
	return ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
}

var _ = Describe("SetupWithManager", func() {
	It("wires up the AdcsIssuerReconciler without error", func() {
		mgr, err := newTestManager()
		Expect(err).NotTo(HaveOccurred())

		err = (&AdcsIssuerReconciler{
			Client: mgr.GetClient(),
			Log:    logr.Discard(),
		}).SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())
	})

	It("wires up the ClusterAdcsIssuerReconciler without error", func() {
		mgr, err := newTestManager()
		Expect(err).NotTo(HaveOccurred())

		err = (&ClusterAdcsIssuerReconciler{
			Client: mgr.GetClient(),
			Log:    logr.Discard(),
		}).SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())
	})

	It("wires up the CertificateRequestReconciler without error", func() {
		mgr, err := newTestManager()
		Expect(err).NotTo(HaveOccurred())

		certReconciler := &CertificateRequestReconciler{
			Client:   mgr.GetClient(),
			Recorder: mgr.GetEventRecorderFor("adcs-certificaterequests-controller"),
			Clock:    clock.RealClock{},
		}
		err = certReconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())
	})

	It("wires up the AdcsRequestReconciler without error", func() {
		mgr, err := newTestManager()
		Expect(err).NotTo(HaveOccurred())

		certReconciler := &CertificateRequestReconciler{
			Client:   mgr.GetClient(),
			Recorder: mgr.GetEventRecorderFor("adcs-certificaterequests-controller"),
			Clock:    clock.RealClock{},
		}

		err = (&AdcsRequestReconciler{
			Client: mgr.GetClient(),
			Log:    logr.Discard(),
			IssuerFactory: issuers.IssuerFactory{
				Client:           mgr.GetClient(),
				AdcsTemplateName: "test",
			},
			Recorder:                     mgr.GetEventRecorderFor("adcs-requests-controller"),
			CertificateRequestController: certReconciler,
		}).SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())
	})
})
