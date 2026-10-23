package controllers

import (
	"context"
	"testing"
	"time"

	cmapiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	clocktesting "k8s.io/utils/clock/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/djkormo/adcs-issuer/api/v1"
)

func certificateRequestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := api.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := cmapi.AddToScheme(scheme); err != nil {
		panic(err)
	}
	return scheme
}

func newCertificateRequestReconciler(objs ...client.Object) (*CertificateRequestReconciler, client.Client) {
	fakeClient := fake.NewClientBuilder().
		WithScheme(certificateRequestScheme()).
		WithStatusSubresource(&cmapi.CertificateRequest{}, &api.AdcsRequest{}).
		WithObjects(objs...).
		Build()

	r := &CertificateRequestReconciler{
		Client:                 fakeClient,
		Recorder:               record.NewFakeRecorder(10),
		Clock:                  clocktesting.NewFakeClock(time.Now()),
		CheckApprovedCondition: false,
	}
	return r, fakeClient
}

func baseCertificateRequest(name, namespace string) *cmapi.CertificateRequest {
	return &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: cmapi.CertificateRequestSpec{
			Request: []byte("test-csr"),
			IssuerRef: cmmeta.ObjectReference{
				Name:  "my-issuer",
				Kind:  "AdcsIssuer",
				Group: api.GroupVersion.Group,
			},
		},
	}
}

func TestCertificateRequestReconcile_NotFound(t *testing.T) {
	r, _ := newCertificateRequestReconciler()

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_IssuerGroupMismatch(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	cr.Spec.IssuerRef.Group = "some.other.group"
	r, _ := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_AlreadyReady(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	cr.Status.Conditions = []cmapi.CertificateRequestCondition{
		{Type: cmapi.CertificateRequestConditionReady, Status: cmmeta.ConditionTrue},
	}
	r, _ := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_AlreadyFailed(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	cr.Status.Conditions = []cmapi.CertificateRequestCondition{
		{Type: cmapi.CertificateRequestConditionReady, Status: cmmeta.ConditionFalse, Reason: cmapi.CertificateRequestReasonFailed},
	}
	r, _ := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_AlreadyDenied(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	cr.Status.Conditions = []cmapi.CertificateRequestCondition{
		{Type: cmapi.CertificateRequestConditionReady, Status: cmmeta.ConditionFalse, Reason: cmapi.CertificateRequestReasonDenied},
	}
	r, _ := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_Denied(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	cr.Status.Conditions = []cmapi.CertificateRequestCondition{
		{Type: cmapi.CertificateRequestConditionDenied, Status: cmmeta.ConditionTrue},
	}
	r, fakeClient := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	updated := &cmapi.CertificateRequest{}
	assert.NoError(t, fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "cr1"}, updated))
	assert.NotNil(t, updated.Status.FailureTime)
	assert.True(t, cmapiutil.CertificateRequestHasCondition(updated, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionFalse,
		Reason: cmapi.CertificateRequestReasonDenied,
	}))
}

func TestCertificateRequestReconcile_NotApproved(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	r, _ := newCertificateRequestReconciler(cr)
	r.CheckApprovedCondition = true

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_AlreadyHasCertificate(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	cr.Status.Certificate = []byte("existing-cert")
	r, _ := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestCertificateRequestReconcile_CreatesNewAdcsRequest(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	r, fakeClient := newCertificateRequestReconciler(cr)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	adcsReq := &api.AdcsRequest{}
	assert.NoError(t, fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "cr1"}, adcsReq))
	assert.Equal(t, cr.Spec.Request, adcsReq.Spec.CSRPEM)
	assert.Equal(t, cr.Spec.IssuerRef, adcsReq.Spec.IssuerRef)
	assert.Len(t, adcsReq.OwnerReferences, 1)
	assert.Equal(t, "cr1", adcsReq.OwnerReferences[0].Name)

	updatedCR := &cmapi.CertificateRequest{}
	assert.NoError(t, fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "cr1"}, updatedCR))
	assert.True(t, cmapiutil.CertificateRequestHasCondition(updatedCR, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionFalse,
		Reason: cmapi.CertificateRequestReasonPending,
	}))
}

func TestCertificateRequestReconcile_ExistingAdcsRequestSameCSR(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	adcsReq := &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "cr1", Namespace: "default"},
		Spec:       api.AdcsRequestSpec{CSRPEM: cr.Spec.Request, IssuerRef: cr.Spec.IssuerRef},
	}
	r, fakeClient := newCertificateRequestReconciler(cr, adcsReq)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	list := &api.AdcsRequestList{}
	assert.NoError(t, fakeClient.List(context.Background(), list))
	assert.Len(t, list.Items, 1)
}

func TestCertificateRequestReconcile_ExistingAdcsRequestDifferentCSR(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	adcsReq := &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "cr1", Namespace: "default"},
		Spec:       api.AdcsRequestSpec{CSRPEM: []byte("different-csr"), IssuerRef: cr.Spec.IssuerRef},
	}
	r, fakeClient := newCertificateRequestReconciler(cr, adcsReq)

	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "cr1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	newReq := &api.AdcsRequest{}
	assert.NoError(t, fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "cr1"}, newReq))
	assert.Equal(t, cr.Spec.Request, newReq.Spec.CSRPEM)
}

func TestRequestDiffers(t *testing.T) {
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{"identical", []byte("same-content"), []byte("same-content"), false},
		{"different-length", []byte("short"), []byte("a-longer-value"), true},
		{"equal-length-different-content", []byte("aaaaa"), []byte("bbbbb"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adcsReq := &api.AdcsRequest{Spec: api.AdcsRequestSpec{CSRPEM: tc.a}}
			certReq := &cmapi.CertificateRequest{Spec: cmapi.CertificateRequestSpec{Request: tc.b}}
			assert.Equal(t, tc.expected, RequestDiffers(adcsReq, certReq))
		})
	}
}

func TestGetCertificateRequest(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	r, _ := newCertificateRequestReconciler(cr)

	got, err := r.GetCertificateRequest(context.Background(), types.NamespacedName{Namespace: "default", Name: "cr1"})
	assert.NoError(t, err)
	assert.Equal(t, "cr1", got.Name)

	_, err = r.GetCertificateRequest(context.Background(), types.NamespacedName{Namespace: "default", Name: "missing"})
	assert.Error(t, err)
}

func TestSetStatus(t *testing.T) {
	cr := baseCertificateRequest("cr1", "default")
	r, fakeClient := newCertificateRequestReconciler(cr)

	err := r.SetStatus(context.Background(), cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "issued %s", "successfully")
	assert.NoError(t, err)

	updated := &cmapi.CertificateRequest{}
	assert.NoError(t, fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "cr1"}, updated))
	assert.True(t, cmapiutil.CertificateRequestHasCondition(updated, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionTrue,
		Reason: cmapi.CertificateRequestReasonIssued,
	}))
}
