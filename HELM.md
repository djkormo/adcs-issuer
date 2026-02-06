helm  template ./charts/adcs-issuer/ --values  ./charts/adcs-issuer/values.yaml 

helm  template ./charts/adcs-issuer/ --values  ./charts/adcs-issuer/values.yaml  --show-only templates/deployment.yaml
helm  template ./charts/adcs-issuer/ --values  ./charts/adcs-issuer/values.yaml  --show-only templates/openshift-rbac.yaml

helm  template ./charts/adcs-issuer/ --values  ./docs/install-examples/values-openshift.yaml  --show-only templates/openshift-rbac.yaml


