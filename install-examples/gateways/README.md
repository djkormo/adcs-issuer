


kubectl apply -f "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml"


helm repo add jetstack https://charts.jetstack.io --force-update



helm search repo  cert-manager/cert-manager 

helm search repo  cert-manager/cert-manager --versions



helm upgrade  --install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.15.4  \
  --set config.enableGatewayAPI=true \
  --set config.apiVersion="controller.config.cert-manager.io/v1alpha1" \
  --set config.kind="ControllerConfiguration" \
  --set enableCertificateOwnerRef=true


kubectl delete -f "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml"


Based on  https://cert-manager.io/docs/usage/gateway/





