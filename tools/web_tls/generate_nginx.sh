#!/usr/bin/env bash

set -eu

repository_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
output_directory="${repository_root}/zbx_env/etc/ssl/nginx"

mkdir -p "${output_directory}"
umask 077

openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
    -subj '/CN=localhost' \
    -addext 'basicConstraints=critical,CA:FALSE' \
    -addext 'extendedKeyUsage=serverAuth' \
    -addext 'subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1' \
    -keyout "${output_directory}/ssl.key" \
    -out "${output_directory}/ssl.crt"

openssl genpkey -genparam -algorithm DH -pkeyopt group:ffdhe2048 \
    -out "${output_directory}/dhparam.pem"

chmod 0644 "${output_directory}/ssl.crt" "${output_directory}/dhparam.pem"

printf 'Nginx TLS certificate has been generated.\n'
