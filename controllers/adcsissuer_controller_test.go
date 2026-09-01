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

func adcsIssuerScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	assert.NoError(t, adcsv1.AddToScheme(scheme))
	return scheme
}

func TestAdcsIssuerReconcile_Found(t *testing.T) {
	// arrange
	issuer := &adcsv1.AdcsIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "my-issuer", Namespace: "default"},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(adcsIssuerScheme(t)).
		WithObjects(issuer).
		Build()

	r := &AdcsIssuerReconciler{
		Client: fakeClient,
		Log:    logr.Discard(),
	}

	// act
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-issuer"},
	})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestAdcsIssuerReconcile_NotFound(t *testing.T) {
	// arrange
	fakeClient := fake.NewClientBuilder().
		WithScheme(adcsIssuerScheme(t)).
		Build()

	r := &AdcsIssuerReconciler{
		Client: fakeClient,
		Log:    logr.Discard(),
	}

	// act
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing-issuer"},
	})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}
