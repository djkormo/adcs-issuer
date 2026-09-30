package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	"github.com/djkormo/adcs-issuer/issuers"
)

const (
	pkcs7TestdataPath = "../issuers/testdata/cfss_rawPKCS7.p7b"
	x509TestdataPath  = "../issuers/testdata/cfss_outputx509.pem"
)

func adcsRequestScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	require.NoError(t, api.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, cmapi.AddToScheme(scheme))
	return scheme
}

// adcsBackendConfig controls how the fake ADCS HTTP backend behaves for a
// given test case.
type adcsBackendConfig struct {
	// certfnshBody/ContentType controls the response to the initial
	// certificate request submission (POST certfnsh.asp).
	certfnshBody        string
	certfnshContentType string
}

func newAdcsBackend(t *testing.T, cfg adcsBackendConfig) *httptest.Server {
	pkcs7Body, err := os.ReadFile(pkcs7TestdataPath)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/certfnsh.asp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", cfg.certfnshContentType)
		_, _ = w.Write([]byte(cfg.certfnshBody))
	})
	mux.HandleFunc("/certnew.cer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(cfg.certfnshBody))
	})
	mux.HandleFunc("/certcarc.asp", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("var nRenewals=0;"))
	})
	mux.HandleFunc("/certnew.p7b", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-pkcs7-certificates")
		_, _ = w.Write(pkcs7Body)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// adcsTestFixture bundles a fake client plus wired reconcilers for the
// AdcsRequest/CertificateRequest controller pair.
type adcsTestFixture struct {
	client                client.Client
	adcsRequestReconciler *AdcsRequestReconciler
	certReconciler        *CertificateRequestReconciler
}

func newAdcsRequestFixture(t *testing.T, server *httptest.Server, objs ...client.Object) *adcsTestFixture {
	caCertPEM, err := os.ReadFile(x509TestdataPath)
	require.NoError(t, err)

	issuerObj := &api.AdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-issuer", Namespace: "default"},
		Spec: api.AdcsIssuerSpec{
			URL:                 server.URL,
			CredentialsRef:      api.LocalObjectReference{Name: "my-secret"},
			CABundle:            caCertPEM,
			StatusCheckInterval: "1s",
			RetryInterval:       "1s",
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-secret", Namespace: "default"},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	}

	allObjs := append([]client.Object{issuerObj, secret}, objs...)

	fakeClient := fake.NewClientBuilder().
		WithScheme(adcsRequestScheme(t)).
		WithStatusSubresource(&api.AdcsRequest{}, &cmapi.CertificateRequest{}).
		WithObjects(allObjs...).
		Build()

	certReconciler := &CertificateRequestReconciler{
		Client:                 fakeClient,
		Recorder:               record.NewFakeRecorder(10),
		Clock:                  clocktesting.NewFakeClock(time.Now()),
		CheckApprovedCondition: false,
	}

	adcsReconciler := &AdcsRequestReconciler{
		Client: fakeClient,
		Log:    logr.Discard(),
		IssuerFactory: issuers.IssuerFactory{
			Client:           fakeClient,
			AdcsTemplateName: "test",
		},
		Recorder:                     record.NewFakeRecorder(10),
		CertificateRequestController: certReconciler,
	}

	return &adcsTestFixture{
		client:                fakeClient,
		adcsRequestReconciler: adcsReconciler,
		certReconciler:        certReconciler,
	}
}

func issuerRefFor(name string) cmmeta.ObjectReference {
	return cmmeta.ObjectReference{
		Name:  name,
		Kind:  "AdcsIssuer",
		Group: api.GroupVersion.Group,
	}
}

func TestAdcsRequestReconcile_NotFound(t *testing.T) {
	server := newAdcsBackend(t, adcsBackendConfig{})
	fixture := newAdcsRequestFixture(t, server)

	res, err := fixture.adcsRequestReconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestAdcsRequestReconcile_GetIssuerError(t *testing.T) {
	server := newAdcsBackend(t, adcsBackendConfig{})
	ar := &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: api.AdcsRequestSpec{
			CSRPEM:    []byte("csr-bytes"),
			IssuerRef: cmmeta.ObjectReference{Name: "my-issuer", Kind: "unsupported", Group: api.GroupVersion.Group},
		},
	}
	fixture := newAdcsRequestFixture(t, server, ar)

	res, err := fixture.adcsRequestReconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "req1"},
	})

	assert.Error(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestAdcsRequestReconcile_UnknownToReady(t *testing.T) {
	server := newAdcsBackend(t, adcsBackendConfig{
		certfnshBody:        "dummy-cert-bytes",
		certfnshContentType: "application/pkix-cert",
	})

	ar := &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: api.AdcsRequestSpec{
			CSRPEM:    []byte("csr-bytes"),
			IssuerRef: issuerRefFor("my-issuer"),
		},
	}
	cr := &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: cmapi.CertificateRequestSpec{
			Request:   []byte("csr-bytes"),
			IssuerRef: issuerRefFor("my-issuer"),
		},
	}
	fixture := newAdcsRequestFixture(t, server, ar, cr)

	res, err := fixture.adcsRequestReconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "req1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	updatedAR := &api.AdcsRequest{}
	require.NoError(t, fixture.client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "req1"}, updatedAR))
	assert.Equal(t, api.Ready, updatedAR.Status.State)

	updatedCR := &cmapi.CertificateRequest{}
	require.NoError(t, fixture.client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "req1"}, updatedCR))
	assert.NotEmpty(t, updatedCR.Status.Certificate)
	assert.Contains(t, string(updatedCR.Status.Certificate), "dummy-cert-bytes")
}

