# ADCS Issuer

![Badge1](https://github.com/djkormo/adcs-issuer/actions/workflows/test.yaml/badge.svg) ![Badge2](https://github.com/djkormo/adcs-issuer/actions/workflows/codeql.yaml/badge.svg) ![Badge3](https://github.com/djkormo/adcs-issuer/actions/workflows/release.yaml/badge.svg) ![Badge4](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-test.yaml/badge.svg) ![Badge5](https://github.com/djkormo/adcs-issuer/actions/workflows/helm-release.yaml/badge.svg) [![trivy](https://github.com/djkormo/adcs-issuer/actions/workflows/trivy.yml/badge.svg)](https://github.com/djkormo/adcs-issuer/actions/workflows/trivy.yml) 

ADCS Issuer is a [Kubernetes](https://kubernetes.io/) [`cert-manager`](https://cert-manager.io)
[`CertificateRequest`](https://cert-manager.io/docs/concepts/certificaterequest/) controller
that uses [Microsoft Active Directory Certificate Services](https://learn.microsoft.com/en-us/windows-server/identity/ad-cs/active-directory-certificate-services-overview)
to sign certificate requests.

It supports NTLM authentication.

This project is a community maintained fork of the [original implementation by Nokia](https://github.com/nokia/adcs-issuer/).

## Getting started

TODO: a short summary of installing and configuring the issuer

## Documentation

Detailed documentation can be found in the [docs folder](./docs/README.md) or on [GitHub Pages](https://djkormo.github.io/adcs-issuer).

## Verifying container images

Released container images (e.g. `ghcr.io/scoobed/adcs-issuer`) are signed with
[`cosign`](https://github.com/sigstore/cosign), and an SBOM (Software Bill of Materials, generated with
[`syft`](https://github.com/anchore/syft) in CycloneDX format) is attached to each image as a signed attestation.
The public key used to verify these is checked into this repository as [`cosign.pub`](./cosign.pub).

To verify the signature on an image, first resolve its digest, then verify by digest
(always verify by digest rather than by tag, since tags are mutable):

```bash
IMG=$(docker inspect --format='{{index .RepoDigests 0}}' ghcr.io/scoobed/adcs-issuer:2.3.0-beta)
# or, without pulling: IMG=$(docker buildx imagetools inspect ghcr.io/scoobed/adcs-issuer:2.3.0-beta --format '{{.Manifest.Digest}}')

cosign verify --key cosign.pub "ghcr.io/scoobed/adcs-issuer@<digest>"
```

To verify the attached SBOM attestation and extract it:

```bash
cosign verify-attestation --key cosign.pub --type cyclonedx "ghcr.io/scoobed/adcs-issuer@<digest>" \
  | jq -r '.payload' | base64 -d | jq '.predicate' > sbom.cdx.json
```

A successful `cosign verify`/`cosign verify-attestation` reports that the claims were validated, their existence
in the transparency log was verified, and the signatures matched the public key above.

## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/adcs-issuer/blob/master/LICENSE).

