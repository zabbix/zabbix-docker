#!/bin/sh

set -eu

MYSQL_USER="$(cat /run/secrets/MYSQL_USER)"
MYSQL_PASSWORD="$(cat /run/secrets/MYSQL_PASSWORD)"

vault secrets list | grep -q '^zabbix/' || vault secrets enable -path=zabbix kv-v2
vault kv put zabbix/database \
	username="${MYSQL_USER}" \
	password="${MYSQL_PASSWORD}"

vault policy write zabbix-database /tools/policy.hcl
vault token lookup "${ZBX_VAULT_TOKEN}" >/dev/null 2>&1 || vault token create \
	-id="${ZBX_VAULT_TOKEN}" \
	-policy=zabbix-database \
	-ttl=4h >/dev/null
