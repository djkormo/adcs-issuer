package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AdcsRequestSpec defines the desired state of AdcsRequest
type AdcsRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Certificate signing request bytes in PEM encoding.
	// This will be used when finalizing the request.
	// This field must be set on the request.
	CSRPEM []byte `json:"csr"`

	// IssuerRef references a properly configured AdcsIssuer which should
	// be used to serve this AdcsRequest.
	// If the Issuer does not exist, processing will be retried.
	// If the Issuer is not an 'ADCS' Issuer, an error will be returned and the
	// ADCSRequest will be marked as failed.
	IssuerRef cmmeta.ObjectReference `json:"issuerRef"`
}

// AdcsRequestStatus defines the observed state of AdcsRequest
type AdcsRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ID of the Request assigned by the ADCS.
	// This will initially be empty when the resource is first created.
	// The ADCSRequest controller will populate this field when the Request is accepted by ADCS.
	// This field will be immutable after it is initially set.
	// +optional
	Id string `json:"id,omitempty"`

	// State contains the current state of this ADCSRequest resource.
	// States 'ready' and 'rejected' are 'final'
	// +optional
	State State `json:"state,omitempty"`

	// Reason optionally provides more information about a why the AdcsRequest is in
	// the current state.
	// +optional
	Reason string `json:"reason,omitempty"`
}

// State represents the state of an ADCSRequest.
// Clients utilising this type must also gracefully handle unknown
// values, as the contents of this enumeration may be added to over time.
// +kubebuilder:validation:Enum=pending;ready;errored;rejected
type State string

const (
	// It is used to represent an unrecognised value.
	Unknown State = ""

	// If a request is marked 'Pending', is's waiting for acceptance on the ADCS.
	// This is a transient state.
	Pending State = "pending"

	// If a request  is 'ready', the certificate has been issued by the ADCS server.
	// This is a final state.
	Ready State = "ready"

	// Errored signifies that the ADCS request has errored for some reason.
	// This is a catch-all state, and is used for marking internal cert-manager
	// errors such as validation failures.
	// This is a final state.
	Errored State = "errored"

	// The 'rejected' state is used when ADCS denied signing the request.
	Rejected State = "rejected"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=adcsrequests,scope=Namespaced
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AdcsRequest is the Schema for the adcsrequests API
type AdcsRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdcsRequestSpec   `json:"spec,omitempty"`
	Status AdcsRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AdcsRequestList contains a list of AdcsRequest
type AdcsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdcsRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdcsRequest{}, &AdcsRequestList{})
}
