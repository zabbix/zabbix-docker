#!/usr/bin/env bash
set -e

docker_process_sql --dbname postgres \
    --set=server_user="$ZBX_SERVER_DB_USER" \
    --set=server_password="$ZBX_SERVER_DB_PASSWORD" \
    --set=db="$ZBX_SERVER_DB_NAME" <<-'EOSQL'
SELECT format('CREATE USER %I WITH ENCRYPTED PASSWORD %L', :'server_user', :'server_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'server_user') \gexec

SELECT format('CREATE DATABASE %I OWNER %I ENCODING %L TEMPLATE template0', :'db', :'server_user', 'UTF8')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'db') \gexec

ALTER DATABASE :"db" OWNER TO :"server_user";
EOSQL

docker_process_sql --dbname "$ZBX_SERVER_DB_NAME" \
    --set=db="$ZBX_SERVER_DB_NAME" \
    --set=schema="$ZBX_SERVER_DB_SCHEMA" \
    --set=server_user="$ZBX_SERVER_DB_USER" <<-'EOSQL'
SELECT format('CREATE SCHEMA %I AUTHORIZATION %I', :'schema', :'server_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = :'schema') \gexec

ALTER SCHEMA :"schema" OWNER TO :"server_user";
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE :"db" FROM PUBLIC;
GRANT CONNECT, TEMPORARY ON DATABASE :"db" TO :"server_user";
GRANT USAGE, CREATE ON SCHEMA :"schema" TO :"server_user";
ALTER ROLE ALL IN DATABASE :"db" SET search_path = :"schema";
EOSQL
