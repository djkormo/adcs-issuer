# ADCS-Issuer setup for developers

## pre-requisite

- make sure the kubernetes current-context is correctly set to the cluster you want to deploy the ADCS-issuer to.

## build and push the adcs-controller

This step is only needed when debugging / changing code.

- #TODO setup skaffold for better workflow

> IMG=pietere/controller:latest make docker-build docker-push

## webhook certificate

Create a certificate using selfsigned issuer from cert-manager in order to get the  the webhook & controller working.

- #TODO this needs to be added to adcs-controller resources
- #TODO the commonName is hard-coded.

cat <<EOF | kubectl -n adcs-issuer-system apply -f -

``` YAML
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: serving-cert
spec:
  commonName: adcs-issuer-webhook-service.adcs-issuer-system.svc
  dnsNames:
  - adcs-issuer-webhook-service.adcs-issuer-system.svc.cluster.local
  issuerRef:
    kind: ClusterIssuer
    name: selfsigned
  secretName: webhook-server-cert
```

## deploy adcs-issuer controller to k8s-cluster

First deploy the crds:
> make install

Replace the image and deploy the adcs-issuer resources:
> IMG=pietere/controller:latest make deploy

## NTLM

#TODO check where to specify the namespace to deploy the credentials/secret to.

cat <<EOF | kubectl -n kube-system apply -f -

``` YAML
apiVersion: v1
kind: Secret
metadata:
  name: test-adcs-issuer-credentials
type: Opaque
data:
  password: cGFzc3dvcmQ=
  username: dXNlcm5hbWU=
```

## start ADCS-simulator

Run the adcs-simulator in another shell:
> make sim

Get the server.pem / root.pem certificate for the adcs-issuer to authenticate against the ADCS-simulator-server.
> openssl s_client -connect localhost:8443 -showcerts

Encode the certificate to base64 and in one line, for caBundle.
> cat <<EOF | base64 -w 0

## (Cluster) ADCS-Issuer

Deploy the AdcsIssuer to the k8s-cluster.
The ClusterAdcsIssuer is deployed the same way.

cat <<EOF | kubectl -n adcs-issuer-system apply -f -

``` YAML
kind: AdcsIssuer
apiVersion: adcs.certmanager.csf.nokia.com/v1
metadata:
  name: test-adcs-issuer
spec:
  caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM5ekNDQWQrZ0F3SUJBZ0lKQU92T05idHY0OU5OTUEwR0NTcUdTSWIzRFFFQkN3VUFNQkl4RURBT0JnTlYKQkFNTUIwRkVRMU5UYVcwd0hoY05NVGt3TnpJd01Ua3lOekV4V2hjTk1qa3dOekUzTVRreU56RXhXakFTTVJBdwpEZ1lEVlFRRERBZEJSRU5UVTJsdE1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBCm1TMjMreXZpRnBURU1HTnQzbHYrSmt3cEFLeEt6QlhSU2MvYkpUWWphOGxTbWc4WTZsQWZrYURBYWdjeUltUkYKY2Y1dzBCNVhVT2R4TFpIM0ROR0VZUEp1bEIzZ2ZlMVYvdm90UzFCZzFMSEtHbWVvS1JYQVZLVERUK25kM3d0KwphQ1l2d2tFVjBjSTRyTnNMd1VWVW5qZktjbEVPTGJQMVRGYUJSOG1VWWdkN0ZqUFd1T1hwbk5Hb3RyNWdkbmhNCmJXVjJ4SEFHWVR5Nno3U05EVHJNQjVWWTVYQVA1ZlRvN09pT2UyUHlnMnFiMmdUMnhUWjZ4OUo0aCtrUTMreEUKRWRaTVVHeUF0WWlFTFhZN3dBZGxwS0hjMDJKejUrVGgzWUdUbWgxbk1MbW84eUpyZUJWMytIVlozVU55TGx6NAppUSszUko1Q2tMcTIxQS9yUW5qZlV3SURBUUFCbzFBd1RqQWRCZ05WSFE0RUZnUVVDTUtMM3htVEUzOWwwbXl0CmxtQkZlU3FabUdBd0h3WURWUjBqQkJnd0ZvQVVDTUtMM3htVEUzOWwwbXl0bG1CRmVTcVptR0F3REFZRFZSMFQKQkFVd0F3RUIvekFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBZXdjM2JtZlFDaDNKanJkc2tsMnliWXE3ZTlKRgpFME9JNjgyWEllWWhVUHdUU1pHMy9Od01BbmtlK3QrRmR5TU9ENVNvL2NxQ2VtTGlCVWJGUnR6QmpyTlBJRUplCnUyNTZYekVxbXBDNmw5K0tGQzVqMzZKQ01leFZ6L2hxUnU4SlFJaitOenJBb0prTCtTNzBMdk1QUk90ZDVOS2UKNnQ3d3VmTE9RRkxHanBuU3lyWHEzbGRHZ0JRWGw3bG5JdVVMd0lJak9YcWR6OTFMa2VQbGVCRVV6QUdmZERmTwpJSU9GWWNwMDVoa2lmbm1SSzA1VTJucDAwWGZMakhRZEVRNDdVZ3NPZW9ncXl3UEg4WWRGaUQvRy9WWnVZMmZKCnVWRFNXQVljRHpwTmIzVWUvam5qQlBjTTlqZTB2YVdhZUE4UG10Y3dVeTZlOGdsVjdBZlZOWXVvQ3c9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  credentialsRef:
    name: test-adcs-issuer-credentials
  statusCheckInterval: 6h
  retryInterval: 1h
  url: https://localhost:8443
```

