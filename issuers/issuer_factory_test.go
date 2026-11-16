package issuers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	api "github.com/djkormo/adcs-issuer/api/v1"
)

// validCABundle is a small self-signed x509 certificate in PEM format, valid
// for use with x509.CertPool.AppendCertsFromPEM.
const validCABundle = `-----BEGIN CERTIFICATE-----
MIIDBTCCAe2gAwIBAgIUa/bP7XhljzDEUVFQgdmG37jaZMgwDQYJKoZIhvcNAQEL
BQAwEjEQMA4GA1UEAwwHdGVzdC1jYTAeFw0yNjA5MDExNDU5MjhaFw0zNjA4Mjkx
NDU5MjhaMBIxEDAOBgNVBAMMB3Rlc3QtY2EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQClrz0G+g586FsU8Fs3wczE/ax+8S7+bSSJRSPp3gzcd+aRIsuP
keVFkH7AFv+8rTtMriCT6wezANBwcZLx4Y+gEkXqAc8FezqSis7hBWdm08G9GB5a
iwYC7SQRu3LM3ZGtEMfoGUgMVluaUdF4vf3yE5iFZEhhw/fMX2ZUZjusFd3Rcghq
kKruAYfovDXf1LF1tMscTPcoH9kBlQ/bdoaduFT5Pl38T3NcuJ74rRK5b6M39BiQ
W8uZVCZjz3xbNeeElgOs2wTeBe0lwOSrrxK+LAbV9KLpERX1UiyV7YloO5NzPrHq
WfpurYsaa2HbGB94OgUFN7JJ0NdmxVlvJPyRAgMBAAGjUzBRMB0GA1UdDgQWBBQ/
y+8N6SLXOpwxYj1NY+jKY1KbUjAfBgNVHSMEGDAWgBQ/y+8N6SLXOpwxYj1NY+jK
Y1KbUjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQADdlEaUa7N
lZiknVi7XZLxQwISl22UMS7dQ4SpjtMormIQO2+AE4nhvhN64efnRLY0s+i3uBGV
DuHwOLs3ERQzLmxF0nsSGaAD0RpLCLIvz/QAB0ZAiTYm6Xq23WPiBMCpl6b2gn2q
POHB2esDLBeljwu6YLigtBG0E9t85pnIL4RczJnls2SDOYEFEYNDlXxf/qyOLHDY
aAiDi5Hw26PCjbOif20UylK2n09pk/vwd1DhLC/uedhrvwIqytoz3LWTPz2Vz35M
SshNa3AtCoRWw0r1IKzS0HRLzy/1aNDgOep2elZw9LxD0i4oJ9OsX4KkJ3qcDpgz
n+yxpKodWxCd
-----END CERTIFICATE-----
`

func newTestScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	assert.NoError(t, clientgoscheme.AddToScheme(scheme))
	assert.NoError(t, api.AddToScheme(scheme))
	return scheme
}

func newTestFactory(t *testing.T, objs ...client.Object) *IssuerFactory {
	t.Helper()
	scheme := newTestScheme(t)
	builder := fake.NewClientBuilder().WithScheme(scheme)
	if len(objs) > 0 {
		builder = builder.WithObjects(objs...)
	}
	return &IssuerFactory{
		Client:                   builder.Build(),
		ClusterResourceNamespace: "cluster-ns",
		AdcsTemplateName:         "DefaultTemplate",
	}
}

func validSecret(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	}
}

func TestGetIssuerUnsupportedKind(t *testing.T) {
	f := newTestFactory(t)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "Bogus", Name: "foo"}, "default")

	assert.Nil(t, issuer)
	assert.EqualError(t, err, "unsupported issuer kind Bogus")
}

func TestGetIssuerRoutesToAdcsIssuer(t *testing.T) {
	adcsIssuer := &api.AdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-issuer", Namespace: "default"},
		Spec: api.AdcsIssuerSpec{
			URL:            "https://adcs.example.com",
			CredentialsRef: api.LocalObjectReference{Name: "creds"},
			CABundle:       []byte(validCABundle),
			TemplateName:   "WebServer",
		},
	}
	secret := validSecret("creds", "default")
	f := newTestFactory(t, adcsIssuer, secret)

	for _, kind := range []string{"adcsissuer", "AdcsIssuer"} {
		issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: kind, Name: "my-issuer"}, "default")
		assert.NoError(t, err)
		assert.NotNil(t, issuer)
		assert.Equal(t, "WebServer", issuer.AdcsTemplateName)
	}
}

