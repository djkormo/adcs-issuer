package adcs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"os"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	"github.com/Azure/go-ntlmssp"
	"k8s.io/klog"
)

type NtlmCertsrv struct {
	url        string
	username   string
	password   string
	ca         string
	httpClient *http.Client
}

const (
	certnew_cer = "certnew.cer"
	certnew_p7b = "certnew.p7b"
	certcarc    = "certcarc.asp"
	certfnsh    = "certfnsh.asp"

	ct_pkix   = "application/pkix-cert"
	ct_pkcs7  = "application/x-pkcs7-certificates"
	ct_html   = "text/html"
	ct_urlenc = "application/x-www-form-urlencoded"
)

func NewNtlmCertsrv(url string, username string, password string, caCertPool *x509.CertPool, verify bool) (AdcsCertsrv, error) {
	log := log.Log.WithName("newNtlm")
	var client *http.Client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            caCertPool,
		},
	}
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("NTLM verification start", "username", username, "password", password, "url", url)
	}
	if username != "" && password != "" {
		// Set up NTLM authentication
		client = &http.Client{
			Transport: ntlmssp.Negotiator{
				RoundTripper: transport,
			},
		}
		if os.Getenv("ENABLE_DEBUG") == "true" {
			log.Info("NTLM verification Using NTLM")
		}
	} else {
		// Plain client with no NTLM
		client = &http.Client{
			Transport: transport,
		}
		if os.Getenv("ENABLE_DEBUG") == "true" {
			log.Info("NTLM verification not using NTLM")
		}
		log.V(5).Info("NTLM verification not using NTL")
	}

	c := &NtlmCertsrv{
		url:        url,
		username:   username,
		password:   password,
		httpClient: client,
	}
	if verify {
		success, err := c.verifyNtlm()
		if !success {
			return nil, err
		}
	}
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("NTLM verification stop", "username", username, "password", password, "url", url)
	}
	return c, nil
}

// Check if NTLM authentication is working for current credentials and URL
func (s *NtlmCertsrv) verifyNtlm() (bool, error) {
	log := log.Log.WithName("verifyNtlm")
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("NTLM verification", "username", s.username, "url", s.url)
	}
	log.V(5).Info("NTLM verification", "username", s.username, "url", s.url)

	req, _ := http.NewRequest("GET", s.url, nil)
	req.SetBasicAuth(s.username, s.password)
	res, err := s.httpClient.Do(req)
	if err != nil {
		log.Error(err, "ADCS server error")
		return false, err
	}
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("NTLM verification successful", "status", res.Status)
	}
	log.V(5).Info("NTLM verification successful", "status", res.Status)
	return true, nil
}

/*
 * Returns:
 * - Certificate response status
 * - Certificate (if status is Ready) or status description (if status is not Ready)
 * - ADCS Request ID
 * - Error
 */
