- [ADCS Issuer](#adcs-issuer)
  - [Description](#description)
    - [Requirements](#requirements)
  - [Installation](#installation)
    - [install cert-manager](#install-cert-manager)
    - [Install  adcs-issuer via helm chart](#install--adcs-issuer-via-helm-chart)
      - [Helm chart parameters](#helm-chart-parameters)
      - [Values](#values)
    - [Prepare your kubernetes resources:](#prepare-your-kubernetes-resources)
      - [Credentials](#credentials)
      - [Issuers](#issuers)
      - [Certificates](#certificates)
      - [Ingresses](#ingresses)
      - [Working with operator](#working-with-operator)
  - [Using adcs simulator](#using-adcs-simulator)
  - [License](#license)


# ADCS Issuer

ADCS Issuer is a [cert-manager's](https://github.com/cert-manager/cert-manager) CertificateRequest controller that uses MS Active Directory Certificate Service to sign certificates 
(see [this design document](https://github.com/cert-manager/cert-manager/blob/master/design/20190708.certificate-request-crd.md) for details on CertificateRequest CRD). 

ADCS provides HTTP GUI that can be normally used to request new certificates or see status of existing requests. 
This implementation is simply a HTTP client that interacts with the ADCS server sending appropriately prepared HTTP requests and interpretting the server's HTTP responses
(the approach inspired by [this Python ADCS client](https://github.com/magnuswatn/certsrv)).

It supports NTLM authentication.



Build statuses:


[![operator pipeline](https://github.com/djkormo/adcs-issuer/actions/workflows/pipeline.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/pipeline.yaml)


[![Code scanning - action](https://github.com/djkormo/adcs-issuer/actions/workflows/codeql.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/codeql.yaml)


[![Publish Docker image on Release](https://github.com/djkormo/adcs-issuer/actions/workflows/main.yml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/main.yml)


[![Release helm charts](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-chart-releaser.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-chart-releaser.yaml)


## Description

### Requirements
In the past ADCS Issuer has been tested with cert-manager v1.9.x and v.12.x and currently supports CertificateRequest CRD API version v1 only.

### NOTE! At present only supported cert-manager versions are supported by adcs-issuer maintaners.
https://cert-manager.io/docs/releases/

<img width="1473" height="357" alt="obraz" src="https://github.com/user-attachments/assets/e76c43eb-a4ff-4734-b9ba-6e56e9ed5980" />


## Installation

This controller is implemented using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder). Automatically generated Makefile contains targets needed for build and installation. 
Generated CRD manifests are stored in `config/crd`. RBAC roles and bindings can be found in config/rbac. There's also a Make target to build controller's Docker image and
store it in local docker repo (Docker must be installed).


### install cert-manager 
```console
helm repo add jetstack https://charts.jetstack.io --force-update
```
```console
helm repo update
```
```console
helm search repo cert-manager
helm search repo cert-manager --versions | grep v1.
```
```console
helm upgrade  --install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.18.3  \
  --set config.enableGatewayAPI=false \
  --set config.apiVersion="controller.config.cert-manager.io/v1alpha1" \
  --set config.kind="ControllerConfiguration" \
  --set crds.enabled=true \
  --set enableCertificateOwnerRef=false

```


### Install  adcs-issuer via helm chart


Using helm chart repo via github repo

```bash
# add helm repo
helm repo add djkormo-adcs-issuer https://djkormo.github.io/adcs-issuer/

# update 
helm repo update djkormo-adcs-issuer

# check all versions 
helm search repo adcs-issuer  --versions

# download values file for some version
helm show values djkormo-adcs-issuer/adcs-issuer --version 2.2.0 > values.yaml

# test installation
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.2.0 \
  --namespace adcs-issuer --values values.yaml  --dry-run

#  install
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.2.2 \
  --namespace adcs-issuer --values values.yaml  --dry-run

# upgrade
helm upgrade project-operator djkormo-adcs-issuer/adcs-issuer  --version 2.2.2 \
  --namespace adcs-issuer --values values.yaml

# uninstall 
helm uninstall adcs-issuer  --namespace  adcs-issuer

```


#### Helm chart parameters

## Chart Overview

ADCS Issuer plugin for cert-manager.

### Chart Details

- **Chart Name:** adcs-issuer
- **Version:** ![Version: 2.2.0](https://img.shields.io/badge/Version-2.2.0-informational?style=flat-square)
- **App Version:** ![AppVersion: 2.2.0](https://img.shields.io/badge/AppVersion-2.2.0-informational?style=flat-square)
- **Chart Type:** ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

**Homepage:** <https://github.com/djkormo/adcs-issuer>

### Source Code

* <https://github.com/djkormo/adcs-issuer>
* <https://djkormo.github.io/adcs-issuer/>

### Requirements

Kubernetes: `>=1.27.0-0`

### Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| controllerManager.affinity.nodeAffinity | object | `{}` |  |
| controllerManager.affinity.podAffinity | object | `{}` |  |
| controllerManager.affinity.podAntiAffinity | object | `{}` |  |
| controllerManager.arguments.cluster-resource-namespace | string | `"adcs-issuer"` |  |
| controllerManager.arguments.disable-approved-check | string | `"false"` |  |
| controllerManager.arguments.enable-leader-election | string | `"true"` |  |
| controllerManager.arguments.zap-log-level | int | `5` |  |
| controllerManager.caCertsSecretName | string | `"ca-certificates"` |  |
| controllerManager.enabledCaCerts | bool | `false` |  |
| controllerManager.enabledWebHooks | bool | `false` |  |
| controllerManager.environment.ENABLE_DEBUG | string | `"false"` |  |
| controllerManager.environment.ENABLE_WEBHOOKS | string | `"false"` |  |
| controllerManager.environment.KUBERNETES_CLUSTER_DOMAIN | string | `"cluster.local"` |  |
| controllerManager.kerberosAuthentication.enabled | bool | `false` |  |
| controllerManager.kerberosAuthentication.krb5Config | string | `"[libdefaults]\n  default_realm = EXAMPLE.COM\n  dns_lookup_kdc = true\n\n[realms]\n  EXAMPLE.COM  = {\n    kdc = dc01.example.com\n  }\n\n[domain_realm]\n  .example.com = EXAMPLE.COM\n  example.com = EXAMPLE.COM\n"` |  |
| controllerManager.manager.image.repository | string | `"djkormo/adcs-issuer"` |  |
| controllerManager.manager.image.tag | string | `"2.2.0"` |  |
| controllerManager.manager.livenessProbe.httpGet.path | string | `"/healthz"` |  |
| controllerManager.manager.livenessProbe.httpGet.port | int | `8081` |  |
| controllerManager.manager.livenessProbe.httpGet.scheme | string | `"HTTP"` |  |
| controllerManager.manager.livenessProbe.periodSeconds | int | `10` |  |
| controllerManager.manager.livenessProbe.timeoutSeconds | int | `10` |  |
| controllerManager.manager.readinessProbe.httpGet.path | string | `"/readyz"` |  |
| controllerManager.manager.readinessProbe.httpGet.port | int | `8081` |  |
| controllerManager.manager.readinessProbe.httpGet.scheme | string | `"HTTP"` |  |
| controllerManager.manager.readinessProbe.initialDelaySeconds | int | `10` |  |
| controllerManager.manager.readinessProbe.periodSeconds | int | `20` |  |
| controllerManager.manager.readinessProbe.timeoutSeconds | int | `20` |  |
| controllerManager.manager.resources.limits.cpu | string | `"100m"` |  |
| controllerManager.manager.resources.limits.memory | string | `"500Mi"` |  |
| controllerManager.manager.resources.requests.cpu | string | `"100m"` |  |
| controllerManager.manager.resources.requests.memory | string | `"120Mi"` |  |
| controllerManager.rbac.certManagerNamespace | string | `"cert-manager"` |  |
| controllerManager.rbac.certManagerServiceAccountName | string | `"cert-manager"` |  |
| controllerManager.rbac.enabled | bool | `true` |  |
| controllerManager.rbac.serviceAccountName | string | `"adcs-issuer"` |  |
| controllerManager.replicas | int | `1` |  |
| controllerManager.securityContext.runAsUser | int | `1000` |  |
| crd.install | bool | `true` |  |
| metricsService.enabled | bool | `true` |  |
| metricsService.nameOverride | string | `nil` |  |
| metricsService.ports[0].name | string | `"http"` |  |
| metricsService.ports[0].port | int | `8080` |  |
| metricsService.ports[0].targetPort | string | `"metrics"` |  |
| metricsService.serviceMonitor.enabled | bool | `true` |  |
| metricsService.serviceMonitor.scheme | string | `"http"` |  |
| metricsService.type | string | `"ClusterIP"` |  |
| nodeSelector | object | `{}` |  |
| simulator.affinity.nodeAffinity | object | `{}` |  |
| simulator.affinity.podAffinity | object | `{}` |  |
| simulator.affinity.podAntiAffinity | object | `{}` |  |
| simulator.arguments.dns | string | `"adcs-sim-service.adcs-issuer.svc,adcs2.example.com"` |  |
| simulator.arguments.ips | string | `"10.10.10.1,10.10.10.2"` |  |
| simulator.arguments.port | int | `8443` |  |
| simulator.caBundle | string | `""` |  |
| simulator.certificateDuration | string | `"2160h"` |  |
| simulator.certificateRenewBefore | string | `"360h"` |  |
| simulator.clusterIssuserName | string | `"adcs-sim-adcsclusterissuer"` |  |
| simulator.configMapName | string | `"adcs-sim-configmap"` |  |
| simulator.containerPort | int | `8443` |  |
| simulator.containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| simulator.containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| simulator.containerSecurityContext.readOnlyRootFilesystem | bool | `true` |  |
| simulator.deploymentName | string | `"adcs-sim-deployment"` |  |
| simulator.enabled | bool | `false` |  |
| simulator.environment.ENABLE_DEBUG | string | `"false"` |  |
| simulator.exampleCertificate.enabled | bool | `true` |  |
| simulator.exampleCertificate.name | string | `"adcs-sim-certificate"` |  |
| simulator.image.repository | string | `"djkormo/adcs-sim"` |  |
| simulator.image.tag | string | `"0.0.6"` |  |
| simulator.issuerGroup | string | `"cert-manager.io"` |  |
| simulator.issuerKind | string | `"Issuer"` |  |
| simulator.issuerName | string | `"adcs-sim-issuer"` |  |
| simulator.livenessProbe.httpGet.path | string | `"/healthz"` |  |
| simulator.livenessProbe.httpGet.port | int | `8443` |  |
| simulator.livenessProbe.httpGet.scheme | string | `"HTTPS"` |  |
| simulator.livenessProbe.periodSeconds | int | `10` |  |
| simulator.livenessProbe.timeoutSeconds | int | `10` |  |
| simulator.podSecurityContext.runAsUser | int | `1000` |  |
| simulator.readinessProbe.httpGet.path | string | `"/readyz"` |  |
| simulator.readinessProbe.httpGet.port | int | `8443` |  |
| simulator.readinessProbe.httpGet.scheme | string | `"HTTPS"` |  |
| simulator.readinessProbe.initialDelaySeconds | int | `10` |  |
| simulator.readinessProbe.periodSeconds | int | `20` |  |
| simulator.readinessProbe.timeoutSeconds | int | `20` |  |
| simulator.resources.limits.cpu | string | `"100m"` |  |
| simulator.resources.limits.memory | string | `"500Mi"` |  |
| simulator.resources.requests.cpu | string | `"100m"` |  |
| simulator.resources.requests.memory | string | `"100Mi"` |  |
| simulator.secretCertificateName | string | `"adcs-sim-certificate-secret"` |  |
| simulator.secretName | string | `"adcs-sim-secret"` |  |
| simulator.serviceName | string | `"adcs-sim-service"` |  |
| simulator.servicePort | int | `8443` |  |
| webhookService.ports[0].port | int | `443` |  |
| webhookService.ports[0].targetPort | int | `9443` |  |
| webhookService.type | string | `"ClusterIP"` |  |


Example of values.yaml file for version 2.0.10 and above

```yaml
crd:
  install: true

controllerManager:
  manager:
    image:
      repository: djkormo/adcs-issuer
      tag: 2.0.10
    resources:
      limits:
        cpu: 100m
        memory: 500Mi
      requests:
        cpu: 100m
        memory: 100Mi

  rbac:
    enabled: true
    serviceAccountName: adcs-issuer
    certManagerNamespace: cert-manager
    certManagerServiceAccountName: cert-manager 


  replicas: 1

  environment:
    KUBERNETES_CLUSTER_DOMAIN: cluster.local
    ENABLE_WEBHOOKS: "false"
    ENABLE_DEBUG: "false"
  arguments:
    enable-leader-election: "true"
    cluster-resource-namespace: adcs-issuer # must be the same as chart namespace
    zap-log-level: 5
    disable-approved-check: "false"

  securityContext:
    runAsUser: 1000

  enabledWebHooks: false
  enabledCaCerts: false
  caCertsSecretName: ca-certificates
metricsService:
  enabled: true
  ports:
  - name: https
    port: 8443
    targetPort: https
  type: ClusterIP
webhookService:
  ports:
  - port: 443
    targetPort: 9443
  type: ClusterIP

simulator:
  enabled: true
  clusterIssuserName: adcs-sim-adcsclusterissuer
  deploymentName: adcs-sim-deployment
  configMapName: adcs-sim-configmap
  secretCertificateName: adcs-sim-certificate-secret
  secretName: adcs-sim-secret
  serviceName: adcs-sim-service 
  image:
    repository: djkormo/adcs-sim
    tag: 0.0.6

  environment:
    ENABLE_DEBUG: "false"

  arguments:
      dns: adcs-sim-service.adcs-issuer.svc,adcs2.example.com 
      ips: 10.10.10.1,10.10.10.2
      port: 8443

  containerPort: 8443
  servicePort: 8443

  livenessProbe:
    httpGet:
      path: /healthz
      port: 8443 # the same as containerPort
      scheme: HTTPS
    timeoutSeconds: 10
    periodSeconds: 10
  readinessProbe:
    httpGet:
      path: /readyz
      port: 8443 # the same as containerPort
      scheme: HTTPS
    timeoutSeconds: 20
    periodSeconds: 20
    initialDelaySeconds: 10

  podSecurityContext:
    runAsUser: 1000

  containerSecurityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    capabilities:
      drop:
      - all

  resources:

    limits:
      cpu: 100m
      memory: 500Mi
    requests:
      cpu: 100m
      memory: 100Mi  

  exampleCertificate:
    enabled: true 
    name: adcs-sim-certificate    
```


Example of values.yaml file for version 2.1.1 and above

```yaml
crd:
  install: true

# ADCS Issuer 

controllerManager:

  manager:
    image:
      repository: djkormo/adcs-issuer
      tag: 2.1.1
    resources:
      limits:
        cpu: 100m
        memory: 500Mi
      requests:
        cpu: 100m
        memory: 100Mi

    livenessProbe:
      httpGet:
        path: /healthz
        port: 8081
        scheme: HTTP
      timeoutSeconds: 10
      periodSeconds: 10

    readinessProbe:
      httpGet:
        path: /readyz
        port: 8081 
        scheme: HTTP
      timeoutSeconds: 20
      periodSeconds: 20
      initialDelaySeconds: 10

  rbac:
    enabled: true
    serviceAccountName: adcs-issuer
    certManagerNamespace: cert-manager
    certManagerServiceAccountName: cert-manager 


  replicas: 1

  environment:
    KUBERNETES_CLUSTER_DOMAIN: cluster.local
    ENABLE_WEBHOOKS: "false"
    ENABLE_DEBUG: "false"
  arguments:
    enable-leader-election: "true"
    cluster-resource-namespace: adcs-issuer # must be the same as chart namespace
    zap-log-level: 5
    disable-approved-check: "false"

  securityContext:
    runAsUser: 1000


  enabledWebHooks: false
  enabledCaCerts: false
  caCertsSecretName: ca-certificates
metricsService:
  enabled: true
  ports:
  - name: https
    port: 8443
    targetPort: https
  type: ClusterIP

webhookService:
  ports:
  - port: 443
    targetPort: 9443
  type: ClusterIP

nodeSelector: {}

# ADCS Simulator 

simulator:
  enabled: false
  clusterIssuserName: adcs-sim-adcsclusterissuer
  deploymentName: adcs-sim-deployment
  configMapName: adcs-sim-configmap
  secretCertificateName: adcs-sim-certificate-secret
  secretName: adcs-sim-secret
  serviceName: adcs-sim-service 
  image:
    repository: djkormo/adcs-simulator
    tag: 0.0.8

  environment:
    ENABLE_DEBUG: "false"

  arguments:
      dns: adcs-sim-service.adcs-issuer.svc,adcs2.example.com 
      ips: 10.10.10.1,10.10.10.2
      port: 8443

  containerPort: 8443
  servicePort: 8443

  livenessProbe:
    httpGet:
      path: /healthz
      port: 8443 # the same as containerPort
      scheme: HTTPS
    timeoutSeconds: 10
    periodSeconds: 10
  readinessProbe:
    httpGet:
      path: /readyz
      port: 8443 # the same as containerPort
      scheme: HTTPS
    timeoutSeconds: 20
    periodSeconds: 20
    initialDelaySeconds: 10

  podSecurityContext:
    runAsUser: 1000

  containerSecurityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    capabilities:
      drop:
      - all

  resources:

    limits:
      cpu: 100m
      memory: 500Mi
    requests:
      cpu: 100m
      memory: 100Mi  

  exampleCertificate:
    enabled: true 
    name: adcs-sim-certificate    


```


### Prepare your kubernetes resources:

#### Credentials

Create credentials for adcs

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: adcs-issuer-credentials
  namespace: adcs-issuer # namespace of  adcs operator
type: Opaque
data:
  password: REDACTED # Password
  username: REDACTED # username

```


#### Issuers

The ADCS service data can be configured in `AdcsIssuer` or `ClusterAdcsIssuer` CRD objects e.g.:


Deploy namespaced  object

```yaml
apiVersion: adcs.certmanager.csf.nokia.com/v1
kind: AdcsIssuer
metadata:
  name: argocd-adcs-issuer
  namespace: argocd # namespace with ingress objects
spec:
  caBundle: REDACTED # ca certificate
  credentialsRef:
    name: adcs-issuer-credentials # reference to kubernetes secret
  statusCheckInterval: 5m
  retryInterval: 5m
  url: https://adcs-host/ # external host
  templateName: adcsTemplate # external template 
```

or

Deploy cluster scoped object

```yaml
apiVersion: adcs.certmanager.csf.nokia.com/v1
kind: ClusterAdcsIssuer
metadata:
  name: adcs-cluster-issuer
spec:
  caBundle: REDACTED # ca certificate
  credentialsRef:
    name: adcs-issuer-credentials # secret with username and password
  statusCheckInterval: 5m
  retryInterval: 5m
  url: https://adcs-host/ # external host
  templateName: adcsTemplate # external template 

```


The `caBundle` parameter is BASE64-encoded CA certificate which is used by the ADCS server itself, which may not be the same certificate that will be used to sign your request.

The `statusCheckInterval` indicates how often the status of the request should be tested. Typically, it can take a few hours or even days before the certificate is issued.

The `retryInterval` says how long to wait before retrying requests that errored.

The `credentialsRef.name` is name of a secret that stores user credentials used for NTLM authentication. The secret must be `Opaque` and contain `password` and `username` fields only e.g.:


The secret used by the `ClusterAdcsIssuer` to authenticate (`credentialsRef`), must be defined in the namespace where the controller's pod is running, or in the namespace specified by the flag  `-clusterResourceNamespace` (default: `kube-system`).



To request a certificate with `AdcsIssuer` the standard `certificate.cert-manager.io` object needs to be created. The `issuerRef` must be set to point to `AdcsIssuer` or `ClusterAdcsIssuer` object
from group `adcs.certmanager.csf.nokie.com` e.g.:


#### Certificates

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  annotations:
  name: adcs-cert
  namespace: <namespace>
spec:
  commonName: example.com
  dnsNames:
  - service1.example.com
  - service2.example.com
  issuerRef:
    group: adcs.certmanager.csf.nokia.com
    kind: AdcsIssuer
    name: test-adcs
  organization:
  - Your organization
  secretName: adcs-cert
```


Cert-manager is responsible for creating the `Secret` with a key and `CertificateRequest` with proper CSR data.


ADCS Issuer creates `AdcsRequest` CRD object that keep actual state of the processing. Its name is always the same as the corresponding `CertificateRequest` object (there is strict one-to-one mapping).
The `AdcsRequest` object stores the ID of request assigned by the ADCS server as wall as the current status which can be one of:
* **Pending** - the request has been sent to ADCS and is waiting for acceptance (status will be checked periodically),
* **Ready** - the request has been successfully processed and the certificate is ready and stored in secret defined in the original `Certificate` object,
* **Rejected** - the request was rejected by ADCS and will be re-tried unless the `Certificate` is updated,
* **Errored**  - unrecoverable problem occured.


```yaml
apiVersion: adcs.certmanager.csf.nokia.com/v1
kind: AdcsRequest
metadata:
  name: adcs-cert-3831834799
  namespace: c1
  ownerReferences:
  - apiVersion: cert-manager.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: CertificateRequest
    name: adcs-cert-3831834799 # the same as AdcsRequest name
    uid: REDACTED
  uid: REDACTED
spec:
  csr: REDACTED # base 64 encoded
  issuerRef:
    group: adcs.certmanager.csf.nokia.com
    kind: AdcsIssuer
    name: test-adcs
status:
  id: "18"
  state: ready
```


#### Ingresses

Using ingress objects

Add annotations for ingress object, here for argocd server


```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/issuer: argocd-adcs-issuer # issuser name
    cert-manager.io/issuer-kind: AdcsIssuer # ClusterAdcsIssuer or AdcsIssuer
    cert-manager.io/issuer-group: adcs.certmanager.csf.nokia.com # api group, here adcs.certmanager.csf.nokia.com
    cert-manager.io/renew-before: 48h # renew 48 hour before

  name: argo-cd-argocd-server
  namespace: argocd
 
spec:
  ingressClassName: nginx
  rules:
  - host: argocd.sample.host
    http:
      paths:
      - backend:
          service:
            name: argocd-server
            port:
              number: 443
        path: /(.*)
        pathType: Prefix
  tls:
  - hosts:
    - argocd.sample.host
    secretName: argocd-tls-certificate # secret for storing certificate
```




Check objects

```bash
kubectl -n argocd get certificate,certificaterequests
```


#### Working with operator


```
kustomize build config/crd > template.yaml
echo "---" >> template.yaml
kustomize build config/default >> template.yaml

make dry-run 

cat all-manifests.yaml | kubectl split-yaml -t "{{.kind}}/{{.name}}.yaml" -p manifests

kubectl apply -R -f manifests -n cert-manager

kubectl -n cert-manager logs deploy/adcs-issuer-controller-manager -c manager 

make build IMG="docker.io/djkormo/adcs-issuer:dev"

make docker-build docker-push IMG="docker.io/djkormo/adcs-issuer:dev"

docker build . -t docker.io/djkormo/adcs-issuer:dev

docker login docker.io/djkormo
docker push docker.io/djkormo/adcs-issuer:dev


git tag 2.1.1
git push origin --tags

```


## Using adcs simulator

Deploy this simulator. 
Note! Using helm chart 2.0.9 and above you can install it using helm commands.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: adcs-sim-deployment
  namespace: adcs-issuer
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: adcs-sim
  template:
    metadata:
      labels:
        control-plane: adcs-sim
    spec:
      containers:
      - args:
        - --dns=adcs-sim-service.cert-manager.svc,adcs2.example.com 
        - --ips=10.10.10.1,10.10.10.2
        - --port=8443
        command:
        - /usr/local/adcs-sim/manager
        image: djkormo/adcs-sim:0.0.5
        imagePullPolicy: Always
        env:
        - name: ENABLE_DEBUG
          value: "false"  
        name: manager
        volumeMounts:

        # emptydirs for storing csr and cert files
        - name: csr
          mountPath: "/usr/local/adcs-sim/ca"

        # ca cert 
        - name: config-pem
          mountPath: "/usr/local/adcs-sim/ca/root.pem"
          subPath: root.pem
          readOnly: true

        # ca key
        - name: config-key
          mountPath: "/usr/local/adcs-sim/ca/root.key"
          subPath: root.key
          readOnly: true

        ports:
        - containerPort: 8443 # the same as --port=8443 in arguments
          name: adcs-sim
          protocol: TCP
        resources:
          limits:
            cpu: 100m
            memory: 500Mi
          requests:
            cpu: 100m
            memory: 100Mi

      terminationGracePeriodSeconds: 10

      volumes:

        - name: csr
          emptyDir:
            sizeLimit: 50Mi 

        - name: config-pem
          configMap:
            name: adcs-sim-configmap # configmap for storing ca cert

        - name: config-key
          configMap:
            name: adcs-sim-configmap # configmap for storing ca key
---
apiVersion: v1
kind: Service
metadata:
  name: adcs-sim-service
  namespace: adcs-issuer
spec:
  ports:
  - port: 8443
    targetPort: 8443
  selector:
    control-plane: adcs-sim
```




Generate the private key of the root CA:

```console
openssl genrsa -out root.pem 4096
```

Generate the self-signed root CA certificate:

```console
openssl req -x509 -sha256 -new -nodes -key root.pem -days 3650 -out root.key -addext "subjectAltName=DNS:example.com,DNS:*.example.com,IP:10.0.0.1" \

  -subj '/C=PL/ST=Warsaw/L=Mordor/O=ADCSSIM/OU=IT/CN=example.com'
```

Review the certificate:

```console
openssl x509 -in root.key -text
```


Use your own ca cert and key. Recomended only for development purpose, you should convert configmap to secret if needed.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: adcs-sim-configmap
  namespace: adcs-issuer
data:

  root.pem: |
    -----BEGIN CERTIFICATE-----
    REDACTED
    -----END CERTIFICATE-----

 
  root.key: |
    -----BEGIN RSA PRIVATE KEY-----
    REDACTED
    -----END RSA PRIVATE KEY-----

```

Deploy credentials and configuration for adcs simulator

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: adcs-issuer-credentials
  namespace: adcs-issuer # namespace of adcs operator
type: Opaque
data:
  password: REDACTED # password
  username: REDACTED # username
---
apiVersion: adcs.certmanager.csf.nokia.com/v1
kind: ClusterAdcsIssuer
metadata:
  name: adcs-cluster-issuer-adcs-sim
spec:
  caBundle: REDACTED # ca bundle from adcs simulator
  credentialsRef:
    name: adcs-issuer-credentials # secret with username and password
  statusCheckInterval: 1m
  retryInterval: 1m
  url: https://adcs-sim-service.cert-manager.svc:8443 # external host via kubernetes service
  templateName: BasicSSLWebServer # external template 
```


Deploy sample certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:

  name: adcs-sim-cert
  namespace: adcs-issuer
spec:
  commonName: example.com
  dnsNames:
  - adcs1.example.com
  - adcs2.example.com

  issuerRef:
    group: adcs.certmanager.csf.nokia.com # api group, here adcs.certmanager.csf.nokia.com
    kind: ClusterAdcsIssuer # ClusterAdcsIssuer or AdcsIssuer
    name: adcs-cluster-issuer-adcs-sim # issuser name

  secretName: adcs-sim-secret # where to store certificate
```

Have fun

```
kubectl -n cert-manager get certificate,certificaterequest,adcsrequest
```
<pre>
NAME                                                   READY   SECRET                              AGE
certificate.cert-manager.io/adcs-sim-cert              True    adcs-sim-secret                     28m

NAME                                                       APPROVED   DENIED   READY   ISSUER                         REQUESTOR                                         AGE
certificaterequest.cert-manager.io/adcs-sim-cert-2v677     True                True    adcs-cluster-issuer-adcs-sim   system:serviceaccount:cert-manager:cert-manager   27m

NAME                                                                 STATE
adcsrequest.adcs.certmanager.csf.nokia.com/adcs-sim-cert-2v677       ready
</pre>



## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/djkormo/adcs-issuer/blob/master/LICENSE).


