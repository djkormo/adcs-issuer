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


[![adcs-simulator pipeline](https://github.com/djkormo/adcs-issuer/actions/workflows/adcs-simulator.yaml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/adcs-simulator.yaml)

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



## Using adcs simulator

Deploy this simulator

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: adcs-sim-deployment
  namespace: cert-manager
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
  namespace: cert-manager
spec:
  ports:
  - port: 8443
    targetPort: 8443
  selector:
    control-plane: adcs-sim
```


Generate the private key of the root CA:

```
openssl genrsa -out root.pem 4096
```

Generate the self-signed root CA certificate:

```
openssl req -x509 -sha256 -new -nodes -key root.pem -days 3650 -out root.key -addext "subjectAltName=DNS:example.com,DNS:*.example.com,IP:10.0.0.1" \

  -subj '/C=PL/ST=Warsaw/L=Mordor/O=ADCSSIM/OU=IT/CN=example.com'
```

Review the certificate:
```
openssl x509 -in root.key -text
```


Use your own ca cert and key. Recomended only for development purpose, you should convert configmap to secret if needed.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: adcs-sim-configmap
  namespace: cert-manager
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
  namespace: cert-manager # namespace of adcs operator
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


Deploy sample certuficate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:

  name: adcs-sim-cert
  namespace: cert-manager
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

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/adcs-issuer/blob/master/LICENSE).
