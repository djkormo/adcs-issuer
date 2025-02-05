#!/usr/bin/env bash

set -euo pipefail  # Improved: added 'u' to catch errors from undeclared variables

# 📍 Setting directories
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
OUTPUT_DIR="${SCRIPT_DIR}/../issuers/testdata"
CA_DIR="${OUTPUT_DIR}/ca"

# 📍 Checking if certificates already exist
if [[ -f "${OUTPUT_DIR}/pkcs7.pem" && -f "${OUTPUT_DIR}/x509.pem" ]]; then
    echo "✅ Certificates already exist. Skipping generation..."
    exit 0
fi

# 📍 Setting default values for CA certificate
COUNTRY="US"
STATE="YourState"
CITY="YourCity"
ORG="YourOrganization"
ORG_UNIT="YourOrganizationalUnit"
COMMON_NAME="adcs-issuer Test CA"
KEY_SIZE=4096
DAYS_VALID=3650  # 10 years

echo "🔧 Creating directories..."
mkdir -pv "${CA_DIR}"

# 📍 Generating CA private key
echo "🔑 Generating CA private key (${KEY_SIZE} bits)..."
openssl genrsa -out "${CA_DIR}/ca.key" ${KEY_SIZE}

# 📍 Creating CA configuration
echo "📜 Creating CA configuration..."
cat > "${CA_DIR}/ca.cnf" << EOF
[req]
default_bits = ${KEY_SIZE}
prompt = no
default_md = sha256
distinguished_name = dn
x509_extensions = v3_ca

[dn]
C = ${COUNTRY}
ST = ${STATE}
L = ${CITY}
O = ${ORG}
OU = ${ORG_UNIT}
CN = ${COMMON_NAME}

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE
keyUsage = critical, digitalSignature, keyCertSign
EOF

# 📍 Generating CA certificate
echo "📜 Generating CA certificate (valid for ${DAYS_VALID} days)..."
openssl req -x509 -new -nodes \
    -key "${CA_DIR}/ca.key" \
    -sha256 \
    -days ${DAYS_VALID} \
    -out "${CA_DIR}/ca.pem" \
    -config "${CA_DIR}/ca.cnf"

# 📍 Copying certificates to test files
echo "📂 Copying certificates..."
cp -v "${CA_DIR}/ca.pem" "${OUTPUT_DIR}/pkcs7.pem"
cp -v "${CA_DIR}/ca.pem" "${OUTPUT_DIR}/x509.pem"

echo "✅ Certificates successfully generated!"