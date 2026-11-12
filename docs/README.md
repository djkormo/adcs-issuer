# `adcs-issuer` Documentation

A documentation rewrite is in progress. In the mean time, please see the [old documentation](./OLD.md).

## CA bundle references

Both `AdcsIssuer` and `ClusterAdcsIssuer` support `spec.caBundleRef`. Only `name` is required; `kind` defaults to `Secret` and `key` to `ca.crt`.

`AdcsIssuer` always reads the referenced resource from its own namespace and does not support `caBundleRef.namespace`. `ClusterAdcsIssuer` supports a namespace override; if omitted, it uses `--cluster-resource-namespace` (`kube-system` for the binary, or `adcs-issuer` with the default Helm values).

See [issuer configuration](./OLD.md#issuers) for details and examples.
