package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	adcsv1 "github.com/djkormo/adcs-issuer/api/v1"
)

func clusterAdcsIssuerScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	assert.NoError(t, adcsv1.AddToScheme(scheme))
	return scheme
}

func TestClusterAdcsIssuerReconcile_Found(t *testing.T) {
	// arrange
	issuer := &adcsv1.ClusterAdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-cluster-issuer"},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(clusterAdcsIssuerScheme(t)).
		WithObjects(issuer).
		Build()

	r := &ClusterAdcsIssuerReconciler{
		Client: fakeClient,
		Log:    logr.Discard(),
	}

	// act
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "my-cluster-issuer"},
	})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestClusterAdcsIssuerReconcile_NotFound(t *testing.T) {
	// arrange
	fakeClient := fake.NewClientBuilder().
		WithScheme(clusterAdcsIssuerScheme(t)).
		Build()

	r := &ClusterAdcsIssuerReconciler{
		Client: fakeClient,
		Log:    logr.Discard(),
	}

	// act
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "missing-cluster-issuer"},
	})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}
