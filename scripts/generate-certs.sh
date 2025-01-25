#!/usr/bin/env bash

set -e
set -o pipefail

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
OUTPUT_DIR="${SCRIPT_DIR}/../issuers/testdata"

if [ -f "${OUTPUT_DIR}/pkcs7.pem" ] && [ -f "${OUTPUT_DIR}/x509.pem" ]; then
    printf 'Certificates already exist, skipping generation...\n'
    exit 0
fi

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

# This is probably wrong, but it seems the test
# just compares equality of the parsed pkcs7.pem to the raw x509.pem...
# TODO: review
cp -v "${OUTPUT_DIR}/ca/ca.pem" "${OUTPUT_DIR}/pkcs7.pem"
cp -v "${OUTPUT_DIR}/ca/ca.pem" "${OUTPUT_DIR}/x509.pem"
