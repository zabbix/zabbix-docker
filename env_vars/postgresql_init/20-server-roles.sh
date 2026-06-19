#!/usr/bin/env bash
set -e

docker_process_sql --dbname "$ZBX_SERVER_DB_NAME" \
    --set=db="$ZBX_SERVER_DB_NAME" \
    --set=schema="$ZBX_SERVER_DB_SCHEMA" \
    --set=server_user="$ZBX_SERVER_DB_USER" \
    --set=server_role="$ZBX_SERVER_DB_ROLE" \
    --set=web_role="$ZBX_WEB_DB_ROLE" <<-'EOSQL'
SELECT format('CREATE ROLE %I', :'server_role')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'server_role') \gexec

GRANT CONNECT ON DATABASE :"db" TO :"server_role";
GRANT USAGE, CREATE ON SCHEMA :"schema" TO :"server_role";
ALTER DEFAULT PRIVILEGES FOR ROLE :"server_user" IN SCHEMA :"schema"
    GRANT DELETE, INSERT, SELECT, UPDATE ON TABLES TO :"server_role";
ALTER DEFAULT PRIVILEGES FOR ROLE :"server_user" IN SCHEMA :"schema"
    GRANT SELECT, UPDATE, USAGE ON SEQUENCES TO :"server_role";

SELECT format('CREATE ROLE %I', :'web_role')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'web_role') \gexec

GRANT CONNECT ON DATABASE :"db" TO :"web_role";
GRANT USAGE ON SCHEMA :"schema" TO :"web_role";
ALTER DEFAULT PRIVILEGES FOR ROLE :"server_user" IN SCHEMA :"schema"
    GRANT DELETE, INSERT, SELECT, UPDATE ON TABLES TO :"web_role";
ALTER DEFAULT PRIVILEGES FOR ROLE :"server_user" IN SCHEMA :"schema"
    GRANT SELECT, UPDATE, USAGE ON SEQUENCES TO :"web_role";
EOSQL
