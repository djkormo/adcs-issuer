package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	adcsv1 "github.com/djkormo/adcs-issuer/api/v1"
)

// AdcsIssuerReconciler reconciles a AdcsIssuer object
type AdcsIssuerReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsissuers/status,verbs=get;update;patch

func (r *AdcsIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("adcsissuer", req.NamespacedName)

	// your logic here

	// Fetch the AdcsIssuer resource being reconciled
	issuer := new(adcsv1.AdcsIssuer)
	if err := r.Client.Get(ctx, req.NamespacedName, issuer); err != nil {
		// We don't log error here as this is probably the 'NotFound'
		// case for deleted object. The AdcsRequest will be automatically deleted for cascading delete.
		//
		// The Manager will log other errors.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Registered issuer")

	return ctrl.Result{}, nil
}

func (r *AdcsIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&adcsv1.AdcsIssuer{}).
		Complete(r)
}