func TestGetIssuerRoutesToClusterAdcsIssuer(t *testing.T) {
	clusterIssuer := &api.ClusterAdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-cluster-issuer"},
		Spec: api.ClusterAdcsIssuerSpec{
			URL:            "https://adcs.example.com",
			CredentialsRef: api.LocalObjectReference{Name: "creds"},
			CABundle:       []byte(validCABundle),
			TemplateName:   "WebServer",
		},
	}
	secret := validSecret("creds", "cluster-ns")
	f := newTestFactory(t, clusterIssuer, secret)

	for _, kind := range []string{"clusteradcsissuer", "ClusterAdcsIssuer"} {
		issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: kind, Name: "my-cluster-issuer"}, "default")
		assert.NoError(t, err)
		assert.NotNil(t, issuer)
		assert.Equal(t, "WebServer", issuer.AdcsTemplateName)
	}
}

func TestGetAdcsIssuerNotFound(t *testing.T) {
	f := newTestFactory(t)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "adcsissuer", Name: "missing"}, "default")

	assert.Nil(t, issuer)
	assert.Error(t, err)
}

func TestGetAdcsIssuerEmptyCABundle(t *testing.T) {
	adcsIssuer := &api.AdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-issuer", Namespace: "default"},
		Spec: api.AdcsIssuerSpec{
			URL:            "https://adcs.example.com",
			CredentialsRef: api.LocalObjectReference{Name: "creds"},
		},
	}
	secret := validSecret("creds", "default")
	f := newTestFactory(t, adcsIssuer, secret)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "adcsissuer", Name: "my-issuer"}, "default")

	assert.Nil(t, issuer)
	assert.EqualError(t, err, "CA Bundle required")
}

func TestGetAdcsIssuerInvalidCABundle(t *testing.T) {
	adcsIssuer := &api.AdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-issuer", Namespace: "default"},
		Spec: api.AdcsIssuerSpec{
			URL:            "https://adcs.example.com",
			CredentialsRef: api.LocalObjectReference{Name: "creds"},
			CABundle:       []byte("not a valid pem certificate"),
		},
	}
	secret := validSecret("creds", "default")
	f := newTestFactory(t, adcsIssuer, secret)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "adcsissuer", Name: "my-issuer"}, "default")

	assert.Nil(t, issuer)
	assert.EqualError(t, err, "error loading ADCS CA bundle")
}

func TestGetClusterAdcsIssuerNotFound(t *testing.T) {
	f := newTestFactory(t)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "clusteradcsissuer", Name: "missing"}, "default")

	assert.Nil(t, issuer)
	assert.Error(t, err)
}

func TestGetClusterAdcsIssuerEmptyCABundle(t *testing.T) {
	clusterIssuer := &api.ClusterAdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-cluster-issuer"},
		Spec: api.ClusterAdcsIssuerSpec{
			URL:            "https://adcs.example.com",
			CredentialsRef: api.LocalObjectReference{Name: "creds"},
		},
	}
	secret := validSecret("creds", "cluster-ns")
	f := newTestFactory(t, clusterIssuer, secret)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "clusteradcsissuer", Name: "my-cluster-issuer"}, "default")

	assert.Nil(t, issuer)
	assert.EqualError(t, err, "CA Bundle required")
}

func TestGetClusterAdcsIssuerInvalidCABundle(t *testing.T) {
	clusterIssuer := &api.ClusterAdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-cluster-issuer"},
		Spec: api.ClusterAdcsIssuerSpec{
			URL:            "https://adcs.example.com",
			CredentialsRef: api.LocalObjectReference{Name: "creds"},
			CABundle:       []byte("not a valid pem certificate"),
		},
	}
	secret := validSecret("creds", "cluster-ns")
	f := newTestFactory(t, clusterIssuer, secret)

	issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: "clusteradcsissuer", Name: "my-cluster-issuer"}, "default")

	assert.Nil(t, issuer)
	assert.EqualError(t, err, "error loading ADCS CA bundle")
}

func TestGetUserPasswordSecretNotFound(t *testing.T) {
	f := newTestFactory(t)

	username, password, realm, err := f.getUserPassword(context.Background(), "missing", "default")

	assert.Empty(t, username)
	assert.Empty(t, password)
	assert.Empty(t, realm)
	assert.Error(t, err)
}

func TestGetUserPasswordMissingUsername(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"},
		Data: map[string][]byte{
			"password": []byte("pass"),
		},
	}
	f := newTestFactory(t, secret)

	_, _, _, err := f.getUserPassword(context.Background(), "creds", "default")

	assert.EqualError(t, err, "user name not set in secret")
}

func TestGetUserPasswordMissingPassword(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"},
		Data: map[string][]byte{
			"username": []byte("user"),
		},
	}
	f := newTestFactory(t, secret)

	_, _, _, err := f.getUserPassword(context.Background(), "creds", "default")

	assert.EqualError(t, err, "password not set in secret")
}

