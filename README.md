# ADCS Issuer

![Badge1](https://github.com/djkormo/adcs-issuer/actions/workflows/codeql.yaml/badge.svg) ![Badge2](https://github.com/djkormo/adcs-issuer/actions/workflows/main.yml/badge.svg) ![Badge3](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-chart-releaser.yaml/badge.svg) ![Badge4](https://github.com/djkormo/adcs-issuer/actions/workflows/golangci-lint.yaml/badge.svg)

ADCS Issuer is a [Kubernetes](https://kubernetes.io/) [`cert-manager`](https://cert-manager.io)
[`CertificateRequest`](https://cert-manager.io/docs/concepts/certificaterequest/) controller
that uses [Microsoft Active Directory Certificate Services](https://learn.microsoft.com/en-us/windows-server/identity/ad-cs/active-directory-certificate-services-overview)
to sign certificate requests.

It supports NTLM authentication.

This project is a community maintained fork of the [original implementation by Nokia](https://github.com/nokia/adcs-issuer/).

## Getting started

TODO: a short summary of installing and configuring the issuer

## Documentation

Detailed documentation can be found in the [docs folder](./docs/README.md) or on [GitHub Pages](djkormo.github.io/adcs-issuer).

## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/adcs-issuer/blob/master/LICENSE).
