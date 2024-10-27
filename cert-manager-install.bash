helm repo add jetstack https://charts.jetstack.io --force-update

helm search repo cert-manager

helm search repo cert-manager --versions

helm upgrade --install \
  cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --create-namespace \
    --version v1.14.7 \
    --set enableCertificateOwnerRef=true \
    --set installCRDs=true