## Example certificate

Apply a certificate resource, that will make a certificaterequest with the issuerRef pointed to AdcsIssuer.

cat <<EOF | kubectl -n adcs-issuer-system apply -f -

``` YAML
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  annotations:
  name: adcs-certificate
spec:
  commonName: localhost
  dnsNames:
  - service1.example.com
  - service2.example.com
  issuerRef:
    group: adcs.certmanager.csf.nokia.com
    kind: AdcsIssuer
    name: test-adcs-issuer
  organization:
  - Your organization
  secretName: adcsIssuer-certificate
```


In one terminal

inside test/adcs-sim

```

go build -o adcs-sim main.go

go run main.go  --workdir=/d/development/kubernetes/go/adcs-issuer/test/adcs-sim --dns=adcs1.example.com,adcs1.example.com --ips=10.10.10.1,10.10.10.2


./adcs-sim  --workdir=/d/development/kubernetes/go/adcs-issuer/test/adcs-sim --dns=adcs1.example.com,adcs1.example.com --ips=10.10.10.1,10.10.10.2


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

Based on
https://stackoverflow.com/questions/10175812/how-to-generate-a-self-signed-ssl-certificate-using-openssl




openssl x509 -in test/adcs-sim/ca/root.pem -noout -text


kubectl -n cert-manager port-forward svc/adcs-sim-service 8443:8443&


https://localhost:8443/certcarc.asp

https://localhost:8443/certfnsh.asp

https://localhost:8443/certnew.cer

https://localhost:8443/certnew.p7b



openssl s_client -connect localhost:8443 -showcerts



username=$(kubectl get secret adcs-issuer-credentials  -n cert-manager -o jsonpath='{.data.username}' | base64 --decode)
password=$(kubectl get secret adcs-issuer-credentials  -n cert-manager -o jsonpath='{.data.password}' | base64 --decode)
url=$(kubectl get adcsissuer adcs-cluster-issuer-adcs-sim  -n cert-manager -o jsonpath='{.spec.url}')
ca=$(kubectl get clusteradcsissuer adcs-cluster-issuer-adcs-sim  -n cert-manager -o jsonpath='{.spec.caBundle}' | base64 --decode  )
echo "username: ${username}"
echo "password: ${password}"
echo "url: ${url}"
echo "ca: ${ca}"
echo ${ca} > ca.crt
curl  -k -u "${username}:${password}" --ntlm "${url}/certfnsh.asp" -vv
curl  -k --cacert ./ca.crt  -u "${username}:${password}" --ntlm "${url}/certfnsh.asp" -vv

curl  -k -u '${username}:${password}' --ntlm '${url}/certsrv/certfnsh.asp' -vv

curl -X POST -k -v -u "${username}:${password}" --ntlm "${url}/certcarc.asp" -vv

curl -X POST -k -u "${username}:${password}" --ntlm "${url}/certfnsh.asp" -vv


Usefull command for testing 

  kubectl -n cert-manager logs deploy/adcs-issuer-controller-manager

  kubectl -n cert-manager logs deploy/adcs-sim-deployment

  kubectl -n cert-manager logs deploy/cert-manager

  kubectl -n cert-manager get certificate,certificaterequest,adcsrequest 

  kubectl -n cert-manager delete certificate --all   
  kubectl -n cert-manager delete certificaterequest --all
  kubectl -n cert-manager delete adcsrequest --all

  kubectl -n cert-manager rollout restart deploy/adcs-issuer-controller-manager
  kubectl -n cert-manager rollout restart deploy/adcs-sim-deployment




polaris audit --helm-chart ./charts/adcs-issuer --helm-values ./charts/adcs-issuer/values.yaml --format=pretty 

```console
Polaris audited Path /tmp/3545659651 at 2024-10-26T16:25:36+02:00
    Nodes: 0 | Namespaces: 0 | Controllers: 1
    Final score: 92

