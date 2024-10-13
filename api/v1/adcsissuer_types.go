/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	// Ommitting  ADCS cacert verification
	// +optional
	//+kubebuilder:default:=false
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// Ommitting  NTLM Support
	// +optional
	//+kubebuilder:default:=false
	SkipNTLM bool `json:"skipNTLM,omitempty"`

	// Timeout for connection (in time.ParseDuration() format)
	// Default 30s .
	// +optional
	//+kubebuilder:default:="30s"
	ConnectionTimeout string `json:"connectionTimeout,omitempty"`
}

// AdcsIssuerStatus defines the observed state of AdcsIssuer
type AdcsIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=adcsissuers,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TemplateName",type=string,JSONPath=`.spec.templateName`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="ConnectionTimeout",type=string,priority=1,JSONPath=`.spec.connectionTimeout`
// +kubebuilder:printcolumn:name="InsecureSkipVerify",type=boolean,priority=1,JSONPath=`.spec.insecureSkipVerify`
// +kubebuilder:printcolumn:name="SkipNTLM",type=boolean,priority=1,JSONPath=`.spec.skipNTLM`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
