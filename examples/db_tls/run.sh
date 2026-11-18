#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/../.." && pwd)"
project=zabbix-db-tls

case "${1:-}" in
    mysql)
        database=mysql
        compose_file=compose.yaml
        db_service=mysql-server
        web_service=zabbix-web-nginx-mysql
        ;;
    postgresql)
        database=postgresql
        compose_file=compose_pgsql.yaml
        db_service=postgres-server
        web_service=zabbix-web-nginx-pgsql
        ;;
    *)
        printf 'Usage: %s mysql|postgresql\n' "$0" >&2
        exit 1
        ;;
esac

env_dir="${ENV_VARS_DIRECTORY:-$(sed -n 's/^ENV_VARS_DIRECTORY=//p' "${repo_dir}/.env")}"
: "${env_dir:?ENV_VARS_DIRECTORY is not set}"
temp_dir="$(mktemp -d)"
trap 'rm -rf -- "${temp_dir}"' EXIT
umask 077

openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
    -subj '/CN=Zabbix database TLS example CA' \
    -addext 'basicConstraints=critical,CA:TRUE' \
    -addext 'keyUsage=critical,keyCertSign,cRLSign' \
    -keyout "${temp_dir}/ca.key" \
    -out "${temp_dir}/ca.crt"

openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
    -subj "/CN=${db_service}" \
    -addext 'basicConstraints=critical,CA:FALSE' \
    -addext 'extendedKeyUsage=serverAuth' \
    -addext "subjectAltName=DNS:${db_service}" \
    -keyout "${temp_dir}/server.key" \
    -out "${temp_dir}/server.crt" \
    -CA "${temp_dir}/ca.crt" \
    -CAkey "${temp_dir}/ca.key"

cd "${repo_dir}"
install -m 0644 "${temp_dir}/ca.crt" "${env_dir}/.ZBX_DBTLSCAFILE"
install -m 0644 "${temp_dir}/server.crt" "${env_dir}/.DB_CERT_FILE"
install -m 0600 "${temp_dir}/server.key" "${env_dir}/.DB_KEY_FILE"

docker compose \
    -p "$project" --project-directory "$repo_dir" \
    -f "${repo_dir}/${compose_file}" -f "${script_dir}/compose.${database}.yaml" \
    up -d --force-recreate --remove-orphans \
    --wait --wait-timeout 180 \
    "$db_service" zabbix-server "$web_service"

printf 'Stop: docker compose -p %s down --remove-orphans\n' "$project"