RoleBinding release-name-adcs-issuer-leader-election-rolebinding in namespace default
    rolebindingClusterAdminClusterRole   🎉 Success
        Security - The RoleBinding does not reference the default cluster-admin ClusterRole or one with wildcard permissions
    rolebindingClusterAdminRole          🎉 Success
        Security - The RoleBinding does not reference a Role with wildcard permissions
    rolebindingClusterRolePodExecAttach  🎉 Success
        Security - The RoleBinding does not reference a ClusterRole allowing pods/exec or pods/attach
    rolebindingRolePodExecAttach         🎉 Success
        Security - The RoleBinding does not reference a Role allowing Pod exec or attach

Service release-name-adcs-issuer-controller-manager-metrics-service in namespace default

ServiceAccount adcs-issuer in namespace default

CustomResourceDefinition adcsissuers.adcs.certmanager.csf.nokia.com

CustomResourceDefinition adcsrequests.adcs.certmanager.csf.nokia.com

CustomResourceDefinition clusteradcsissuers.adcs.certmanager.csf.nokia.com

ClusterRole release-name-adcs-issuer-cert-manager-controller-approve-adcs-certmanager-csf-nokia-com
    clusterrolePodExecAttach             🎉 Success
        Security - The ClusterRole does not allow pods/exec or pods/attach

ClusterRole release-name-adcs-issuer-manager-role
    clusterrolePodExecAttach             🎉 Success
        Security - The ClusterRole does not allow pods/exec or pods/attach

ClusterRole release-name-adcs-issuer-proxy-role
    clusterrolePodExecAttach             🎉 Success
        Security - The ClusterRole does not allow pods/exec or pods/attach

ClusterRoleBinding release-name-adcs-issuer-cert-manager-controller-approve-adcs-certmanager-csf-nokia-com
    clusterrolebindingClusterAdmin       🎉 Success
        Security - The ClusterRoleBinding does not reference the default cluster-admin ClusterRole or one with wildcard permissions
    clusterrolebindingPodExecAttach      🎉 Success
        Security - The ClusterRoleBinding does not reference a ClusterRole allowing pods/exec or pods/attach

