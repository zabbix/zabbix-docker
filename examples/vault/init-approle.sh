#!/bin/sh

set -eu

MYSQL_USER="$(cat /run/secrets/MYSQL_USER)"
MYSQL_PASSWORD="$(cat /run/secrets/MYSQL_PASSWORD)"

vault secrets list | grep -q '^zabbix/' || vault secrets enable -path=zabbix kv-v2
vault kv put zabbix/database \
	username="${MYSQL_USER}" \
	password="${MYSQL_PASSWORD}"

vault policy write zabbix-database /tools/policy.hcl
vault auth list | grep -q '^approle/' || vault auth enable approle
vault write auth/approle/role/zabbix \
	token_policies=zabbix-database \
	token_ttl=1h \
	token_max_ttl=4h
vault write auth/approle/role/zabbix/role-id \
	role_id="${ZBX_VAULTAPPROLEID}"
vault write auth/approle/role/zabbix/secret-id/destroy \
	secret_id="${ZBX_VAULTAPPSECRETID}" >/dev/null
vault write auth/approle/role/zabbix/custom-secret-id \
	secret_id="${ZBX_VAULTAPPSECRETID}" >/dev/null
