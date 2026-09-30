package adcs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/stretchr/testify/assert"
)

// newKerberosCertsrv builds a KerberosCertsrv struct directly, bypassing the
// NewKerberosCertsrv constructor's live krb5 login. spnego.Client.Do only
// engages the (nil) krb5Client when the server responds 401 with
// WWW-Authenticate: Negotiate, which our mock server never does.
func newKerberosCertsrv(url string, client *http.Client) *KerberosCertsrv {
	return &KerberosCertsrv{
		url:        url,
		httpClient: spnego.NewClient(nil, client, ""),
	}
}

func TestKerberosRequestCertificate(t *testing.T) {
	t.Run("ready via direct pkix content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", ct_pkix)
			_, _ = w.Write([]byte("CERTDATA"))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, cert, id, err := s.RequestCertificate("csr", "template")
		assert.NoError(t, err)
		assert.Equal(t, Ready, status)
		assert.Equal(t, "CERTDATA", cert)
		assert.Equal(t, "none", id)
	})

	t.Run("pending via certnew.cer id pattern", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certfnsh.asp":
				w.Header().Set("Content-type", ct_html)
				_, _ = w.Write([]byte(`<script>location="certnew.cer?ReqID=101&ENC=b64"</script>`))
			case "/certnew.cer":
				assert.Equal(t, "101", r.URL.Query().Get("ReqID"))
				w.Header().Set("content-type", ct_html)
				_, _ = w.Write([]byte(dispositionHTMLBody("Taken Under Submission", true)))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, desc, id, err := s.RequestCertificate("csr", "template")
		assert.NoError(t, err)
		assert.Equal(t, Pending, status)
		assert.Equal(t, "101", id)
		assert.Contains(t, desc, "Taken Under Submission")
		assert.Contains(t, desc, "Some Last Status")
	})

	t.Run("rejected via Your Request Id pattern", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certfnsh.asp":
				w.Header().Set("Content-type", ct_html)
				_, _ = w.Write([]byte("Your Request Id is 202. Please wait."))
			case "/certnew.cer":
				assert.Equal(t, "202", r.URL.Query().Get("ReqID"))
				w.Header().Set("content-type", ct_html)
				_, _ = w.Write([]byte(dispositionHTMLBody("Denied by CA Administrator", false)))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, desc, id, err := s.RequestCertificate("csr", "template")
		assert.NoError(t, err)
		assert.Equal(t, Rejected, status)
		assert.Equal(t, "202", id)
		assert.Contains(t, desc, "Denied by CA Administrator")
	})

	t.Run("errored unknown disposition", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certfnsh.asp":
				w.Header().Set("Content-type", ct_html)
				_, _ = w.Write([]byte("certnew.cer?ReqID=303&ENC=b64"))
			case "/certnew.cer":
				assert.Equal(t, "303", r.URL.Query().Get("ReqID"))
				w.Header().Set("content-type", ct_html)
				_, _ = w.Write([]byte(dispositionHTMLBody("Some other status", false)))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, _, id, err := s.RequestCertificate("csr", "template")
		assert.NoError(t, err)
		assert.Equal(t, Errored, status)
		assert.Equal(t, "303", id)
	})

	t.Run("no id match, quoted disposition error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", ct_html)
			_, _ = w.Write([]byte(`The disposition message is "Certificate Template not supported"`))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		_, _, _, err := s.RequestCertificate("csr", "template")
		assert.Error(t, err)
		assert.EqualError(t, err, "Certificate Template not supported")
	})

	t.Run("no id match, fallback unknown error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", ct_html)
			_, _ = w.Write([]byte("Some unrelated content with no markers"))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		_, _, _, err := s.RequestCertificate("csr", "template")
		assert.Error(t, err)
		assert.EqualError(t, err, "Unknown error occured")
	})
}