ClusterRoleBinding release-name-adcs-issuer-manager-rolebinding
    clusterrolebindingClusterAdmin       🎉 Success
        Security - The ClusterRoleBinding does not reference the default cluster-admin ClusterRole or one with wildcard permissions
    clusterrolebindingPodExecAttach      🎉 Success
        Security - The ClusterRoleBinding does not reference a ClusterRole allowing pods/exec or pods/attach

ClusterRoleBinding release-name-adcs-issuer-proxy-rolebinding
    clusterrolebindingClusterAdmin       🎉 Success
        Security - The ClusterRoleBinding does not reference the default cluster-admin ClusterRole or one with wildcard permissions
    clusterrolebindingPodExecAttach      🎉 Success
        Security - The ClusterRoleBinding does not reference a ClusterRole allowing pods/exec or pods/attach

Deployment release-name-adcs-issuer-controller-manager in namespace default
    deploymentMissingReplicas            😬 Warning
        Reliability - Only one replica is scheduled
    metadataAndInstanceMismatched        😬 Warning
        Reliability - Label app.kubernetes.io/instance must match metadata.name
    missingPodDisruptionBudget           😬 Warning
        Reliability - Should have a PodDisruptionBudget
    pdbMinAvailableGreaterThanHPAMinReplicas 🎉 Success
        Reliability - PDB and HPA are correctly configured
    hostIPCSet                           🎉 Success
        Security - Host IPC is not configured
    hostNetworkSet                       🎉 Success
        Security - Host network is not configured
    hostPIDSet                           🎉 Success
        Security - Host PID is not configured
    missingNetworkPolicy                 😬 Warning
        Security - A NetworkPolicy should match pod labels and contain applied egress and ingress rules
    priorityClassNotSet                  😬 Warning
        Reliability - Priority class should be set
    topologySpreadConstraint             🎉 Success
        Reliability - Pod has a valid topology spread constraint
    automountServiceAccountToken         😬 Warning
        Security - The ServiceAccount will be automounted
  Container manager
    hostPortSet                          🎉 Success
        Security - Host port is not configured
    memoryLimitsMissing                  🎉 Success
        Efficiency - Memory limits are set
    insecureCapabilities                 🎉 Success
        Security - Container does not have any insecure capabilities
    livenessProbeMissing                 🎉 Success
        Reliability - Liveness probe is configured
    memoryRequestsMissing                🎉 Success
        Efficiency - Memory requests are set
    notReadOnlyRootFilesystem            🎉 Success
        Security - Filesystem is read only
    readinessProbeMissing                🎉 Success
        Reliability - Readiness probe is configured
    sensitiveContainerEnvVar             🎉 Success
        Security - The container does not set potentially sensitive environment variables
    cpuLimitsMissing                     🎉 Success
        Efficiency - CPU limits are set
    dangerousCapabilities                🎉 Success
        Security - Container does not have any dangerous capabilities
    linuxHardening                       🎉 Success
        Security - One of AppArmor, Seccomp, SELinux, or dropping Linux Capabilities are used to restrict containers using unwanted privileges
    tagNotSpecified                      🎉 Success
        Reliability - Image tag is specified
    cpuRequestsMissing                   🎉 Success
        Efficiency - CPU requests are set
    privilegeEscalationAllowed           🎉 Success
        Security - Privilege escalation not allowed
    pullPolicyNotAlways                  🎉 Success
        Reliability - Image pull policy is "Always"
    runAsPrivileged                      🎉 Success
        Security - Not running as privileged
    runAsRootAllowed                     🎉 Success
        Security - Is not allowed to run as root

Role release-name-adcs-issuer-leader-election-role in namespace default
    rolePodExecAttach                    🎉 Success
        Security - The Role does not allow pods/exec or pods/attach



🚀 Upload your Polaris findings to Fairwinds Insights to see remediation advice, add teammates, integrate with Slack or Jira, and more:

❯ polaris audit --helm-chart ./charts/adcs-issuer --helm-values ./charts/adcs-issuer/values.yaml --format=pretty --upload-insights --cluster-name=my-cluster
```
