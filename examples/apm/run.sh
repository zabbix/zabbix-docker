#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/../.." && pwd)"
project=zabbix-apm
compose_args=(-f "${repo_dir}/compose.yaml" -f "${script_dir}/compose.yaml")

case "${1:-}" in
    "") ;;
    --cluster)
        project=zabbix-apm-cluster
        compose_args+=(-f "${script_dir}/compose.cluster.yaml")
        ;;
    *)
        printf 'Usage: %s [--cluster]\n' "$0" >&2
        exit 1
        ;;
esac

docker compose \
    -p "$project" --project-directory "$repo_dir" \
    "${compose_args[@]}" \
    --profile clickhouse \
    up -d --build \
    --wait --wait-timeout 180 \
    clickhouse server-telemetry-init \
    zabbix-server zabbix-proxy-sqlite3 zabbix-web-nginx-mysql \
    demo-inventory demo-checkout demo-loadgen

printf 'Stop: docker compose -p %s down --remove-orphans\n' "$project"
