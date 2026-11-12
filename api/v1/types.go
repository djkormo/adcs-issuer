package v1

type LocalObjectReference struct {
	// Name of the referent.
	Name string `json:"name"`
}

// CABundleReference identifies a key in a Secret or ConfigMap in the issuer's namespace.
type CABundleReference struct {
	// Name of the referent.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Kind of the referent. Defaults to Secret.
	// +optional
	// +kubebuilder:validation:Enum=Secret;ConfigMap
	Kind string `json:"kind,omitempty"`

	// Key containing the PEM encoded CA bundle. Defaults to ca.crt.
	// +optional
	Key string `json:"key,omitempty"`
}

// NamespacedCABundleReference identifies a CA bundle for a ClusterAdcsIssuer.
type NamespacedCABundleReference struct {
	CABundleReference `json:",inline"`

	// Namespace of the referent. Defaults to --cluster-resource-namespace
	// (kube-system unless configured otherwise).
	// +optional
	Namespace string `json:"namespace,omitempty"`
}
