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

	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmapiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	api "github.com/djkormo/adcs-issuer/api/v1"
	core "k8s.io/api/core/v1"
	apimacherrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

// AdcsRequestReconciler reconciles a AdcsRequest object
type CertificateRequestReconciler struct {
	client.Client
	Recorder record.EventRecorder

	Clock                  clock.Clock
	CheckApprovedCondition bool
}

var (
	certificateRequestGvk = cmapi.SchemeGroupVersion.WithKind("CertificateRequest")
)

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests/status;certificates/finalizers;certificaterequests/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=patch;create

func (r *CertificateRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "certificaterequest", req.NamespacedName)

	// your logic here

	// Fetch the CertificateRequest resource being reconciled
	cr, err := r.GetCertificateRequest(ctx, req.NamespacedName)
	if err != nil {
		// We don't log error here as this is probably the 'NotFound'
		// case for deleted object. The AdcsRequest will be automatically deleted for cascading delete.
		//
		// The Manager will log other errors.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the CertificateRequest's issuerRef and if it does not match the api
	// group name, log a message at a debug level and stop processing.
	if cr.Spec.IssuerRef.Group != api.GroupVersion.Group {
		klog.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", cr.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}

	// Ignore CertificateRequest if it is already Ready
	if cmapiutil.CertificateRequestHasCondition(&cr, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionTrue,
	}) {
		log.V(4).Info("CertificateRequest is Ready. Ignoring.")
		return ctrl.Result{}, nil
	}
	// Ignore CertificateRequest if it is already Failed
	if cmapiutil.CertificateRequestHasCondition(&cr, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionFalse,
		Reason: cmapi.CertificateRequestReasonFailed,
	}) {
		log.V(4).Info("CertificateRequest is Failed. Ignoring.")
		return ctrl.Result{}, nil
	}
	// Ignore CertificateRequest if it already has a Denied Ready Reason
	if cmapiutil.CertificateRequestHasCondition(&cr, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionFalse,
		Reason: cmapi.CertificateRequestReasonDenied,
	}) {
		log.V(4).Info("CertificateRequest already has a Ready condition with Denied Reason. Ignoring.")
		return ctrl.Result{}, nil
	}

	// If CertificateRequest has been denied, mark the CertificateRequest as
	// Ready=Denied and set FailureTime if not already.
	if cmapiutil.CertificateRequestIsDenied(&cr) {
		log.V(4).Info("CertificateRequest has been denied. Marking as failed.")

		if cr.Status.FailureTime == nil {
			nowTime := metav1.NewTime(r.Clock.Now())
			cr.Status.FailureTime = &nowTime
		}

		message := "The CertificateRequest was denied by an approval controller"
		return ctrl.Result{}, r.SetStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, message)
	}

	if r.CheckApprovedCondition {
		// If CertificateRequest has not been approved, exit early.
		if !cmapiutil.CertificateRequestIsApproved(&cr) {
			log.V(4).Info("certificate request has not been approved")
			return ctrl.Result{}, nil
		}
	}

	// If the certificate data is already set then we skip this request as it
	// has already been completed in the past.
	if len(cr.Status.Certificate) > 0 {
		klog.V(4).Info("existing certificate data found in status, skipping already completed CertificateRequest")
		return ctrl.Result{}, nil
	}

	adcsReq := new(api.AdcsRequest)
	// Check if AdcsRequest with the same name already exists
	err = r.Client.Get(ctx, req.NamespacedName, adcsReq)
	if err != nil {
		if !apimacherrors.IsNotFound(err) {
			log.Error(err, "failed to check for existing AdcsRequest resource")
			return ctrl.Result{}, err
		}
	}

	if err == nil {
		log.Info("AdcsRequest already exists")
		// The ADCS Request already exists. If the CSR is different we delete it and create a new one
		if !RequestDiffers(adcsReq, &cr) {
			log.Info("No change in request")
			return ctrl.Result{}, nil
		}
		log.Info("AdcsRequest differs. Deleting.")
		// NOTE: Remember that CertificateRequest name contains its CSR hash so the name
		// is usually unique for each CSR.
		// So, this is actually a rare case because any change in Certificate will produce
		// new CertificateRequest with different name.
		if err := r.Client.Delete(ctx, adcsReq); err != nil {
			log.Error(err, "failed to delete existing AdcsRequest")
		}
	}

	log.Info("Creating new AdcsRequest")
	err = r.createAdcsRequest(ctx, &cr)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.SetStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "Processing ADCS request")
	if err != nil {
		return ctrl.Result{}, err
	}

	log.V(4).Info("setstatus", "ctx", ctx, "cr", &cr)

	return ctrl.Result{}, nil
}

func (r *CertificateRequestReconciler) createAdcsRequest(ctx context.Context, cmRequest *cmapi.CertificateRequest) error {
	spec := api.AdcsRequestSpec{
		CSRPEM:    cmRequest.Spec.Request,
		IssuerRef: cmRequest.Spec.IssuerRef,
	}
	return r.Create(ctx, &api.AdcsRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cmRequest.Name,
			Namespace:       cmRequest.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cmRequest, certificateRequestGvk)},
		},
		Spec: spec,
	})
}

func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

func RequestDiffers(adcsReq *api.AdcsRequest, certReq *cmapi.CertificateRequest) bool {
	a := adcsReq.Spec.CSRPEM
	b := certReq.Spec.Request
	if len(a) != len(b) {
		return true
	}
	for i, v := range a {
		if v != b[i] {
			return true
		}
	}
	return false
}

func (r *CertificateRequestReconciler) GetCertificateRequest(ctx context.Context, key client.ObjectKey) (cmapi.CertificateRequest, error) {
	cr := new(cmapi.CertificateRequest)
	if err := r.Client.Get(ctx, key, cr); err != nil {
		// We don't log error here as this is probably the 'NotFound'
		// case for deleted object. The AdcsRequest will be automatically deleted for cascading delete.
		//
		// The Manager will log other errors.
		return *cr, err
	}
	return *cr, nil
}

func (r *CertificateRequestReconciler) SetStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmeta.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	cmapiutil.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmeta.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)

	return r.Client.Status().Update(ctx, cr)
}
