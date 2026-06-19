#!/usr/bin/env bash
set -e

docker_process_sql --dbname "$ZBX_SERVER_DB_NAME" \
    --set=server_role="$ZBX_SERVER_DB_ROLE" \
    --set=server_user="$ZBX_SERVER_DB_USER" \
    --set=server_password="$ZBX_SERVER_DB_PASSWORD" \
    --set=web_role="$ZBX_WEB_DB_ROLE" \
    --set=web_user="$ZBX_WEB_DB_USER" \
    --set=web_password="$ZBX_WEB_DB_PASSWORD" <<-'EOSQL'
SELECT format('CREATE USER %I WITH ENCRYPTED PASSWORD %L', :'server_user', :'server_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'server_user') \gexec
GRANT :"server_role" TO :"server_user";

SELECT format('CREATE USER %I WITH ENCRYPTED PASSWORD %L', :'web_user', :'web_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'web_user') \gexec
GRANT :"web_role" TO :"web_user";
EOSQL
