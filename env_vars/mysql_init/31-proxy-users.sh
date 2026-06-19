#!/usr/bin/env bash
set -e

docker_process_sql --database=mysql <<-EOSQL
CREATE USER IF NOT EXISTS '${ZBX_PROXY_DB_USER}'@'${ZBX_PROXY_DB_USER_HOST}'
  $(identified_by_sql "$ZBX_PROXY_DB_PASSWORD_SQL");
GRANT '${ZBX_PROXY_DB_ROLE}' TO '${ZBX_PROXY_DB_USER}'@'${ZBX_PROXY_DB_USER_HOST}';
$(default_role_sql "$ZBX_PROXY_DB_ROLE" "$ZBX_PROXY_DB_USER" "$ZBX_PROXY_DB_USER_HOST")

FLUSH PRIVILEGES;
EOSQL
