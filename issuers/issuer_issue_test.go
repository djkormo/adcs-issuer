package issuers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/djkormo/adcs-issuer/adcs"
	api "github.com/djkormo/adcs-issuer/api/v1"
)

// validSelfSignedCert is a small self-signed x509 certificate in PEM format,
// used as a stand-in CA chain returned by the fake AdcsCertsrv.
const validSelfSignedCert = `-----BEGIN CERTIFICATE-----
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

// fakeCertsrv is a hand-written test double implementing adcs.AdcsCertsrv,
// allowing full control of return values and error injection per test case.
type fakeCertsrv struct {
	requestCertificateCalls     int
	getExistingCertificateCalls int

	requestStatus adcs.AdcsResponseStatus
	requestDesc   string
	requestID     string
	requestErr    error

	existingStatus adcs.AdcsResponseStatus
	existingDesc   string
	existingID     string
	existingErr    error

	caChain    string
	caChainErr error

	caCert    string
	caCertErr error
}

func (f *fakeCertsrv) RequestCertificate(csr string, template string) (adcs.AdcsResponseStatus, string, string, error) {
	f.requestCertificateCalls++
	return f.requestStatus, f.requestDesc, f.requestID, f.requestErr
}

func (f *fakeCertsrv) GetExistingCertificate(id string) (adcs.AdcsResponseStatus, string, string, error) {
	f.getExistingCertificateCalls++
	return f.existingStatus, f.existingDesc, f.existingID, f.existingErr
}

func (f *fakeCertsrv) GetCaCertificate() (string, error) {
	return f.caCert, f.caCertErr
}

func (f *fakeCertsrv) GetCaCertificateChain() (string, error) {
	return f.caChain, f.caChainErr
}

func newTestIssuer(fake *fakeCertsrv) *Issuer {
	return &Issuer{
		certServ:         fake,
		AdcsTemplateName: "WebServer",
	}
}

func TestIssueNewRequestReady(t *testing.T) {
	fake := &fakeCertsrv{
		requestStatus: adcs.Ready,
		requestDesc:   "certificate-bytes",
		requestID:     "id-1",
		caChain:       validSelfSignedCert,
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.NoError(t, err)
	assert.Equal(t, []byte("certificate-bytes"), cert)
	assert.NotNil(t, ca)
	assert.Equal(t, api.Ready, ar.Status.State)
	assert.Equal(t, "id-1", ar.Status.Id)
	assert.Equal(t, "certificate obtained successfully", ar.Status.Reason)
	assert.Equal(t, 1, fake.requestCertificateCalls)
	assert.Equal(t, 0, fake.getExistingCertificateCalls)
}

func TestIssueNewRequestPending(t *testing.T) {
	fake := &fakeCertsrv{
		requestStatus: adcs.Pending,
		requestDesc:   "still working",
		requestID:     "id-2",
		caChain:       validSelfSignedCert,
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.NoError(t, err)
	assert.Nil(t, cert)
	assert.NotNil(t, ca)
	assert.Equal(t, api.Pending, ar.Status.State)
	assert.Equal(t, "id-2", ar.Status.Id)
	assert.Equal(t, "still working", ar.Status.Reason)
}

func TestIssueNewRequestRejected(t *testing.T) {
	fake := &fakeCertsrv{
		requestStatus: adcs.Rejected,
		requestDesc:   "denied",
		requestID:     "id-3",
		caChain:       validSelfSignedCert,
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.NoError(t, err)
	assert.Nil(t, cert)
	assert.NotNil(t, ca)
	assert.Equal(t, api.Rejected, ar.Status.State)
	assert.Equal(t, "id-3", ar.Status.Id)
	assert.Equal(t, "denied", ar.Status.Reason)
}

func TestIssueNewRequestErrored(t *testing.T) {
	fake := &fakeCertsrv{
		requestStatus: adcs.Errored,
		requestDesc:   "something broke",
		requestID:     "id-4",
		caChain:       validSelfSignedCert,
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.NoError(t, err)
	assert.Nil(t, cert)
	assert.NotNil(t, ca)
	assert.Equal(t, api.Errored, ar.Status.State)
	assert.Equal(t, "id-4", ar.Status.Id)
	assert.Equal(t, "something broke", ar.Status.Reason)
}

func TestIssuePendingWithIdUsesExistingCertificate(t *testing.T) {
	fake := &fakeCertsrv{
		existingStatus: adcs.Ready,
		existingDesc:   "existing-cert-bytes",
		existingID:     "id-5",
		caChain:        validSelfSignedCert,
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{
		Status: api.AdcsRequestStatus{
			State: api.Pending,
			Id:    "id-5",
		},
	}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.NoError(t, err)
	assert.Equal(t, []byte("existing-cert-bytes"), cert)
	assert.NotNil(t, ca)
	assert.Equal(t, api.Ready, ar.Status.State)
	assert.Equal(t, 0, fake.requestCertificateCalls)
	assert.Equal(t, 1, fake.getExistingCertificateCalls)
}

func TestIssuePendingWithoutIdReturnsError(t *testing.T) {
	fake := &fakeCertsrv{}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{
		Status: api.AdcsRequestStatus{
			State: api.Pending,
			Id:    "",
		},
	}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.EqualError(t, err, "adcs id not set")
	assert.Nil(t, cert)
	assert.Nil(t, ca)
	assert.Equal(t, 0, fake.requestCertificateCalls)
	assert.Equal(t, 0, fake.getExistingCertificateCalls)
}

func TestIssueFinalStateReturnsImmediately(t *testing.T) {
	finalStates := []api.State{api.Ready, api.Rejected, api.Errored}
	for _, state := range finalStates {
		t.Run(string(state), func(t *testing.T) {
			fake := &fakeCertsrv{}
			issuer := newTestIssuer(fake)
			ar := &api.AdcsRequest{
				Status: api.AdcsRequestStatus{
					State: state,
				},
			}

			cert, ca, err := issuer.Issue(context.Background(), ar)

			assert.NoError(t, err)
			assert.Nil(t, cert)
			assert.Nil(t, ca)
			assert.Equal(t, 0, fake.requestCertificateCalls)
			assert.Equal(t, 0, fake.getExistingCertificateCalls)
		})
	}
}

func TestIssueRequestCertificateErrorPropagates(t *testing.T) {
	fake := &fakeCertsrv{
		requestErr: fmt.Errorf("boom"),
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.EqualError(t, err, "boom")
	assert.Nil(t, cert)
	assert.Nil(t, ca)
}

func TestIssueGetExistingCertificateErrorPropagates(t *testing.T) {
	fake := &fakeCertsrv{
		existingErr: fmt.Errorf("existing boom"),
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{
		Status: api.AdcsRequestStatus{
			State: api.Pending,
			Id:    "id-6",
		},
	}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.EqualError(t, err, "existing boom")
	assert.Nil(t, cert)
	assert.Nil(t, ca)
}

func TestIssueGetCaCertificateChainErrorPropagates(t *testing.T) {
	fake := &fakeCertsrv{
		requestStatus: adcs.Ready,
		requestDesc:   "certificate-bytes",
		requestID:     "id-7",
		caChainErr:    fmt.Errorf("chain boom"),
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.EqualError(t, err, "chain boom")
	assert.Nil(t, cert)
	assert.Nil(t, ca)
}

func TestIssueCaChainParseErrorPropagates(t *testing.T) {
	fake := &fakeCertsrv{
		requestStatus: adcs.Ready,
		requestDesc:   "certificate-bytes",
		requestID:     "id-8",
		caChain:       "this is not a valid pem or pkcs7 chain",
	}
	issuer := newTestIssuer(fake)
	ar := &api.AdcsRequest{}

	cert, ca, err := issuer.Issue(context.Background(), ar)

	assert.Error(t, err)
	assert.EqualError(t, err, "error decoding the pem block")
	assert.Nil(t, cert)
	assert.Nil(t, ca)
}
