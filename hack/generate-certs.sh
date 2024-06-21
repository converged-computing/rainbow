#!/bin/bash

here=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
root=${1}

if [ "${root}" == "" ]; then
    root=$here
fi

cd ${root}
echo "Building in ${root}"
ls
sleep 3

# delete pem file
rm *.pem

# Create CA private key and self-signed certificate
# adding -nodes to not encrypt the private key
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout ca-key.pem -out ca-cert.pem -subj "/C=US/ST=California/CN=localhost"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# Create Web Server private key and CSR
# adding -nodes to not encrypt the private key
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=US/ST=California/CN=localhost"

# Sign the Web Server Certificate Request (CSR)
openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile $here/server-ext.conf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

# Verify certificate
echo "Verifying certificate"
openssl verify -CAfile ca-cert.pem server-cert.pem

# Generate client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=US/ST=California/CN=localhost"

#  Sign the Client Certificate Request (CSR)
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile $here/client-ext.conf

echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text

ls .