func (s *NtlmCertsrv) GetExistingCertificate(id string) (AdcsResponseStatus, string, string, error) {
	log := log.Log.WithName("GetExistingCertificate")
	var certStatus AdcsResponseStatus = Unknown

	url := fmt.Sprintf("%s/%s?ReqID=%s&ENC=b64", s.url, certnew_cer, id)
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("Making url request", "url", url)
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("User-agent", "Mozilla")
	res, err := s.httpClient.Do(req)

	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("Making url request", "res", res)
	}

	if err != nil {
		log.Error(err, "ADCS Certserv error")
		return certStatus, "", id, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		switch ct := strings.Split(res.Header.Get(http.CanonicalHeaderKey("content-type")), ";"); ct[0] {
		case ct_html:
			// Denied or pending
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Error(err, "Cannot read ADCS Certserv response")
				return certStatus, "", id, err
			}
			bodyString := string(body)
			dispositionMessage := "unknown"
			exp := regexp.MustCompile(`Disposition message:[^\t]+\t\t([^\r\n]+)`)
			found := exp.FindStringSubmatch(bodyString)
			if len(found) > 1 {
				dispositionMessage = found[1]
				expPending := regexp.MustCompile(`.*Taken Under Submission*.`)
				expRejected := regexp.MustCompile(`.*Denied by*.`)
				switch true {
				case expPending.MatchString(bodyString):
					certStatus = Pending
				case expRejected.MatchString(bodyString):
					certStatus = Rejected
				default:
					certStatus = Errored
				}

			} else {
				// If the response page is not formatted as we expect it
				// we just log the entire page
				disp := bodyString
				if len(found) == 1 {
					// Or at least the 'Disposition message' section
					disp = found[0]
				}
				err = fmt.Errorf("disposition message unknown: %s", disp)
				log.Error(err, "Unknown error with ADCS")
			}

			lastStatusMessage := ""
			exp = regexp.MustCompile(`LastStatus:[^\t]+\t\t([^\r\n]+)`)
			found = exp.FindStringSubmatch(bodyString)
			if len(found) > 1 {
				lastStatusMessage = " " + found[1]
			} else {
				log.Info("Last status unknown.")
				log.V(5).Info("Last status unknown.")
			}
			return certStatus, dispositionMessage + lastStatusMessage, id, err

		case ct_pkix:
			// Certificate
			cert, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Error(err, "Cannot read ADCS Certserv response")
				return certStatus, "", id, err
			}
			return Ready, string(cert), id, nil
		default:
			err = fmt.Errorf("unexpected content type %s:", ct)
			log.Error(err, "Unexpected content type")
			return certStatus, "", id, err
		}
	}
	return certStatus, "", id, fmt.Errorf("ADCS Certsrv response status %s. Error: %s", res.Status, err.Error())

}

/*
 * Returns:
 * - Certificate response status
 * - Certificate (if status is Ready) or status description (if status is not Ready)
 * - ADCS Request ID (if known)
 * - Error
 */
func (s *NtlmCertsrv) RequestCertificate(csr string, template string) (AdcsResponseStatus, string, string, error) {
	log := log.Log.WithName("RequestCertificate").WithValues("template", template)
	var certStatus AdcsResponseStatus = Unknown

	log.V(1).Info("Starting certificate request")

	url := fmt.Sprintf("%s/%s", s.url, certfnsh)
	log.Info("Starting certificate request", "url", url)
	params := neturl.Values{
		"Mode":                {"newreq"},
		"CertRequest":         {csr},
		"CertAttrib":          {"CertificateTemplate:" + template},
		"FriendlyType":        {"Saved-Request Certificate"},
		"TargetStoreFlags":    {"0"},
		"SaveCert":            {"yes"},
		"CertificateTemplate": {template},
	}

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(params.Encode()))

	if err != nil {
		log.Error(err, "Cannot create request")
		return certStatus, "", "", err
	}
	req.SetBasicAuth(s.username, s.password)
	klog.V(5).Infof("Username as BasicAuth: \n %v ", s.username)

	req.Header.Set("User-agent", "Mozilla")
	req.Header.Set("Content-type", ct_urlenc)

	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("Sending request", "request", req)
	}

	res, err := s.httpClient.Do(req)
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("Sending request", "error", err)
		log.Info("Sending request", "response", res)
		//log.Info("Sending request", "response Header", res.Header)
		//log.Info("Sending request", "response Request.URL", res.Request.URL)
		//log.Info("Sending request", "Status Request.URL", res.Status)
	}

	if err != nil {
		log.Error(err, "ADCS Certserv error")
		return certStatus, "", "", err
	}

	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("Sending request", "response", res)
	}

	body, err := ioutil.ReadAll(res.Body)

	log.Info("Body", "body", body)

	if res.Header.Get("Content-type") == ct_pkix {
		// klog.V(4).Infof("klog_v4: returned [Ready] %v", Ready)
		return Ready, string(body), "none", nil
	}
	if err != nil {
		log.Error(err, "Cannot read ADCS Certserv response")
		return certStatus, "", "", err
	}

	bodyString := string(body)

	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("Body", "body", bodyString)
	}

	exp := regexp.MustCompile(`certnew.cer\?ReqID=([0-9]+)&`)
	found := exp.FindStringSubmatch(bodyString)
	certId := ""
	if len(found) > 1 {
		certId = found[1]
	} else {
		exp = regexp.MustCompile(`Your Request Id is ([0-9]+).`)
		found = exp.FindStringSubmatch(bodyString)
		if len(found) > 1 {
			certId = found[1]
		} else {
			errorString := ""
			exp = regexp.MustCompile(`The disposition message is "([^"]+)`)
			found = exp.FindStringSubmatch(bodyString)
			var errorContext []interface{}
			if len(found) > 1 {
				errorString = found[1]
			} else {
				errorString = "Unknown error occured"
				errorContext = []interface{}{"body", bodyString}
			}
			err := errors.New(errorString)
			// TODO
			log.Error(err, "Couldn't obtain new certificate ID", errorContext...)
			return certStatus, "", "", fmt.Errorf(errorString)
		}
	}

	return s.GetExistingCertificate(certId)
}

