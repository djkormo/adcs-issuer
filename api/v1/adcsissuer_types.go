package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AdcsIssuerSpec defines the desired state of AdcsIssuer
type AdcsIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URL is the base URL for the ADCS instance
	URL string `json:"url"`

	// CredentialsRef is a reference to a Secret containing the username and
	// password for the ADCS server.
	// The secret must contain two keys, 'username' and 'password'.
	CredentialsRef LocalObjectReference `json:"credentialsRef"`

	// CABundle is a PEM encoded TLS certifiate to use to verify connections to
	// the ADCS server.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// How often to check for request status in the server (in time.ParseDuration() format)
	// Default 6 hours.
	// +optional
	StatusCheckInterval string `json:"statusCheckInterval,omitempty"`

	// How often to retry in case of communication errors (in time.ParseDuration() format)
	// Default 1 hour.
	// +optional
	RetryInterval string `json:"retryInterval,omitempty"`

	// Which ADCS Template should this issuer use
	// Defaults to the what is specified in main.go or as an cli option.
	// +optional
	TemplateName string `json:"templateName,omitempty"`

	// InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
	// If InsecureSkipVerify is true, crypto/tls accepts any certificate presented by the server.
	// In this mode, TLS is susceptible to machine-in-the-middle attacks unless custom verification is used.
	// This should be used only for testing or in combination with VerifyConnection or VerifyPeerCertificate.
	// Default: false
	// +optional
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
}

// AdcsIssuerStatus defines the observed state of AdcsIssuer
type AdcsIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=adcsissuers,scope=Namespaced
// +kubebuilder:subresource:status

// AdcsIssuer is the Schema for the adcsissuers API
type AdcsIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdcsIssuerSpec   `json:"spec,omitempty"`
	Status AdcsIssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AdcsIssuerList contains a list of AdcsIssuer
type AdcsIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdcsIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdcsIssuer{}, &AdcsIssuerList{})
}
