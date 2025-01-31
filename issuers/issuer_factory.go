package issuers

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"github.com/djkormo/adcs-issuer/adcs"
	api "github.com/djkormo/adcs-issuer/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultStatusCheckInterval = "6h"
	defaultRetryInterval       = "1h"
)

type IssuerFactory struct {
	client.Client
	ClusterResourceNamespace string
	AdcsTemplateName         string
}

func (f *IssuerFactory) GetIssuer(ctx context.Context, ref cmmeta.ObjectReference, namespace string) (*Issuer, error) {
	key := client.ObjectKey{Namespace: namespace, Name: ref.Name}

	switch strings.ToLower(ref.Kind) {
	case "adcsissuer":
		return f.getAdcsIssuer(ctx, key)
	case "clusteradcsissuer":
		return f.getClusterAdcsIssuer(ctx, key)
	}
	return nil, fmt.Errorf("Unsupported issuer kind %s.", ref.Kind)
}

// Get AdcsIssuer object from K8s and create Issuer
func (f *IssuerFactory) getAdcsIssuer(ctx context.Context, key client.ObjectKey) (*Issuer, error) {

	log := ctrl.LoggerFrom(ctx, "AdcsIssuer", key)

	issuer := new(api.AdcsIssuer)
	if err := f.Client.Get(ctx, key, issuer); err != nil {
		return nil, err
	}
	// TODO: add checking issuer status

	username, password, err := f.getUserPassword(ctx, issuer.Spec.CredentialsRef.Name, issuer.Namespace)
	if err != nil {
		return nil, err
	}

	certs := issuer.Spec.CABundle
	if len(certs) == 0 {
		return nil, fmt.Errorf("CA Bundle required")
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("error loading ADCS CA bundle")
	}

	certServ, err := adcs.NewNtlmCertsrv(issuer.Spec.URL, username, password, caCertPool, false)
	if err != nil {
		return nil, err
	}

	statusCheckInterval := getInterval(
		issuer.Spec.StatusCheckInterval,
		defaultStatusCheckInterval,
		log.WithValues("interval", "statusCheckInterval"))
	retryInterval := getInterval(
		issuer.Spec.RetryInterval,
		defaultRetryInterval,
		log.WithValues("interval", "retryInterval"))
	adcsTemplateName := f.AdcsTemplateName
	if issuer.Spec.TemplateName != "" {
		adcsTemplateName = issuer.Spec.TemplateName
	}
	return &Issuer{
		f.Client,
		certServ,
		retryInterval,
		statusCheckInterval,
		adcsTemplateName,
	}, nil
}

// Get ClusterAdcsIssuer object from K8s and create Issuer
func (f *IssuerFactory) getClusterAdcsIssuer(ctx context.Context, key client.ObjectKey) (*Issuer, error) {
	log := ctrl.LoggerFrom(ctx, "ClusterAdcsIssuer", key)
	key.Namespace = ""

	issuer := new(api.ClusterAdcsIssuer)
	if err := f.Client.Get(ctx, key, issuer); err != nil {
		return nil, err
	}
	// TODO: add checking issuer status

	username, password, err := f.getUserPassword(ctx, issuer.Spec.CredentialsRef.Name, f.ClusterResourceNamespace)
	if err != nil {
		return nil, err
	}

	certs := issuer.Spec.CABundle
	if len(certs) == 0 {
		return nil, fmt.Errorf("CA Bundle required")
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("error loading ADCS CA bundle")
	}

	certServ, err := adcs.NewNtlmCertsrv(issuer.Spec.URL, username, password, caCertPool, false)
	if err != nil {
		return nil, err
	}

	statusCheckInterval := getInterval(
		issuer.Spec.StatusCheckInterval,
		defaultStatusCheckInterval,
		log.WithValues("interval", "statusCheckInterval"))
	retryInterval := getInterval(
		issuer.Spec.RetryInterval,
		defaultRetryInterval,
		log.WithValues("interval", "retryInterval"))
	adcsTemplateName := f.AdcsTemplateName
	if issuer.Spec.TemplateName != "" {
		adcsTemplateName = issuer.Spec.TemplateName
	}
	return &Issuer{
		f.Client,
		certServ,
		retryInterval,
		statusCheckInterval,
		adcsTemplateName,
	}, nil
}

func getInterval(specValue string, def string, log logr.Logger) time.Duration {
	interval, _ := time.ParseDuration(def)
	if specValue != "" {
		i, err := time.ParseDuration(specValue)
		if err != nil {
			log.Error(err, "Cannot parse interval. Using default.")
		} else {
			interval = i
		}
	} else {
		log.Info("Using default")
	}
	return interval
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (f *IssuerFactory) getUserPassword(ctx context.Context, secretName string, namespace string) (string, string, error) {
	secret := new(corev1.Secret)
	if err := f.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, secret); err != nil {
		return "", "", err
	}
	if _, ok := secret.Data["username"]; !ok {
		return "", "", fmt.Errorf("User name not set in secret")
	}
	if _, ok := secret.Data["password"]; !ok {
		return "", "", fmt.Errorf("Password not set in secret")

	}

	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}
