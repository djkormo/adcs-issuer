# ADCS Issuer

ADCS Issuer is a [cert-manager's](https://github.com/jetstack/cert-manager) CertificateRequest controller that uses MS Active Directory Certificate Service to sign certificates 
(see [this design document](https://github.com/jetstack/cert-manager/blob/master/design/20190708.certificate-request-crd.md) for details on CertificateRequest CRD). 

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
ADCS Issuer has been tested with cert-manager v1.9.x and v.12.x and currently supports CertificateRequest CRD API version v1 only.

### Locally operations

#### Installing cert manager 

```
# version 1.9.0
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.0/cert-manager.yaml
# version 1.12.6
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.6/cert-manager.yaml

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


git tag 2.0.8
git push origin --tags


```

### Helm chart

Testing locally

```


helm lint chart/adcs-issuer

helm template charts/adcs-issuer -n cert-manager --values charts/adcs-issuer/values.yaml

helm template charts/adcs-issuer -n adcs-issuer --values charts/adcs-issuer/values.yaml > adcs-issuer-all.yaml

kubectl -n cert-manager apply -f adcs-issuer-all.yaml 

kubectl -n cert-manager rollout restart deploy adcs-issuer-controller-manager

kubectl -n cert-manager logs deploy/adcs-issuer-controller-manager -f

```

Using helm chart repo via github repo

```bash
# add helm repo
helm repo add djkormo-adcs-issuer https://djkormo.github.io/adcs-issuer/

# update 
helm repo update djkormo-adcs-issuer

# check all versions 
helm search repo adcs-issuer  --versions

# download values file for some version
helm show values djkormo-adcs-issuer/adcs-issuer --version 2.0.8 > values.yaml

# test installation
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.0.8 \
  --namespace cert-manager --values values.yaml  --dry-run

#  install
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.0.8 \
  --namespace cert-manager --values values.yaml  --dry-run

# upgrade
helm upgrade project-operator djkormo-adcs-issuer/adcs-issuer  --version 2.0.8 \
  --namespace cert-manager --values values.yaml

# uninstall 
helm uninstall adcs-issuer  --namespace  cert-manager

```

Example of values.yaml file for version 2.0.8 and above 

```yaml
adcs-issuer:
  crd:
    install: true

  controllerManager:
    manager:
      image:
        repository: djkormo/adcs-issuer
        tag: 2.0.8
      resources:
        limits:
          cpu: 100m
          memory: 500Mi
        requests:
          cpu: 100m
          memory: 100Mi

    rbac:
      enabled: true
      serviceAccountName: adcs-issuer # service account for rbac
      certManagerNamespace: cert-manager # cert manager serviceaccount
      certManagerServiceAccountName: cert-manager  # cert manager namespace


    replicas: 1

    environment:
      KUBERNETES_CLUSTER_DOMAIN: cluster.local
      ENABLE_WEBHOOKS: "false"
      ENABLE_DEBUG: "false"
    arguments:
      enable-leader-election: "true"
      cluster-resource-namespace: cert-manager # namespoace for cluster scoped resources, common secret
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


```


Create credentials for adcd

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: adcs-issuer-credentials
  namespace: cert-manager # namespace of cert managera and adcs operator
type: Opaque
data:
  password: REDACTED # Password
  username: REDACTED # username

```


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

Add annotatioons for ingress object, here for argocd server

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/issuer: argocd-adcs-issuer # issuser name
    cert-manager.io/issuer-kind: AdcsIssuer # ClusterAdcsIssuer or AdcsIssuer
    cert-manager.io/issuer-group: adcs.certmanager.csf.nokia.com # api group, here adcs.certmanager.csf.nokia.com
    cert-manager.io/duration: 17520h # 2 years
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


## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/adcs-issuer/blob/master/LICENSE).
