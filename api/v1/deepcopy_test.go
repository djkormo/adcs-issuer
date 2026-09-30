package v1

import (
	"testing"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFullAdcsIssuer() *AdcsIssuer {
	return &AdcsIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdcsIssuer",
			APIVersion: "adcs.certmanager.csf.nokia.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-issuer",
			Namespace: "default",
			Labels:    map[string]string{"foo": "bar"},
		},
		Spec: AdcsIssuerSpec{
			URL:                 "https://adcs.example.com",
			CredentialsRef:      LocalObjectReference{Name: "creds"},
			CABundle:            []byte("ca-bundle-bytes"),
			StatusCheckInterval: "6h",
			RetryInterval:       "1h",
			TemplateName:        "WebServer",
		},
		Status: AdcsIssuerStatus{},
	}
}

func newFullClusterAdcsIssuer() *ClusterAdcsIssuer {
	return &ClusterAdcsIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAdcsIssuer",
			APIVersion: "adcs.certmanager.csf.nokia.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-cluster-issuer",
			Labels: map[string]string{"foo": "bar"},
		},
		Spec: ClusterAdcsIssuerSpec{
			URL:                 "https://adcs.example.com",
			CredentialsRef:      LocalObjectReference{Name: "creds"},
			CABundle:            []byte("cluster-ca-bundle-bytes"),
			StatusCheckInterval: "6h",
			RetryInterval:       "1h",
			TemplateName:        "WebServer",
		},
		Status: ClusterAdcsIssuerStatus{},
	}
}

func newFullAdcsRequest() *AdcsRequest {
	return &AdcsRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdcsRequest",
			APIVersion: "adcs.certmanager.csf.nokia.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-request",
			Namespace: "default",
			Labels:    map[string]string{"foo": "bar"},
		},
		Spec: AdcsRequestSpec{
			CSRPEM: []byte("csr-bytes"),
			IssuerRef: cmmeta.ObjectReference{
				Kind:  "AdcsIssuer",
				Name:  "issuer-name",
				Group: "adcs.certmanager.csf.nokia.com",
			},
		},
		Status: AdcsRequestStatus{
			Id:     "1234",
			State:  Pending,
			Reason: "waiting",
		},
	}
}

func TestLocalObjectReferenceDeepCopy(t *testing.T) {
	in := &LocalObjectReference{Name: "secret-name"}

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	var into LocalObjectReference
	in.DeepCopyInto(&into)
	assert.Equal(t, *in, into)

	// Also cover the nil receiver branch.
	var nilRef *LocalObjectReference
	assert.Nil(t, nilRef.DeepCopy())
}

func TestAdcsIssuerSpecDeepCopy(t *testing.T) {
	in := newFullAdcsIssuer().Spec

	out := in.DeepCopy()
	assert.Equal(t, &in, out)
	assert.NotSame(t, &in, out)

	// Mutating the copy's slice must not affect the original.
	out.CABundle[0] = 'X'
	assert.NotEqual(t, in.CABundle[0], out.CABundle[0])

	var nilSpec *AdcsIssuerSpec
	assert.Nil(t, nilSpec.DeepCopy())

	// Spec with nil CABundle should not allocate a new slice.
	emptySpec := AdcsIssuerSpec{URL: "u"}
	var into AdcsIssuerSpec
	emptySpec.DeepCopyInto(&into)
	assert.Nil(t, into.CABundle)
}

func TestAdcsIssuerStatusDeepCopy(t *testing.T) {
	in := &AdcsIssuerStatus{}
	out := in.DeepCopy()
	assert.Equal(t, in, out)

	var nilStatus *AdcsIssuerStatus
	assert.Nil(t, nilStatus.DeepCopy())
}

func TestAdcsIssuerDeepCopy(t *testing.T) {
	in := newFullAdcsIssuer()

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	// Prove deep copy: mutate nested slice/map on the copy, original unaffected.
	out.Spec.CABundle[0] = 'Z'
	assert.NotEqual(t, in.Spec.CABundle[0], out.Spec.CABundle[0])

	out.Labels["foo"] = "changed"
	assert.NotEqual(t, in.Labels["foo"], out.Labels["foo"])

	var nilIssuer *AdcsIssuer
	assert.Nil(t, nilIssuer.DeepCopy())

	obj := in.DeepCopyObject()
	assert.Equal(t, in, obj)

	var nilIssuerObj *AdcsIssuer
	assert.Nil(t, nilIssuerObj.DeepCopyObject())
}

func TestAdcsIssuerListDeepCopy(t *testing.T) {
	in := &AdcsIssuerList{
		TypeMeta: metav1.TypeMeta{Kind: "AdcsIssuerList"},
		ListMeta: metav1.ListMeta{ResourceVersion: "1"},
		Items:    []AdcsIssuer{*newFullAdcsIssuer(), *newFullAdcsIssuer()},
	}

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	out.Items[0].Spec.CABundle[0] = 'Y'
	assert.NotEqual(t, in.Items[0].Spec.CABundle[0], out.Items[0].Spec.CABundle[0])

	var nilList *AdcsIssuerList
	assert.Nil(t, nilList.DeepCopy())

	obj := in.DeepCopyObject()
	assert.Equal(t, in, obj)

	var nilListObj *AdcsIssuerList
	assert.Nil(t, nilListObj.DeepCopyObject())

	// Also cover the nil-Items branch.
	emptyList := &AdcsIssuerList{}
	emptyOut := emptyList.DeepCopy()
	assert.Nil(t, emptyOut.Items)
}