func TestKerberosGetExistingCertificate(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", ct_pkix)
			_, _ = w.Write([]byte("CERTBYTES"))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, cert, id, err := s.GetExistingCertificate("42")
		assert.NoError(t, err)
		assert.Equal(t, Ready, status)
		assert.Equal(t, "CERTBYTES", cert)
		assert.Equal(t, "42", id)
	})

	t.Run("pending", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", ct_html)
			_, _ = w.Write([]byte(dispositionHTMLBody("Taken Under Submission", true)))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, desc, id, err := s.GetExistingCertificate("42")
		assert.NoError(t, err)
		assert.Equal(t, Pending, status)
		assert.Equal(t, "42", id)
		assert.Contains(t, desc, "Taken Under Submission")
	})

	t.Run("rejected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", ct_html)
			_, _ = w.Write([]byte(dispositionHTMLBody("Denied by CA Administrator", false)))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, desc, id, err := s.GetExistingCertificate("42")
		assert.NoError(t, err)
		assert.Equal(t, Rejected, status)
		assert.Equal(t, "42", id)
		assert.Contains(t, desc, "Denied by CA Administrator")
	})

	t.Run("errored unknown disposition", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", ct_html)
			_, _ = w.Write([]byte(dispositionHTMLBody("Some other status", false)))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, _, id, err := s.GetExistingCertificate("42")
		assert.NoError(t, err)
		assert.Equal(t, Errored, status)
		assert.Equal(t, "42", id)
	})

	t.Run("disposition message not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", ct_html)
			_, _ = w.Write([]byte("no disposition markers here"))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, desc, id, err := s.GetExistingCertificate("42")
		assert.Error(t, err)
		assert.Equal(t, Unknown, status)
		assert.Equal(t, "42", id)
		assert.Contains(t, desc, "unknown")
		assert.Contains(t, err.Error(), "disposition message unknown")
	})

	t.Run("unexpected content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/plain")
			_, _ = w.Write([]byte("plain text"))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		status, _, _, err := s.GetExistingCertificate("42")
		assert.Error(t, err)
		assert.Equal(t, Unknown, status)
		assert.Contains(t, err.Error(), "unexpected content type")
	})

	t.Run("non-200 status panics on nil error dereference", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		// Source code calls err.Error() on a nil error in this branch; this
		// documents the existing behavior without modifying production code.
		assert.Panics(t, func() {
			_, _, _, _ = s.GetExistingCertificate("42")
		})
	})
}

func TestKerberosObtainCaCertificate(t *testing.T) {
	t.Run("GetCaCertificate success with renewal marker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certcarc.asp":
				_, _ = w.Write([]byte("var nRenewals=5;"))
			case "/certnew.cer":
				assert.Equal(t, "5", r.URL.Query().Get("Renewal"))
				w.Header().Set("content-type", ct_pkix)
				_, _ = w.Write([]byte("CACERT"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		cert, err := s.GetCaCertificate()
		assert.NoError(t, err)
		assert.Equal(t, "CACERT", cert)
	})

	t.Run("GetCaCertificateChain success without renewal marker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certcarc.asp":
				_, _ = w.Write([]byte("no renewal info here"))
			case "/certnew.p7b":
				assert.Equal(t, "0", r.URL.Query().Get("Renewal"))
				w.Header().Set("content-type", ct_pkcs7)
				_, _ = w.Write([]byte("CACHAIN"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		cert, err := s.GetCaCertificateChain()
		assert.NoError(t, err)
		assert.Equal(t, "CACHAIN", cert)
	})

	t.Run("wrong content type error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certcarc.asp":
				_, _ = w.Write([]byte("var nRenewals=1;"))
			case "/certnew.cer":
				w.Header().Set("content-type", "text/plain")
				_, _ = w.Write([]byte("not a cert"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		_, err := s.GetCaCertificate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected content type")
	})

	t.Run("non-200 status panics on nil error dereference", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/certcarc.asp":
				_, _ = w.Write([]byte("var nRenewals=1;"))
			case "/certnew.cer":
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("boom"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		assert.Panics(t, func() {
			_, _ = s.GetCaCertificate()
		})
	})
}

func TestKerberosVerifyKerberos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		s := newKerberosCertsrv(server.URL, server.Client())
		ok, err := s.verifyKerberos()
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("http client error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		client := server.Client()
		url := server.URL
		server.Close() // closing before use forces a connection error

		s := newKerberosCertsrv(url, client)
		ok, err := s.verifyKerberos()
		assert.False(t, ok)
		assert.Error(t, err)
	})
}

func TestNewKerberosCertsrv(t *testing.T) {
	t.Run("errors when krb5.conf is unavailable", func(t *testing.T) {
		// /etc/krb5.conf is very unlikely to exist in this sandbox, so
		// config.Load should fail and NewKerberosCertsrv should surface
		// that error without attempting any login or network activity.
		c, err := NewKerberosCertsrv("http://example.invalid", "user", "REALM", "pass", nil, false)
		assert.Error(t, err)
		assert.Nil(t, c)
	})
}