func TestGetUserPasswordMissingRealmWhenKerberos(t *testing.T) {
	t.Setenv("ADCS_AUTH_MODE", "kerberos")
	secret := validSecret("creds", "default")
	f := newTestFactory(t, secret)

	_, _, _, err := f.getUserPassword(context.Background(), "creds", "default")

	assert.EqualError(t, err, "realm not set in secret")
}

func TestGetUserPasswordSuccess(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
			"realm":    []byte("EXAMPLE.COM"),
		},
	}
	f := newTestFactory(t, secret)

	username, password, realm, err := f.getUserPassword(context.Background(), "creds", "default")

	assert.NoError(t, err)
	assert.Equal(t, "user", username)
	assert.Equal(t, "pass", password)
	assert.Equal(t, "EXAMPLE.COM", realm)
}

func TestGetIntervalUsesDefaultWhenEmpty(t *testing.T) {
	result := getInterval("", "6h", logr.Discard())
	assert.Equal(t, 6*time.Hour, result)
}

func TestGetIntervalUsesSpecValueWhenValid(t *testing.T) {
	result := getInterval("30m", "6h", logr.Discard())
	assert.Equal(t, 30*time.Minute, result)
}

func TestGetIntervalFallsBackToDefaultWhenInvalid(t *testing.T) {
	result := getInterval("not-a-duration", "1h", logr.Discard())
	assert.Equal(t, time.Hour, result)
}