func TestClusterAdcsIssuerSpecDeepCopy(t *testing.T) {
	in := newFullClusterAdcsIssuer().Spec

	out := in.DeepCopy()
	assert.Equal(t, &in, out)
	assert.NotSame(t, &in, out)

	out.CABundle[0] = 'X'
	assert.NotEqual(t, in.CABundle[0], out.CABundle[0])

	var nilSpec *ClusterAdcsIssuerSpec
	assert.Nil(t, nilSpec.DeepCopy())

	emptySpec := ClusterAdcsIssuerSpec{URL: "u"}
	var into ClusterAdcsIssuerSpec
	emptySpec.DeepCopyInto(&into)
	assert.Nil(t, into.CABundle)
}

func TestClusterAdcsIssuerStatusDeepCopy(t *testing.T) {
	in := &ClusterAdcsIssuerStatus{}
	out := in.DeepCopy()
	assert.Equal(t, in, out)

	var nilStatus *ClusterAdcsIssuerStatus
	assert.Nil(t, nilStatus.DeepCopy())
}

func TestClusterAdcsIssuerDeepCopy(t *testing.T) {
	in := newFullClusterAdcsIssuer()

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	out.Spec.CABundle[0] = 'Z'
	assert.NotEqual(t, in.Spec.CABundle[0], out.Spec.CABundle[0])

	out.Labels["foo"] = "changed"
	assert.NotEqual(t, in.Labels["foo"], out.Labels["foo"])

	var nilIssuer *ClusterAdcsIssuer
	assert.Nil(t, nilIssuer.DeepCopy())

	obj := in.DeepCopyObject()
	assert.Equal(t, in, obj)

	var nilIssuerObj *ClusterAdcsIssuer
	assert.Nil(t, nilIssuerObj.DeepCopyObject())
}

func TestClusterAdcsIssuerListDeepCopy(t *testing.T) {
	in := &ClusterAdcsIssuerList{
		TypeMeta: metav1.TypeMeta{Kind: "ClusterAdcsIssuerList"},
		ListMeta: metav1.ListMeta{ResourceVersion: "1"},
		Items:    []ClusterAdcsIssuer{*newFullClusterAdcsIssuer(), *newFullClusterAdcsIssuer()},
	}

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	out.Items[0].Spec.CABundle[0] = 'Y'
	assert.NotEqual(t, in.Items[0].Spec.CABundle[0], out.Items[0].Spec.CABundle[0])

	var nilList *ClusterAdcsIssuerList
	assert.Nil(t, nilList.DeepCopy())

	obj := in.DeepCopyObject()
	assert.Equal(t, in, obj)

	var nilListObj *ClusterAdcsIssuerList
	assert.Nil(t, nilListObj.DeepCopyObject())

	emptyList := &ClusterAdcsIssuerList{}
	emptyOut := emptyList.DeepCopy()
	assert.Nil(t, emptyOut.Items)
}

func TestAdcsRequestSpecDeepCopy(t *testing.T) {
	in := newFullAdcsRequest().Spec

	out := in.DeepCopy()
	assert.Equal(t, &in, out)
	assert.NotSame(t, &in, out)

	out.CSRPEM[0] = 'X'
	assert.NotEqual(t, in.CSRPEM[0], out.CSRPEM[0])

	var nilSpec *AdcsRequestSpec
	assert.Nil(t, nilSpec.DeepCopy())

	// nil CSRPEM branch.
	emptySpec := AdcsRequestSpec{}
	var into AdcsRequestSpec
	emptySpec.DeepCopyInto(&into)
	assert.Nil(t, into.CSRPEM)
}

func TestAdcsRequestStatusDeepCopy(t *testing.T) {
	in := &AdcsRequestStatus{Id: "1", State: Ready, Reason: "done"}
	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	var nilStatus *AdcsRequestStatus
	assert.Nil(t, nilStatus.DeepCopy())
}

func TestAdcsRequestDeepCopy(t *testing.T) {
	in := newFullAdcsRequest()

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	out.Spec.CSRPEM[0] = 'Z'
	assert.NotEqual(t, in.Spec.CSRPEM[0], out.Spec.CSRPEM[0])

	out.Labels["foo"] = "changed"
	assert.NotEqual(t, in.Labels["foo"], out.Labels["foo"])

	var nilRequest *AdcsRequest
	assert.Nil(t, nilRequest.DeepCopy())

	obj := in.DeepCopyObject()
	assert.Equal(t, in, obj)

	var nilRequestObj *AdcsRequest
	assert.Nil(t, nilRequestObj.DeepCopyObject())
}

func TestAdcsRequestListDeepCopy(t *testing.T) {
	in := &AdcsRequestList{
		TypeMeta: metav1.TypeMeta{Kind: "AdcsRequestList"},
		ListMeta: metav1.ListMeta{ResourceVersion: "1"},
		Items:    []AdcsRequest{*newFullAdcsRequest(), *newFullAdcsRequest()},
	}

	out := in.DeepCopy()
	assert.Equal(t, in, out)
	assert.NotSame(t, in, out)

	out.Items[0].Spec.CSRPEM[0] = 'Y'
	assert.NotEqual(t, in.Items[0].Spec.CSRPEM[0], out.Items[0].Spec.CSRPEM[0])

	var nilList *AdcsRequestList
	assert.Nil(t, nilList.DeepCopy())

	obj := in.DeepCopyObject()
	assert.Equal(t, in, obj)

	var nilListObj *AdcsRequestList
	assert.Nil(t, nilListObj.DeepCopyObject())

	emptyList := &AdcsRequestList{}
	emptyOut := emptyList.DeepCopy()
	assert.Nil(t, emptyOut.Items)
}
