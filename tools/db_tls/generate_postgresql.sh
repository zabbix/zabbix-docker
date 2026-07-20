#!/usr/bin/env bash

set -eu

ca_cn='Zabbix PostgreSQL TLS test CA'
server_cn='postgres-server'
client_cn='zabbix'

repo_root="$(realpath "$(dirname "${BASH_SOURCE[0]}")/../..")"
out_dir="${repo_root}/env_vars"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/zabbix-postgresql-tls.XXXXXX")"
trap 'rm -rf -- "${tmp_dir}"' EXIT

umask 077

# Generate a local certificate authority for the PostgreSQL TLS example
openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 3650 \
    -subj "/CN=${ca_cn}" \
    -addext 'basicConstraints=critical,CA:TRUE' \
    -addext 'keyUsage=critical,keyCertSign,cRLSign' \
    -keyout "${tmp_dir}/ca.key" \
    -out "${tmp_dir}/ca.crt" \
    >/dev/null 2>&1

# Generate the PostgreSQL server private key and certificate signing request
openssl req -newkey rsa:2048 -nodes -sha256 \
    -subj "/CN=${server_cn}" \
    -addext 'basicConstraints=critical,CA:FALSE' \
    -addext 'extendedKeyUsage=serverAuth' \
    -addext "subjectAltName=DNS:${server_cn}" \
    -keyout "${tmp_dir}/server.key" \
    -out "${tmp_dir}/server.csr" \
    >/dev/null 2>&1

# Sign the PostgreSQL server certificate with the local certificate authority
openssl x509 -req -sha256 -days 365 -copy_extensions copy \
    -in "${tmp_dir}/server.csr" \
    -CA "${tmp_dir}/ca.crt" \
    -CAkey "${tmp_dir}/ca.key" \
    -CAcreateserial \
    -out "${tmp_dir}/server.crt" \
    >/dev/null 2>&1

# Generate the Zabbix database client private key and signing request
openssl req -newkey rsa:2048 -nodes -sha256 \
    -subj "/CN=${client_cn}" \
    -addext 'basicConstraints=critical,CA:FALSE' \
    -addext 'extendedKeyUsage=clientAuth' \
    -keyout "${tmp_dir}/client.key" \
    -out "${tmp_dir}/client.csr" \
    >/dev/null 2>&1

# Sign the Zabbix database client certificate with the local certificate authority
openssl x509 -req -sha256 -days 365 -copy_extensions copy \
    -in "${tmp_dir}/client.csr" \
    -CA "${tmp_dir}/ca.crt" \
    -CAkey "${tmp_dir}/ca.key" \
    -CAserial "${tmp_dir}/ca.srl" \
    -out "${tmp_dir}/client.crt" \
    >/dev/null 2>&1

mkdir -p "${out_dir}"
install -m 0600 "${tmp_dir}/ca.key" "${out_dir}/.DB_CA_KEY_FILE"
install -m 0644 "${tmp_dir}/ca.crt" "${out_dir}/.ZBX_DB_CA_FILE"
install -m 0644 "${tmp_dir}/server.crt" "${out_dir}/.DB_CERT_FILE"
install -m 0600 "${tmp_dir}/server.key" "${out_dir}/.DB_KEY_FILE"
install -m 0644 "${tmp_dir}/client.crt" "${out_dir}/.ZBX_DB_CERT_FILE"
install -m 0600 "${tmp_dir}/client.key" "${out_dir}/.ZBX_DB_KEY_FILE"

printf 'PostgreSQL TLS certificates have been generated...\n'
