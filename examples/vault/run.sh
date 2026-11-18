#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/../.." && pwd)"

case "${1:---approle}" in
    --approle)
        mode=approle
        project=zabbix-vault-approle
        ;;
    --token)
        mode=token
        project=zabbix-vault-token
        ;;
    *)
        printf 'Usage: %s [--approle|--token]\n' "$0" >&2
        exit 1
        ;;
esac

docker compose \
    -p "$project" --project-directory "$repo_dir" \
    -f "${repo_dir}/compose.yaml" \
    -f "${script_dir}/compose.yaml" \
    -f "${script_dir}/compose.${mode}.yaml" \
    up -d \
    --wait --wait-timeout 180 \
    vault zabbix-server zabbix-web-nginx-mysql

printf 'Stop: docker compose -p %s down --remove-orphans\n' "$project"