func TestAdcsRequestReconcile_Pending(t *testing.T) {
	pendingBody := "Certificate Pending\r\n\tDisposition message: \t\tTaken Under Submission\r\n"
	server := newAdcsBackend(t, adcsBackendConfig{
		certfnshBody: pendingBody,
	})

	ar := &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: api.AdcsRequestSpec{
			CSRPEM:    []byte("csr-bytes"),
			IssuerRef: issuerRefFor("my-issuer"),
		},
		Status: api.AdcsRequestStatus{
			State: api.Pending,
			Id:    "123",
		},
	}
	cr := &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: cmapi.CertificateRequestSpec{
			Request:   []byte("csr-bytes"),
			IssuerRef: issuerRefFor("my-issuer"),
		},
	}
	fixture := newAdcsRequestFixture(t, server, ar, cr)

	res, err := fixture.adcsRequestReconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "req1"},
	})

	assert.NoError(t, err)
	assert.True(t, res.Requeue)
	assert.Equal(t, time.Second, res.RequeueAfter)

	updatedAR := &api.AdcsRequest{}
	require.NoError(t, fixture.client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "req1"}, updatedAR))
	assert.Equal(t, api.Pending, updatedAR.Status.State)
}

func testTerminalState(t *testing.T, state api.State, expectedReason string, expectedMessage string) {
	server := newAdcsBackend(t, adcsBackendConfig{})

	ar := &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: api.AdcsRequestSpec{
			CSRPEM:    []byte("csr-bytes"),
			IssuerRef: issuerRefFor("my-issuer"),
		},
		Status: api.AdcsRequestStatus{
			State: state,
		},
	}
	cr := &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "req1", Namespace: "default"},
		Spec: cmapi.CertificateRequestSpec{
			Request:   []byte("csr-bytes"),
			IssuerRef: issuerRefFor("my-issuer"),
		},
	}
	fixture := newAdcsRequestFixture(t, server, ar, cr)

	res, err := fixture.adcsRequestReconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "req1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	updatedCR := &cmapi.CertificateRequest{}
	require.NoError(t, fixture.client.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "req1"}, updatedCR))

	found := false
	for _, cond := range updatedCR.Status.Conditions {
		if cond.Type == cmapi.CertificateRequestConditionReady && cond.Reason == expectedReason {
			found = true
			assert.Equal(t, expectedMessage, cond.Message)
		}
	}
	assert.True(t, found, fmt.Sprintf("expected condition with reason %s not found", expectedReason))
}

func TestAdcsRequestReconcile_Rejected(t *testing.T) {
	testTerminalState(t, api.Rejected, cmapi.CertificateRequestReasonPending, "ADCS request rejected")
}

func TestAdcsRequestReconcile_Errored(t *testing.T) {
	testTerminalState(t, api.Errored, cmapi.CertificateRequestReasonFailed, "ADCS request errored")
}
