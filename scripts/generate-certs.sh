#!/usr/bin/env bash

set -e
set -o pipefail

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
OUTPUT_DIR="${SCRIPT_DIR}/../issuers/testdata"

set -x

mkdir -pv "${OUTPUT_DIR}/ca"

# Create the CA key
openssl genrsa -out "${OUTPUT_DIR}/ca/ca.key" 2048
# Create a configuration file for the Root CA

# Create CA config
cat > "${OUTPUT_DIR}/ca/ca.cnf" << EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn
x509_extensions = v3_ca

[dn]
C = US
ST = YourState
L = YourCity
O = YourOrganization
OU = YourOrganizationalUnit
CN = adcs-issuer Test CA

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE
keyUsage = critical, digitalSignature, keyCertSign
EOF

# Generate the CA cert
openssl req -x509 -new -nodes \
    -key "${OUTPUT_DIR}/ca/ca.key" \
    -sha256 \
    -days 3650 \
    -out "${OUTPUT_DIR}/ca/ca.pem" \
    -config "${OUTPUT_DIR}/ca/ca.cnf"

# Create a PEM-encoded PKCS#7 bundle containing the CA certificate.
openssl crl2pkcs7 -nocrl \
    -certfile "${OUTPUT_DIR}/ca/ca.pem" \
    -outform PEM \
    -out "${OUTPUT_DIR}/pkcs7.pem"
cp -v "${OUTPUT_DIR}/ca/ca.pem" "${OUTPUT_DIR}/x509.pem"
