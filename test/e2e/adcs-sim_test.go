package e2e

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"testing"

	"github.com/nokia/adcs-issuer/adcs"
	"github.com/stretchr/testify/assert"
)

func TestADCSSim(t *testing.T) {
	//adcsSimCertPool := load server CA so the client trusts adcs-sim.
	adcsSimCertPool := &x509.CertPool{}
	cs, err := adcs.NewNtlmCertsrv("https://localhost:8443", "", "", adcsSimCertPool, true)
	assert.NoError(t, err)

	csr := &x509.CertificateRequest{
		Version:            3,
		SignatureAlgorithm: x509.SHA512WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		Subject: pkix.Name{
			Organization: []string{"organization"},
			CommonName:   "commonName",
		},
		DNSNames:        []string{"dnsNames"},
		IPAddresses:     []net.IP{},
		ExtraExtensions: []pkix.Extension{},
	}

	keySize := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	assert.NoError(t, err)
	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, csr, privateKey)

	var pemBuffer bytes.Buffer
	err = pem.Encode(&pemBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	assert.NoError(t, err)

	const adcsCertTemplate = "BasicSSLWebServer"
	adcsResponseStatus, desc, id, err := cs.RequestCertificate(pemBuffer.String(), adcsCertTemplate)
	assert.NoError(t, err)

	//TODO assert
	fmt.Println("adcsResponseStatus", adcsResponseStatus)
	fmt.Println("desc", desc)
	fmt.Println("id", id)
}
