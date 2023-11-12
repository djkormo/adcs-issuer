# ADCS Issuer

ADCS Issuer is a [cert-manager's](https://github.com/jetstack/cert-manager) CertificateRequest controller that uses MS Active Directory Certificate Service to sign certificates 
(see [this design document](https://github.com/jetstack/cert-manager/blob/master/design/20190708.certificate-request-crd.md) for details on CertificateRequest CRD). 

ADCS provides HTTP GUI that can be normally used to request new certificates or see status of existing requests. 
This implementation is simply a HTTP client that interacts with the ADCS server sending appropriately prepared HTTP requests and interpretting the server's HTTP responses
(the approach inspired by [this Python ADCS client](https://github.com/magnuswatn/certsrv)).

It supports NTLM authentication.



### TODO 


1. Correct RBAC for cert-manager



Build statuses:


[![operator pipeline](https://github.com/djkormo/adcs-issuer/actions/workflows/pipeline.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/pipeline.yaml)


[![Code scanning - action](https://github.com/djkormo/adcs-issuer/actions/workflows/codeql.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/codeql.yaml)


[![Publish Docker image on Release](https://github.com/djkormo/adcs-issuer/actions/workflows/main.yml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/main.yml)


[![Release helm charts](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-chart-releaser.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-chart-releaser.yaml)

## Description

### Requirements
ADCS Issuer has been tested with cert-manager v1.9.x and currently supports CertificateRequest CRD API version v1 only.


### Locally operations

#### Installing cert manager 

```

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.0/cert-manager.yaml

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



git tag 2.0.4
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

```
# add helm repo
helm repo add djkormo-adcs-issuer https://djkormo.github.io/adcs-issuer/

# update 
helm repo update djkormo-adcs-issuer

# check all versions 
helm search repo adcs-issuer  --versions

# download values file for some version
helm show values djkormo-adcs-issuer/adcs-issuer --version 2.0.4 > values.yaml

# test installation
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.0.4 \
  --namespace cert-manager --values values.yaml  --dry-run

#  install
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.0.4 \
  --namespace cert-manager --values values.yaml  --dry-run

# upgrade
helm upgrade project-operator djkormo-adcs-issuer/adcs-issuer  --version 2.0.5 \
  --namespace cert-manager --values values.yaml

# uninstall 
helm uninstall adcs-issuer  --namespace  cert-manager

```


## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/adcs-issuer/blob/master/LICENSE).
