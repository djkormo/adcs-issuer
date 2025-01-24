package issuers

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	//cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	//cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/djkormo/adcs-issuer/adcs"
	api "github.com/djkormo/adcs-issuer/api/v1"
	"github.com/fullsailor/pkcs7"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Issuer struct {
	client.Client
	certServ            adcs.AdcsCertsrv
	RetryInterval       time.Duration
	StatusCheckInterval time.Duration
	AdcsTemplateName    string
}

// Go to ADCS for a certificate. If current status is 'Pending' then
// check for existing request. Otherwise ask for new.
// The current status is set in the passed request.
// If status is 'Ready' the returns include certificate and CA cert respectively.
func (i *Issuer) Issue(ctx context.Context, ar *api.AdcsRequest) ([]byte, []byte, error) {
	log := ctrl.LoggerFrom(ctx)
	var adcsResponseStatus adcs.AdcsResponseStatus
	var desc string
	var id string
	var err error
	if ar.Status.State != api.Unknown {
		// Of all the statuses only Pending requires processing.
		// All others are final
		if ar.Status.State == api.Pending {
			// Check the status of the request on the ADCS
			if ar.Status.Id == "" {
				return nil, nil, fmt.Errorf("adcs id not set")
			}
			adcsResponseStatus, desc, id, err = i.certServ.GetExistingCertificate(ar.Status.Id)
		} else {
			// Nothing to do
			return nil, nil, nil
		}
	} else {
		// New request
		adcsResponseStatus, desc, id, err = i.certServ.RequestCertificate(string(ar.Spec.CSRPEM), i.AdcsTemplateName)

		if log.V(5).Enabled() {
			log.V(5).Info("new adcsRequest", "adcs response status", adcsResponseStatus, "desc", desc, "id", id)

		}
	}
	if err != nil {
		// This is a local error
		return nil, nil, err
	}

	var cert []byte
	switch adcsResponseStatus {
	case adcs.Pending:
		// It must be checked again later
		ar.Status.State = api.Pending
		ar.Status.Id = id
		ar.Status.Reason = desc
	case adcs.Ready:
		// Certificate obtained successfully
		ar.Status.State = api.Ready
		ar.Status.Id = id
		ar.Status.Reason = "certificate obtained successfully"
		cert = []byte(desc)
	case adcs.Rejected:
		// Certificate request rejected by ADCS
		ar.Status.State = api.Rejected
		ar.Status.Id = id
		ar.Status.Reason = desc
	case adcs.Errored:
		// Unknown problem occured on ADCS
		ar.Status.State = api.Errored
		ar.Status.Id = id
		ar.Status.Reason = desc
	}

	// Get a certificateChain from the server.
	certChain, err := i.certServ.GetCaCertificateChain()

	if err != nil {
		return nil, nil, err
	}

	// Parse and encode the certificateChain to a valid x509 certificate.
	ca, err := parseCaCert([]byte(certChain), log)

	if err != nil {
		log.Error(err, "something went wrong parsing to x509")
		return nil, nil, err
	}

	if log.V(4).Enabled() {
		s := string(cert)
		log.V(4).Info("parsed certificate", "certificate", s)
	}

	// log.V(4).Info("will return cert", "cert", cert)

	return cert, ca, nil

}

// x509Bytes is a slice of bytes
// type x509Bytes []byte

// ParseCaCert accepts bytes representing a certificate and returns x509 certificate encoded pem

func parseCaCert(cc []byte, log logr.Logger) ([]byte, error) {

	// decode Pem from certificate into block
	block, rest := pem.Decode([]byte(cc))
	if block == nil {

		if log.V(3).Enabled() {
			s := string(rest)
			log.Info("tried to decode pem", "rest", s)

		}
		return nil, errors.New("error decoding the pem block")
	}

	// parse the decoded pem block to x509 encoded block
	b, err := tryParseX509(block)
	if err != nil {
		return nil, err
	}

	// encodes the x509 encoded block to a valid x509 certificate encoded pem.
	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: b,
	})

	return pem, nil
}

// TryParseX509 accepts *pem.Block and returns this with a valid x509 encoded block.
func tryParseX509(block *pem.Block) ([]byte, error) {
	// if certificate is already x509 encoded, return the certificate, otherwise continue and parse.
	_, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		return block.Bytes, nil
	}

	b, err := pkcs7.Parse(block.Bytes)
	if err == nil {
		if len(b.Certificates) == 0 {
			return nil, fmt.Errorf("expected one or more certificates")
		}
		return b.Certificates[0].Raw, nil
	}

	err = fmt.Errorf("parsing PKCS7: %w", err)
	return nil, err
}
