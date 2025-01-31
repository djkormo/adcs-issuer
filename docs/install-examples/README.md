## install cert-manager 
```
helm repo add jetstack https://charts.jetstack.io --force-update
```
```
helm repo update
```
```
helm search repo cert-manager
helm search repo cert-manager --versions | grep v1.
```
```
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.12.6  \
  --set installCRDs=true
```

helm search repo cert-manager-csi-driver
helm search repo cert-manager-csi-driver --versions

helm install \
  cert-manager-csi-driver jetstack/cert-manager-csi-driver \
  --namespace cert-manager \
  --version v0.7.1



## Install adcs-issuer 

### add helm repo

```
helm repo add djkormo-adcs-issuer https://djkormo.github.io/adcs-issuer/
```

### update

``` 
helm repo update djkormo-adcs-issuer
```

### check all versions 
```
helm search repo adcs-issuer  --versions
```


### install in cert-manager namespace

```console 
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.0.8 \
  --namespace cert-manager --values values-cert-manager-namespace.yaml  --create-namespace
```


```
kubectl -n cert-manager get deploy
kubectl -n cert-manager logs deploy/adcs-issuer-controller-manager
```


### install adcs issuer in adcs-issuer namespace

```console 
helm install adcs-issuer  djkormo-adcs-issuer/adcs-issuer --version 2.0.8 \
  --namespace adcs-issuer --values values-adcs-issuer-namespace.yaml --create-namespace
```

#### Checks 
```
kubectl -n adcs-issuer get deploy
kubectl -n adcs-issuer logs deploy/adcs-issuer-controller-manager
```

## install adcs-simulator

### install adcs-simulator in adcs-issuer namespace

```
kubectl apply -R -f adcs-simulator/adcs-issuer-namespace/ -n adcs-issuer

```



#### Checks 
```
kubectl -n adcs-issuer logs deploy/adcs-sim-deployment
kubectl -n adcs-issuer get clusteradcsissuer,adcsrequest
kubectl  -n adcs-issuer get secrets --field-selector type=kubernetes.io/tls 
kubectl -n adcs-issuer get certificate  -ojsonpath="{range .items[*]}{.metadata.name} not before: {.status.notBefore} not after: {.status.notAfter}{'\n'}{end}" 
```

<pre>

NAME                                        READY   SECRET            AGE
certificate.cert-manager.io/adcs-sim-cert   True    adcs-sim-secret   25s

NAME                                                                            AGE
clusteradcsissuer.adcs.certmanager.csf.nokia.com/adcs-cluster-issuer-adcs-sim   22m

NAME              TYPE                DATA   AGE
adcs-sim-secret   kubernetes.io/tls   2      20m

adcs-sim-cert not before: 2024-02-07T21:05:58Z not after: 2025-02-06T21:05:58Z

</pre>


#### Installed helm charts 

```console
helm list -A
```
<pre>
NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                           APP VERSION
adcs-issuer     adcs-issuer     1               2024-02-07 21:57:18.0829885 +0100 CET   deployed        adcs-issuer-2.0.8               2.0.8
cert-manager    cert-manager    1               2024-02-07 22:16:03.1434831 +0100 CET   deployed        cert-manager-v1.12.6            v1.12.6
</pre>
