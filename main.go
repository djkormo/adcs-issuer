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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	adcsv1 "github.com/djkormo/adcs-issuer/api/v1"
	"github.com/djkormo/adcs-issuer/controllers"
	"github.com/djkormo/adcs-issuer/healthcheck"
	"github.com/djkormo/adcs-issuer/issuers"
	"github.com/djkormo/adcs-issuer/version"

	zaplogfmt "github.com/sykesm/zap-logfmt"
	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/utils/clock"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

const (
	defaultWebhooksPort    int = 9443
	defaultMetricsPort     int = 8080 // TODO hardcoded port
	defaultHealthCheckPort int = 8081 // TODO hardcoded port
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = certmanager.AddToScheme(scheme)
	_ = adcsv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var healthcheckAddr string
	var webhooksPort string
	var enableLeaderElection bool
	var clusterResourceNamespace string
	var disableApprovedCheck bool
	var adcsTemplateName string

	flag.StringVar(&metricsAddr, "metrics-bind-address", fmt.Sprintf(":%d", defaultMetricsPort), "The address the metric endpoint binds to.")
	flag.StringVar(&healthcheckAddr, "healthcheck-addr", fmt.Sprintf(":%d", defaultHealthCheckPort), "The address the healthcheck endpoints binds to.")
	flag.StringVar(&webhooksPort, "webhooks-port", strconv.Itoa(defaultWebhooksPort), "Port for webhooks requests.")
	flag.BoolVar(&disableApprovedCheck, "disable-approved-check", false,
		"Disables waiting for CertificateRequests to have an approved condition before signing.")

	/* 	port, err := strconv.Atoi(webhooksPort)
	   	if err != nil {
	   		setupLog.Error(err, "invalid webhooks port. Using default.")
	   		port = defaultWebhooksPort
	   	} */
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&clusterResourceNamespace, "cluster-resource-namespace", "kube-system", "Namespace where cluster-level resources are stored.")
	flag.StringVar(&adcsTemplateName, "adcsTemplateName", "BasicSSLWebServer", "Name of ADCS Template.")

	// Options for configuring logging
	opts := zap.Options{
		Development: false, //was true
	}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()

	// based on https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/

	configLog := uzap.NewProductionEncoderConfig()
	// changing  time format to RFC3339Nano -> 2006-01-02T15:04:05.999999999Z07:00"
	configLog.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format(time.RFC3339Nano))
	}
	logfmtEncoder := zaplogfmt.NewEncoder(configLog)

	// Construct a new logr.logger.
	logger := zap.New(zap.UseDevMode(false), zap.WriteTo(os.Stdout), zap.Encoder(logfmtEncoder))
	ctrl.SetLogger(logger)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Starting ADCS Issuer", "Version", version.Version, "BuildTime", version.BuildTime, "Release", version.Release, "Commit", version.Commit)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		//MetricsBindAddress:     metricsAddr, //unknown field MetricsBindAddress in struct literal of type manager.Options
		HealthProbeBindAddress: healthcheckAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "adcs-issuer-controller",
		//Port:                   port, //unknown field Port in struct literal of type manager.Options
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = mgr.AddHealthzCheck("healthz", healthcheck.HealthCheck)
	if err != nil {
		setupLog.Error(err, "unable to start AddHealthzCheck")
		os.Exit(1)
	}

	err = mgr.AddReadyzCheck("readyz", healthcheck.HealthCheck)
	if err != nil {
		setupLog.Error(err, "unable to start AddReadyzCheck")
		os.Exit(1)
	}

	certificateRequestReconciler := &controllers.CertificateRequestReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor("adcs-certificaterequests-controller"),

		Clock:                  clock.RealClock{},
		CheckApprovedCondition: !disableApprovedCheck,
	}
	if err = (certificateRequestReconciler).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertificateRequest")
		os.Exit(1)
	}

	if err = (&controllers.AdcsRequestReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AdcsRequest"),
		IssuerFactory: issuers.IssuerFactory{
			Client:                   mgr.GetClient(),
			ClusterResourceNamespace: clusterResourceNamespace,
			AdcsTemplateName:         adcsTemplateName,
		},
		Recorder:                     mgr.GetEventRecorderFor("adcs-requests-controller"),
		CertificateRequestController: certificateRequestReconciler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AdcsRequest")
		os.Exit(1)
	}

	if err = (&controllers.AdcsIssuerReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AdcsIssuer"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AdcsIssuer")
		os.Exit(1)
	}

	/* 	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
	   		if err = (&adcsv1.AdcsIssuer{}).SetupWebhookWithManager(mgr); err != nil {
	   			setupLog.Error(err, "unable to create webhook", "webhook", "AdcsIssuer")
	   			os.Exit(1)
	   		}
	   	}
	*/
	if err = (&controllers.ClusterAdcsIssuerReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ClusterAdcsIssuer"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterAdcsIssuer")
		os.Exit(1)
	}
	/* 	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&adcsv1.ClusterAdcsIssuer{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "ClusterAdcsIssuer")
			os.Exit(1)
		}
	} */
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
