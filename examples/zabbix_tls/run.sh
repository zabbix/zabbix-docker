#!/usr/bin/env bash
set -euo pipefail

mode="${1:-}"
if [[ "${mode}" != cert && "${mode}" != psk ]]; then
    printf 'Usage: %s cert|psk\n' "$0" >&2
    exit 1
fi

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/../.." && pwd)"
project=zabbix-tls

data_dir="${DATA_DIRECTORY:-$(sed -n 's/^DATA_DIRECTORY=//p' "${repo_dir}/.env")}"
: "${data_dir:?DATA_DIRECTORY is not set}"
cd "${repo_dir}"
output_dir="${data_dir}/run/zabbix/secrets/tls"
temp_dir="$(mktemp -d)"
trap 'rm -rf -- "${temp_dir}"' EXIT
umask 077

case "${mode}" in
    cert)
        openssl req -x509 -newkey rsa:2048 -sha256 -nodes -days 365 \
            -subj '/CN=Zabbix TLS example CA' \
            -addext 'basicConstraints=critical,CA:TRUE' \
            -addext 'keyUsage=critical,keyCertSign,cRLSign' \
            -keyout "${temp_dir}/ca.key" \
            -out "${temp_dir}/ca.crt"

        for component in server proxy agent frontend; do
            openssl req -x509 -newkey rsa:2048 -sha256 -nodes -days 365 \
                -subj "/CN=zabbix-${component}" \
                -addext 'basicConstraints=critical,CA:FALSE' \
                -addext 'keyUsage=digitalSignature,keyEncipherment' \
                -addext 'extendedKeyUsage=serverAuth,clientAuth' \
                -keyout "${temp_dir}/${component}.key" \
                -out "${temp_dir}/${component}.crt" \
                -CA "${temp_dir}/ca.crt" \
                -CAkey "${temp_dir}/ca.key"
        done

        mkdir -p "${output_dir}"
        install -m 0644 "${temp_dir}/ca.crt" "${output_dir}/ca.crt"
        for component in server proxy agent frontend; do
            install -m 0644 "${temp_dir}/${component}.crt" "${output_dir}/${component}.crt"
            install -m 0644 "${temp_dir}/${component}.key" "${output_dir}/${component}.key"
        done
        ;;
    psk)
        openssl rand -out "${temp_dir}/proxy.psk" -hex 32
        openssl rand -out "${temp_dir}/agent.psk" -hex 32
        mkdir -p "${output_dir}"
        install -m 0644 "${temp_dir}/proxy.psk" "${output_dir}/proxy.psk"
        install -m 0644 "${temp_dir}/agent.psk" "${output_dir}/agent.psk"
        ;;
esac

docker compose \
    -p "$project" --project-directory "$repo_dir" \
    -f "${repo_dir}/compose.yaml" -f "${script_dir}/compose.${mode}.yaml" \
    up -d --force-recreate --remove-orphans \
    --wait --wait-timeout 180 \
    zabbix-proxy-sqlite3 zabbix-agent

printf 'Stop: docker compose -p %s down --remove-orphans\n' "$project"
