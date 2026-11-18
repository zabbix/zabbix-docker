#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/../.." && pwd)"
project=zabbix-history-providers
compose_files=(-f "${repo_dir}/compose.yaml" -f "${script_dir}/compose.yaml")

case "${1:-}" in
    "") ;;
    --tls)
        project=zabbix-history-providers-tls
        data_dir="${DATA_DIRECTORY:-$(sed -n 's/^DATA_DIRECTORY=//p' "${repo_dir}/.env")}"
        : "${data_dir:?DATA_DIRECTORY is not set}"
        cd "${repo_dir}"

        tls_dir="${data_dir}/etc/history_providers"
        provider_dir="${data_dir}/run/zabbix/secrets/providers/history-0"
        temp_dir="$(mktemp -d)"
        trap 'rm -rf -- "${temp_dir}"' EXIT
        umask 077

        openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
            -subj '/CN=Zabbix history providers example CA' \
            -addext 'basicConstraints=critical,CA:TRUE' \
            -addext 'keyUsage=critical,keyCertSign,cRLSign' \
            -keyout "${temp_dir}/ca.key" \
            -out "${temp_dir}/ca.crt"

        openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
            -subj '/CN=clickhouse' \
            -addext 'basicConstraints=critical,CA:FALSE' \
            -addext 'extendedKeyUsage=serverAuth' \
            -addext 'subjectAltName=DNS:clickhouse' \
            -keyout "${temp_dir}/clickhouse.key" \
            -out "${temp_dir}/clickhouse.crt" \
            -CA "${temp_dir}/ca.crt" \
            -CAkey "${temp_dir}/ca.key"

        openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
            -subj '/CN=zabbix-server' \
            -addext 'basicConstraints=critical,CA:FALSE' \
            -addext 'extendedKeyUsage=clientAuth' \
            -keyout "${temp_dir}/client.key" \
            -out "${temp_dir}/client.crt" \
            -CA "${temp_dir}/ca.crt" \
            -CAkey "${temp_dir}/ca.key"

        mkdir -p "${tls_dir}" "${provider_dir}"
        install -m 0644 "${temp_dir}/clickhouse.crt" "${tls_dir}/clickhouse.crt"
        install -m 0600 "${temp_dir}/clickhouse.key" "${tls_dir}/clickhouse.key"
        install -m 0644 "${temp_dir}/ca.crt" "${provider_dir}/ca.crt"
        install -m 0644 "${temp_dir}/client.crt" "${provider_dir}/client.crt"
        install -m 0600 "${temp_dir}/client.key" "${provider_dir}/client.key"

        compose_files+=(-f "${script_dir}/compose.tls.yaml")
        ;;
    *)
        printf 'Usage: %s [--tls]\n' "$0" >&2
        exit 1
        ;;
esac

docker compose \
    -p "$project" --project-directory "$repo_dir" \
    "${compose_files[@]}" \
    --profile history-providers --profile clickhouse --profile elasticsearch \
    up -d \
    --wait --wait-timeout 180 \
    clickhouse elasticsearch server-history-init zabbix-server zabbix-web-nginx-mysql

printf 'Stop: docker compose -p %s down --remove-orphans\n' "$project"
