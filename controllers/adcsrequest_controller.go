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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	"k8s.io/klog"
	"time"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	api "github.com/nokia/adcs-issuer/api/v1"
	"github.com/nokia/adcs-issuer/issuers"
	corev1 "k8s.io/api/core/v1"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	globals "github.com/nokia/adcs-issuer/globals"
	sig_controller "sigs.k8s.io/controller-runtime/pkg/controller"
)

// default values
var check_interval time.Duration = 5 * time.Minute

//var retry_interval time.Duration = 5 * time.Minute
//var timeout_duration time.Duration = 30 * time.Second

// AdcsRequestReconciler reconciles a AdcsRequest object
type AdcsRequestReconciler struct {
	client.Client
	Log                          logr.Logger
	IssuerFactory                issuers.IssuerFactory
	Recorder                     record.EventRecorder
	CertificateRequestController *CertificateRequestReconciler
	Tracer                       trace.Tracer
}

// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests/status,verbs=get;update;patch

func (r *AdcsRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("adcsrequest", req.NamespacedName)

	ctx, span := r.Tracer.Start(context.Background(), "AdcsRequestReconciler")
	span.AddEvent("AdcsRequestReconciler start",
		trace.WithAttributes(attribute.String("name", req.Name),
			attribute.String("namespace", req.Namespace)))

	defer span.End()

	// your logic here
	log.Info("Processing request")
	if klog.V(3) {
		klog.Infof("requesting to template: %v", r.IssuerFactory.AdcsTemplateName)
	}

	// Fetch the AdcsRequest resource being reconciled
	ar := new(api.AdcsRequest)
	if err := r.Client.Get(ctx, req.NamespacedName, ar); err != nil {
		// We don't log error here as this is probably the 'NotFound'
		// case for deleted object.
		//
		// The Manager will log other errors.

		return ctrl.Result{}, client.IgnoreNotFound(err)

	}

	log.V(3).Info("Running request", "Processing request", req.Name)

	// Find the issuer
	issuer, err := r.IssuerFactory.GetIssuer(ctx, ar.Spec.IssuerRef, ar.Namespace)
	if err != nil {
		log.WithValues("issuer", ar.Spec.IssuerRef).Error(err, "Couldn't get issuer")
		return ctrl.Result{}, err
	}

	span.AddEvent("AdcsRequestReconciler details", trace.WithAttributes(
		attribute.String("CertID", ar.Status.Id),
		attribute.String("State", string(ar.Status.State)),
		attribute.String("Reason", ar.Status.Reason),
		attribute.String("CSR", string(ar.Spec.CSRPEM)),
		attribute.String("name", req.Name),
		attribute.String("namespace", req.Namespace)))

	if log.V(3).Enabled() {
		log.V(3).Info("Running request", "template", issuer.AdcsTemplateName)
	}

	cert, caCert, err := issuer.Issue(ctx, ar)
	if err != nil {
		// This is a local error.
		// We don't change the request status and just put it back on the queue
		// to re-try later.
		log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)

		span.AddEvent("AdcsRequestReconciler Failed request will be re-tried", trace.WithAttributes(
			attribute.String("CertID", ar.Status.Id),
			attribute.String("State", string(ar.Status.State)),
			attribute.String("Reason", ar.Status.Reason),
			attribute.String("CSR", string(ar.Spec.CSRPEM)),
			attribute.String("Error", fmt.Sprint(err)),
			attribute.String("name", req.Name),
			attribute.String("namespace", req.Namespace)))

		span.SetStatus(codes.Error, fmt.Sprint(err))

		return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
	}

	// Get the original CertificateRequest to set result in
	cr, err := r.CertificateRequestController.GetCertificateRequest(ctx, req.NamespacedName)
	if err != nil {
		log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)
		return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
	}

	switch ar.Status.State {
	case api.Pending:
		// Check again later
		log.Info(fmt.Sprintf("Pending request will be re-tried in %v", issuer.StatusCheckInterval))
		err = r.setStatus(ctx, ar)
		if err != nil {
			log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)
			return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
		}
		return ctrl.Result{Requeue: true, RequeueAfter: issuer.StatusCheckInterval}, nil
	case api.Ready:

		// Combine the certificates, as we need the intermediate certs in with the CA.
		combinedCert := cert
		if caCert != nil {
			combinedCert = append(cert, caCert...)
		}
		cr.Status.Certificate = combinedCert

		if log.V(5).Enabled() {
			s := string(cert)
			log.V(5).Info("certificate obtained", "certificate", s)
		}

		// CA cert is inside the cert above
		// cr.Status.CA = caCert
		err = r.CertificateRequestController.SetStatus(ctx, &cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "ADCS request successful")
		if err != nil {
			log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)
			return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
		}

	case api.Rejected:
		// This is a little hack for strange cert-manager behavior in case of failed request. Cert-manager automatically
		// re-tries such requests (re-created CertificateRequest object) what doesn't make sense in case of rejection.
		// We keep the Reason 'Pending' to prevent from re-trying while the actual status is in the Status Condition's Message field.
		// TODO: change it when cert-manager handles this better.
		err = r.CertificateRequestController.SetStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "ADCS request rejected")
		if err != nil {
			log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)
			return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
		}

	case api.Errored:
		err = r.CertificateRequestController.SetStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "ADCS request errored")
		if err != nil {
			log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)

			span.AddEvent("AdcsRequestReconciler Failed request will be re-tried", trace.WithAttributes(
				attribute.String("CertID", ar.Status.Id),
				attribute.String("State", string(ar.Status.State)),
				attribute.String("Reason", ar.Status.Reason),
				attribute.String("CSR", string(ar.Spec.CSRPEM)),
				attribute.String("Error", fmt.Sprint(err)),
				attribute.String("name", req.Name),
				attribute.String("namespace", req.Namespace)))

			span.SetStatus(codes.Error, fmt.Sprint(err))

			return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
		}
	}

	err = r.setStatus(ctx, ar)
	if err != nil {
		log.Error(err, "Failed request will be re-tried", "retry interval", issuer.RetryInterval)
		return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
	}

	span.AddEvent("AdcsRequestReconciler end",
		trace.WithAttributes(attribute.String("name", req.Name),
			attribute.String("namespace", req.Namespace)))

	return ctrl.Result{RequeueAfter: check_interval}, nil //requeue after check_interval
}

func (r *AdcsRequestReconciler) setStatus(ctx context.Context, ar *api.AdcsRequest) error {

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if ar.Status.State == api.Errored || ar.Status.State == api.Rejected {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(ar, eventType, string(ar.Status.State), ar.Status.Reason)

	return r.Client.Status().Update(ctx, ar)
}

func (r *AdcsRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.AdcsRequest{}).
		WithOptions(sig_controller.Options{MaxConcurrentReconciles: globals.MaxConcurrentReconciles}).
		Complete(r)
}

func (r *AdcsRequestReconciler) GetExternalResourcesStatus(ctx context.Context, cr *api.AdcsRequest, caBundle []byte,
	url string, username string, password string, token string, certID int32, serialNumber string, insecureSkipVerify bool, licenseKey string, issuerSpec *api.AdcsIssuerSpec, secret corev1.Secret) (int, error) {
	return 0, nil
}
