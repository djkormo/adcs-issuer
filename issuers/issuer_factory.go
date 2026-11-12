package issuers

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
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
	return nil, fmt.Errorf("unsupported issuer kind %s", ref.Kind)
}

// Get AdcsIssuer object from K8s and create Issuer
func (f *IssuerFactory) getAdcsIssuer(ctx context.Context, key client.ObjectKey) (*Issuer, error) {

	log := ctrl.LoggerFrom(ctx, "AdcsIssuer", key)

	issuer := new(api.AdcsIssuer)
	if err := f.Get(ctx, key, issuer); err != nil {
		return nil, err
	}
	// TODO: add checking issuer status

	username, password, realm, pwError := f.getUserPassword(ctx, issuer.Spec.CredentialsRef.Name, issuer.Namespace)
	if pwError != nil {
		return nil, pwError
	}

	certs, err := f.getCaBundle(ctx, issuer.Spec.CABundleRef, issuer.Namespace)
	if err != nil {
		if len(issuer.Spec.CABundle) == 0 {
			return nil, err
		}
		log.Error(err, "Cannot load caBundleRef. Using inline caBundle.")
	}
	if len(certs) == 0 {
		certs = issuer.Spec.CABundle
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("CA Bundle required")
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("error loading ADCS CA bundle")
	}

	authMode := os.Getenv("ADCS_AUTH_MODE")
	var (
		certServ adcs.AdcsCertsrv
		adcsErr  error
	)

	// Choose method based on env
	if authMode == "kerberos" {
		certServ, adcsErr = adcs.NewKerberosCertsrv(issuer.Spec.URL, username, realm, password, caCertPool, false)
	} else { // default is ntlm
		certServ, adcsErr = adcs.NewNtlmCertsrv(issuer.Spec.URL, username, password, caCertPool, false)
	}
	if adcsErr != nil {
		return nil, adcsErr
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
	if err := f.Get(ctx, key, issuer); err != nil {
		return nil, err
	}
	// TODO: add checking issuer status

	username, password, realm, pwError := f.getUserPassword(ctx, issuer.Spec.CredentialsRef.Name, f.ClusterResourceNamespace)
	if pwError != nil {
		return nil, pwError
	}

	caBundleNamespace := f.ClusterResourceNamespace
	if issuer.Spec.CABundleRef.Namespace != "" {
		caBundleNamespace = issuer.Spec.CABundleRef.Namespace
	}
	certs, err := f.getCaBundle(ctx, issuer.Spec.CABundleRef.CABundleReference, caBundleNamespace)
	if err != nil {
		if len(issuer.Spec.CABundle) == 0 {
			return nil, err
		}
		log.Error(err, "Cannot load caBundleRef. Using inline caBundle.")
	}
	if len(certs) == 0 {
		certs = issuer.Spec.CABundle
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("CA Bundle required")
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("error loading ADCS CA bundle")
	}

	authMode := os.Getenv("ADCS_AUTH_MODE")
	var (
		certServ adcs.AdcsCertsrv
		adcsErr  error
	)

	// Choose method based on env
	if authMode == "kerberos" {
		certServ, adcsErr = adcs.NewKerberosCertsrv(issuer.Spec.URL, username, realm, password, caCertPool, false)
	} else { // default is ntlm
		certServ, adcsErr = adcs.NewNtlmCertsrv(issuer.Spec.URL, username, password, caCertPool, false)
	}
	if adcsErr != nil {
		return nil, adcsErr
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
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

func (f *IssuerFactory) getUserPassword(ctx context.Context, secretName string, namespace string) (string, string, string, error) {
	secret := new(corev1.Secret)
	if err := f.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, secret); err != nil {
		return "", "", "", err
	}
	if _, ok := secret.Data["username"]; !ok {
		return "", "", "", fmt.Errorf("user name not set in secret")
	}
	if _, ok := secret.Data["password"]; !ok {
		return "", "", "", fmt.Errorf("password not set in secret")
	}

	authMode := os.Getenv("ADCS_AUTH_MODE")
	if _, ok := secret.Data["realm"]; !ok && authMode == "kerberos" {
		return "", "", "", fmt.Errorf("realm not set in secret")
	}

	return string(secret.Data["username"]), string(secret.Data["password"]), string(secret.Data["realm"]), nil
}

func (f *IssuerFactory) getCaBundle(ctx context.Context, ref api.CABundleReference, namespace string) ([]byte, error) {
	if ref == (api.CABundleReference{}) {
		return nil, nil
	}
	if ref.Name == "" {
		return nil, fmt.Errorf("caBundleRef.name is required")
	}
	key := ref.Key
	if key == "" {
		key = "ca.crt"
	}
	kind := ref.Kind
	if kind == "" {
		kind = "Secret"
	}
	objectKey := client.ObjectKey{Namespace: namespace, Name: ref.Name}

	var bundle []byte
	switch kind {
	case "Secret":
		secret := new(corev1.Secret)
		if err := f.Get(ctx, objectKey, secret); err != nil {
			return nil, fmt.Errorf("loading caBundleRef Secret %s: %w", objectKey, err)
		}
		bundle = secret.Data[key]
	case "ConfigMap":
		configMap := new(corev1.ConfigMap)
		if err := f.Get(ctx, objectKey, configMap); err != nil {
			return nil, fmt.Errorf("loading caBundleRef ConfigMap %s: %w", objectKey, err)
		}
		if value, ok := configMap.Data[key]; ok {
			bundle = []byte(value)
		} else {
			bundle = configMap.BinaryData[key]
		}
	default:
		return nil, fmt.Errorf("unsupported caBundleRef kind %q: expected Secret or ConfigMap", kind)
	}
	if len(bundle) == 0 {
		return nil, fmt.Errorf("caBundleRef key %q is missing or empty in %s %s", key, kind, objectKey)
	}
	return bundle, nil
}
