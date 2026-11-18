#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/../.." && pwd)"
project=zabbix-web-tls

case "${1:-}" in
    nginx)
        output_path=etc/ssl/nginx
        web_service=zabbix-web-nginx-mysql
        ;;
    apache)
        output_path=etc/ssl/apache2
        web_service=zabbix-web-apache-mysql
        ;;
    *)
        printf 'Usage: %s nginx|apache\n' "$0" >&2
        exit 1
        ;;
esac

data_dir="${DATA_DIRECTORY:-$(sed -n 's/^DATA_DIRECTORY=//p' "${repo_dir}/.env")}"
: "${data_dir:?DATA_DIRECTORY is not set}"
cd "${repo_dir}"
output_dir="${data_dir}/${output_path}"
mkdir -p "${output_dir}"
umask 077

openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 365 \
    -subj '/CN=localhost' \
    -addext 'basicConstraints=critical,CA:FALSE' \
    -addext 'extendedKeyUsage=serverAuth' \
    -addext 'subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1' \
    -keyout "${output_dir}/ssl.key" \
    -out "${output_dir}/ssl.crt"

chmod 0644 "${output_dir}/ssl.crt"
chmod 0600 "${output_dir}/ssl.key"

if [[ "${1}" = nginx ]]; then
    openssl genpkey -genparam -algorithm DH -pkeyopt group:ffdhe2048 \
        -out "${output_dir}/dhparam.pem"
    chmod 0644 "${output_dir}/dhparam.pem"
fi

docker compose \
    -p "$project" --project-directory "$repo_dir" \
    -f "${repo_dir}/compose.yaml" \
    up -d --force-recreate --remove-orphans \
    --wait --wait-timeout 180 \
    zabbix-server "$web_service"

printf 'Stop: docker compose -p %s down --remove-orphans\n' "$project"