func (s *NtlmCertsrv) obtainCaCertificate(certPage string, expectedContentType string) (string, error) {
	log := log.Log.WithName("obtainCaCertificate")

	// Check for newest renewal number
	url := fmt.Sprintf("%s/%s", s.url, certcarc)
	// klog.V(4).Infof("inside obtainCaCertificate: going to url: %v ", url)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("User-agent", "Mozilla")
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("obtainCaCertificate start", "req", req, "url", url)
		log.Info("obtainCaCertificate start", "password", s.password, "username", s.username)
	}
	res1, err := s.httpClient.Do(req)

	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Info("obtainCaCertificate start", "res1", res1)
	}

	if err != nil {
		log.Error(err, "ADCS Certserv error")
		return "", err
	}
	defer res1.Body.Close()
	body, err := ioutil.ReadAll(res1.Body)
	if err != nil {
		log.Error(err, "Cannot read ADCS Certserv response")
		return "", err
	}

	renewal := "0"
	exp := regexp.MustCompile(`var nRenewals=([0-9]+);`)
	found := exp.FindStringSubmatch(string(body))
	if len(found) > 1 {
		renewal = found[1]
	} else {
		log.Info("Renewal not found. Using '0'.")
	}

	// Get CA cert (newest renewal number)
	url = fmt.Sprintf("%s/%s?ReqID=CACert&ENC=b64&Renewal=%s", s.url, certPage, renewal)
	req, _ = http.NewRequest("GET", url, nil)
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("User-agent", "Mozilla")

	res2, err := s.httpClient.Do(req)

	// klog.V(4).Infof("Response Getting CAcert obtainingCaCertificate: %v ", res2)

	if err != nil {
		log.Error(err, "ADCS Certserv error")
		return "", err
	}
	defer res2.Body.Close()

	if res2.StatusCode == http.StatusOK {
		ct := res2.Header.Get(http.CanonicalHeaderKey("content-type"))
		if expectedContentType != ct {
			err = errors.New("Unexpected content type")
			log.Error(err, err.Error(), "content type", ct)
			return "", err
		}
		body, err := ioutil.ReadAll(res2.Body)
		if err != nil {
			log.Error(err, "Cannot read ADCS Certserv response")
			return "", err
		}
		// klog.V(4).Infof("return body adcs certserv response: %v ", body)
		return string(body), nil
	}
	return "", fmt.Errorf("ADCS Certsrv response status %s. Error: %s", res2.Status, err.Error())
}
func (s *NtlmCertsrv) GetCaCertificate() (string, error) {
	log.Log.WithName("GetCaCertificate").Info("Getting CA from ADCS Certsrv", "url", s.url)
	return s.obtainCaCertificate(certnew_cer, ct_pkix)
}
func (s *NtlmCertsrv) GetCaCertificateChain() (string, error) {
	log.Log.WithName("GetCaCertificateChain").Info("Getting CA Chain from ADCS Certsrv", "url", s.url)
	return s.obtainCaCertificate(certnew_p7b, ct_pkcs7)
}