func TestGetCaBundle(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "trust", Namespace: "issuer-ns"},
		Data: map[string][]byte{
			"ca.crt":     []byte("default bundle"),
			"bundle.pem": []byte("custom bundle"),
			"empty":      {},
		},
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "trust", Namespace: "issuer-ns"},
		Data: map[string]string{
			"ca.crt":     "configmap bundle",
			"bundle.pem": "custom configmap bundle",
		},
		BinaryData: map[string][]byte{"binary.pem": []byte("binary bundle")},
	}
	tests := []struct {
		name      string
		ref       api.CABundleReference
		namespace string
		want      string
		wantErr   string
	}{
		{name: "no reference", namespace: "issuer-ns"},
		{name: "name only", ref: api.CABundleReference{Name: "trust"}, namespace: "issuer-ns", want: "default bundle"},
		{name: "custom secret key", ref: api.CABundleReference{Name: "trust", Kind: "Secret", Key: "bundle.pem"}, namespace: "issuer-ns", want: "custom bundle"},
		{name: "configmap default key", ref: api.CABundleReference{Name: "trust", Kind: "ConfigMap"}, namespace: "issuer-ns", want: "configmap bundle"},
		{name: "custom configmap key", ref: api.CABundleReference{Name: "trust", Kind: "ConfigMap", Key: "bundle.pem"}, namespace: "issuer-ns", want: "custom configmap bundle"},
		{name: "configmap binary data", ref: api.CABundleReference{Name: "trust", Kind: "ConfigMap", Key: "binary.pem"}, namespace: "issuer-ns", want: "binary bundle"},
		{name: "missing secret key", ref: api.CABundleReference{Name: "trust", Key: "missing"}, namespace: "issuer-ns", wantErr: `key "missing" is missing or empty in Secret issuer-ns/trust`},
		{name: "missing configmap key", ref: api.CABundleReference{Name: "trust", Kind: "ConfigMap", Key: "missing"}, namespace: "issuer-ns", wantErr: `key "missing" is missing or empty in ConfigMap issuer-ns/trust`},
		{name: "empty key", ref: api.CABundleReference{Name: "trust", Key: "empty"}, namespace: "issuer-ns", wantErr: `key "empty" is missing or empty`},
		{name: "namespace isolation", ref: api.CABundleReference{Name: "trust"}, namespace: "other-ns", wantErr: `Secret other-ns/trust`},
		{name: "missing configmap", ref: api.CABundleReference{Name: "missing", Kind: "ConfigMap"}, namespace: "issuer-ns", wantErr: `ConfigMap issuer-ns/missing`},
		{name: "missing name", ref: api.CABundleReference{Kind: "Secret"}, namespace: "issuer-ns", wantErr: "caBundleRef.name is required"},
		{name: "unsupported kind", ref: api.CABundleReference{Name: "trust", Kind: "Pod"}, namespace: "issuer-ns", wantErr: `unsupported caBundleRef kind "Pod"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFactory(t, secret, configMap)
			bundle, err := f.getCaBundle(context.Background(), tt.ref, tt.namespace)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Empty(t, bundle)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(bundle))
		})
	}
}

func TestGetIssuerCABundleRef(t *testing.T) {
	t.Setenv("ADCS_AUTH_MODE", "ntlm")
	for _, issuerKind := range []string{"AdcsIssuer", "ClusterAdcsIssuer"} {
		for _, sourceKind := range []string{"Secret", "ConfigMap"} {
			namespaces := []string{""}
			if issuerKind == "ClusterAdcsIssuer" {
				namespaces = append(namespaces, "shared-trust")
			}
			for _, override := range namespaces {
				t.Run(issuerKind+"/"+sourceKind+"/"+override, func(t *testing.T) {
					namespace := "issuer-ns"
					if issuerKind == "ClusterAdcsIssuer" {
						namespace = "cluster-ns"
					}
					caNamespace := namespace
					if override != "" {
						caNamespace = override
					}
					ref := api.CABundleReference{Name: "trust", Kind: sourceKind, Key: "bundle.pem"}
					var issuerObject client.Object
					if issuerKind == "AdcsIssuer" {
						issuerObject = &api.AdcsIssuer{
							ObjectMeta: metav1.ObjectMeta{Name: "issuer", Namespace: namespace},
							Spec: api.AdcsIssuerSpec{
								URL:            "https://adcs.example.com",
								CredentialsRef: api.LocalObjectReference{Name: "creds"},
								CABundleRef:    ref,
								CABundle:       []byte("invalid inline bundle must not be used"),
							},
						}
					} else {
						issuerObject = &api.ClusterAdcsIssuer{
							ObjectMeta: metav1.ObjectMeta{Name: "issuer"},
							Spec: api.ClusterAdcsIssuerSpec{
								URL:            "https://adcs.example.com",
								CredentialsRef: api.LocalObjectReference{Name: "creds"},
								CABundleRef: api.NamespacedCABundleReference{
									CABundleReference: ref,
									Namespace:         override,
								},
								CABundle: []byte("invalid inline bundle must not be used"),
							},
						}
					}
					var source client.Object
					if sourceKind == "Secret" {
						source = &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{Name: ref.Name, Namespace: caNamespace},
							Data:       map[string][]byte{ref.Key: []byte(validCABundle)},
						}
					} else {
						source = &corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{Name: ref.Name, Namespace: caNamespace},
							Data:       map[string]string{ref.Key: validCABundle},
						}
					}
					f := newTestFactory(t, issuerObject, source, validSecret("creds", namespace))
					issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: issuerKind, Name: "issuer"}, "issuer-ns")
					require.NoError(t, err)
					assert.NotNil(t, issuer)

					require.NoError(t, f.Delete(context.Background(), source))
					issuer, err = f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: issuerKind, Name: "issuer"}, "issuer-ns")
					require.ErrorContains(t, err, "error loading ADCS CA bundle")
					assert.Nil(t, issuer)
				})
			}
		}
	}
}

func TestGetIssuerCABundleRefLookupError(t *testing.T) {
	t.Setenv("ADCS_AUTH_MODE", "ntlm")
	for _, issuerKind := range []string{"AdcsIssuer", "ClusterAdcsIssuer"} {
		for _, inline := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/inline=%t", issuerKind, inline), func(t *testing.T) {
				var bundle []byte
				if inline {
					bundle = []byte(validCABundle)
				}
				ref := api.CABundleReference{Name: "missing"}
				var issuerObject client.Object
				namespace := "issuer-ns"
				if issuerKind == "AdcsIssuer" {
					issuerObject = &api.AdcsIssuer{
						ObjectMeta: metav1.ObjectMeta{Name: "issuer", Namespace: namespace},
						Spec: api.AdcsIssuerSpec{
							URL:            "https://adcs.example.com",
							CredentialsRef: api.LocalObjectReference{Name: "creds"},
							CABundleRef:    ref,
							CABundle:       bundle,
						},
					}
				} else {
					namespace = "cluster-ns"
					issuerObject = &api.ClusterAdcsIssuer{
						ObjectMeta: metav1.ObjectMeta{Name: "issuer"},
						Spec: api.ClusterAdcsIssuerSpec{
							URL:            "https://adcs.example.com",
							CredentialsRef: api.LocalObjectReference{Name: "creds"},
							CABundleRef:    api.NamespacedCABundleReference{CABundleReference: ref},
							CABundle:       bundle,
						},
					}
				}
				f := newTestFactory(t, issuerObject, validSecret("creds", namespace))
				issuer, err := f.GetIssuer(context.Background(), cmmeta.ObjectReference{Kind: issuerKind, Name: "issuer"}, "issuer-ns")
				if inline {
					require.NoError(t, err)
					assert.NotNil(t, issuer)
				} else {
					require.ErrorContains(t, err, "loading caBundleRef Secret "+namespace+"/missing")
					assert.Nil(t, issuer)
				}
			})
		}
	}
}
