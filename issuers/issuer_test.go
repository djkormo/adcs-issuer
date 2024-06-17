package issuers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("issuer_test")
)

// TODO: provide proper PKCS7 certificates for testing.
// TODO: create makefile to populate testdata (openssl)
// 2 tests will fail!
// testdata/pkcs7.pem and testdata/x509.pem aren't provided, for the tests to be able to run, please provide own certs of these format.

func TestParsingCaCertShouldReturnX509(t *testing.T) {
	// arrange
	pkcs7Pem, err := os.ReadFile("testdata/pkcs7.pem")
	assert.NoError(t, err)

	validX509Certificate, err := os.ReadFile("testdata/x509.pem")
	assert.NoError(t, err)
	// act

	parsedCaCert, err := parseCaCert(pkcs7Pem, log)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, validX509Certificate, parsedCaCert)
}

func TestIncorrectFormatPkcs(t *testing.T) {
	//arrange
	incorrectPKCS7Cert, err := os.ReadFile("testdata/incorrectPKCS7Cert.pem")
	assert.NoError(t, err)

	// act

	ca, err := parseCaCert(incorrectPKCS7Cert, log)

	// assert
	assert.EqualError(t, err, "parsing PKCS7: ber2der: BER tag length is more than available data")
	assert.Nil(t, ca, "expecting ca to be empty")
}

func TestEmptyPkcs(t *testing.T) {
	// arrange
	emptyPKCS7 := []byte(``)

	// act

	ca, err := parseCaCert(emptyPKCS7, log)

	// assert
	assert.EqualError(t, err, "error decoding the pem block")
	assert.Nil(t, ca, "expecting ca to be empty")
}

func TestIncorrectCertFormat(t *testing.T) {
	// arrange
	incorrectCertFormat := []byte(`This is not correct!`)

	// act

	ca, err := parseCaCert(incorrectCertFormat, log)

	// assert
	assert.Error(t, err)
	assert.EqualError(t, err, "error decoding the pem block")
	assert.Nil(t, ca, "expecting ca to be empty ")
}

func TestParseCaCertCorrectPKCS7(t *testing.T) {
	// arrange
	// raw format pkcs7.p7b from cfss testdata (https://github.com/cloudflare/cfssl/tree/master/helpers/testdata)
	rawPkcs7, err := os.ReadFile("testdata/cfss_rawPKCS7.p7b")
	assert.NoError(t, err)
	cfssOutputX509, err := os.ReadFile("testdata/cfss_outputx509.pem")
	assert.NoError(t, err)

	// act

	ca, err := parseCaCert(rawPkcs7, log)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, cfssOutputX509, ca)
}

func TestCorrectX509Cert(t *testing.T) {
	// arrange
	// raw format pkcs7.p7b from cfss testdata (https://github.com/cloudflare/cfssl/tree/master/helpers/testdata)
	x509, err := os.ReadFile("testdata/x509.pem")

	if err != nil {
		fmt.Println("TestCorrectX509Cert")
	}
	// act

	parsedCaCert, err := parseCaCert(x509, log)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, x509, parsedCaCert)
}
