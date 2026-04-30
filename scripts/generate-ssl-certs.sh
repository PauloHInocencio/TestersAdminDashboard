#!/bin/zsh
#Script to generate self-signed SSL certificates for PostgreSQL

# Tells the shell to exit immediately if any command fails.
set -e

# Creates a certs directory to store all certificate files.
# The -p flag means "create parent directories if needed, and don't error if it already exists."
CERT_DIR="./certs"
echo "Creating certificate directory..."
mkdir -p "$CERT_DIR"

# - genrsa: Generates an RSA privite key
# - 2048: Key size in bits (2048-bit is standard, though 4096-bit is more secure)
# - This private key is the secret that only the server should have - it's used to decrypt messages and prove identity
echo "Creating private key..."
openssl genrsa -out "$CERT_DIR/server.key" 2048

# Sets permissions so only the owner can read/write the file (6 = red+write for owner, 0 = no permissions for group/others)
# Critical security step - if others can read your private key, they can impersonate your server!
echo "Setting private key permissions..."
chmod 600 "$CERT_DIR/server.key"

# - req -new: Creates a new certificate sining request
# - -key: Uses the private key generated earlier
# - -subj: The subject information (who is requesting the certificate):
#     - C=BR: Country (Brazil)
#     - ST=SP: State/Province (São Paulo)
#     - L=Bauru: Location/City (Bauru)
#     - O=TestersAdmin: Organization
#     - CN=postgres17-testers: Common Name (the hostname/domain)
# The CSR is like an application form saying "I want a certificate for this identity."
echo "Generating certificate signing request..."
openssl req -new -key "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.csr" \
  -subj "/C=BR/ST=SP/L=Bauru/O=TestersAdmin/CN=postgres17-testers"

# -signkey: This is the self-signing part! Instead of sending the CSR to a Certificate Authority (CA),
#  we sign it ourselves with our own private key
echo  "Generating self-signed certificate (valid for 365 days)..."
openssl x509 -req -days 365 \
  -in "$CERT_DIR/server.csr" \
  -signkey "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.crt"

# Copies the server certificate to also act as the root CA certificate.
# In a real CA system, these would be different.
echo "Generating root certificate (same as server for self-signed)..."
cp "$CERT_DIR/server.crt" "$CERT_DIR/root.crt"

# Makes certificates readable by everyone (6 for owner = read+write, 4 for group/others = read only).
# Unlike private keys, certificates are meant to be public.
echo "Setting certificate permissions..."
chmod 644 "$CERT_DIR/server.crt"
chmod 644 "$CERT_DIR/root.crt"

echo "SSL certificates generated successfully in $CERT_DIR/"
echo "" 
echo "Files created:"
echo " - server.key (private key)"
echo " - server.crt (server certificate)"
echo " - root.crt (root CA certificate)"
echo ""
echo "Note: These are self-signed certificates for development."
echo "For production, use certificates from a trusted CA."