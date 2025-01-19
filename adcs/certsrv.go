package adcs

type AdcsResponseStatus int

const (
	Unknown  AdcsResponseStatus = 0
	Pending  AdcsResponseStatus = 1
	Ready    AdcsResponseStatus = 2
	Errored  AdcsResponseStatus = 3
	Rejected AdcsResponseStatus = 4
)

type AdcsCertsrv interface {
	// Request new certificate.
	// Returns (cert status, certificate or description, id, error)
	// If cert status is 'Unknown' the state of the certificate info couldn't be obtained from  certsrv. Check for error.
	// If cert status is 'Ready' the cert is returned immediately in 'certificate'.
	// If cert status is 'Pending' the cert can be obtained later with getExistingCertificate using the 'id' (see 'description' for more details)
	// If cert status is 'Error' see 'description' for details.
	RequestCertificate(csr string, template string) (AdcsResponseStatus, string, string, string, string, error)

	// Get previously requested certicate from Certserv
	// Returns (cert status, certificate or description, id, error)
	// If cert status is 'Unknown' the state of the certificate info couldn't be obtained from certsrv. Check for error.
	// If cert status is 'Ready' the cert is returned in 'certificate'.
	// If cert status is 'Pending' the cert can be obtained later with getExistingCertificate using the 'id' (see 'description' for more details)
	// If cert status is 'Error' see 'description' for details.
	GetExistingCertificate(id string) (AdcsResponseStatus, string, string, string, string, error)

	// Get the certsrv' CA cert
	// Returns ( certificate, error)
	GetCaCertificate() (string, error)

	// Get the certsrv' CA chain
	// Returns (certificate, error)
	GetCaCertificateChain() (string, error)
}
