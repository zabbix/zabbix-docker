# Enable PostgreSQL timescaleDB feature
: "${ENABLE_TIMESCALEDB:=false}"

set_pg_env() {
    [ -n "${DB_SERVER_ZBX_PASS:-}" ] && export PGPASSWORD="${DB_SERVER_ZBX_PASS}"

    if [ "${POSTGRES_USE_IMPLICIT_SEARCH_PATH,,}" = "false" ] && [ -n "${DB_SERVER_SCHEMA:-}" ]; then
        export PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
    fi

    if [ -n "${ZBX_DBTLSCONNECT:-}" ]; then
        local pg_sslmode
        pg_sslmode="${ZBX_DBTLSCONNECT//_/-}"
        export PGSSLMODE="${pgsslmode//required/require}"
        export PGSSLROOTCERT="${ZBX_DBTLSCAFILE:-}"
        export PGSSLCERT="${ZBX_DBTLSCERTFILE:-}"
        export PGSSLKEY="${ZBX_DBTLSKEYFILE:-}"
    fi
}

clear_pg_env() {
    unset PGPASSWORD PGOPTIONS PGSSLMODE PGSSLROOTCERT PGSSLCERT PGSSLKEY
}

# Check prerequisites for PostgreSQL database
check_db_variables() {
    local default_db_name="${1:-}"

    : "${DB_SERVER_HOST=postgres-server}"
    : "${DB_SERVER_PORT:=5432}"
    : "${DB_SERVER_SCHEMA:=public}"
    : "${POSTGRES_USE_IMPLICIT_SEARCH_PATH:=false}"

    file_env POSTGRES_USER
    file_env POSTGRES_PASSWORD

    DB_SERVER_ROOT_USER="${POSTGRES_USER:-postgres}"
    DB_SERVER_ROOT_PASS="${POSTGRES_PASSWORD:-}"

    DB_SERVER_ZBX_USER="${POSTGRES_USER:-zabbix}"
    DB_SERVER_ZBX_PASS="${POSTGRES_PASSWORD:-zabbix}"

    DB_SERVER_DBNAME="${POSTGRES_DB:-$default_db_name}"

    psql_connect_args=(--port "${DB_SERVER_PORT}")
    [ -n "${DB_SERVER_HOST}" ] && psql_connect_args=(--host "${DB_SERVER_HOST}")
}

get_vault_secrets() {
    local wait_timeout=5
    local curl_opts=(-s -m 10 -k)
    local vaultdata errors
    local cyberark_opts

    if [ -z "${ZBX_VAULTURL:-}" ] || [ -z "${ZBX_VAULTPREFIX:-}" ] || [ -z "${ZBX_VAULTDBPATH:-}" ]; then
        error "Missing variables! If ZBX_VAULT is used then ZBX_VAULTURL, ZBX_VAULTPREFIX and ZBX_VAULTDBPATH must be set"
    fi
    local vault_url="${ZBX_VAULTURL}${ZBX_VAULTPREFIX}${ZBX_VAULTDBPATH}"

    if [ "${ZBX_VAULT:-}" = "HashiCorp" ]; then
        while ! vaultdata="$(curl "${curl_opts[@]}" -H "X-Vault-Token: $VAULT_TOKEN" "$vault_url")"; do
            info "**** Vault is not available. Waiting ${wait_timeout} seconds... ****"
            sleep "$wait_timeout"
        done
        errors="$(printf '%s' "$vaultdata" | jq -r '.errors // empty')"
        if [ -n "${errors}" ]; then
            error "Error getting secrets from vault: $errors"
        fi
        DB_SERVER_ZBX_USER="$(printf '%s' "$vaultdata" | jq -r '.data.data.username')"
        DB_SERVER_ZBX_PASS="$(printf '%s' "$vaultdata" | jq -r '.data.data.password')"

    elif [ "${ZBX_VAULT:-}" = "CyberArk" ]; then
        cyberark_opts=(-H "Content-type: application/json" --cert "$ZBX_VAULTCERTFILE")

        # if key is defined use it
        if [ -n "${ZBX_VAULTKEYFILE:-}" ]; then
            cyberark_opts+=(--key "$ZBX_VAULTKEYFILE")
        fi
        while ! vaultdata="$(curl "${curl_opts[@]}" "${cyberark_opts[@]}" "$vault_url")"; do
            info "**** Vault is not available. Waiting ${wait_timeout} seconds... ****"
            sleep "$wait_timeout"
        done

        errors=$(printf '%s' "$vaultdata" | jq -r '.ErrorCode // empty')
        if [ -n "${errors}" ]; then
            error "Error getting secrets from vault: $errors"
        fi

        DB_SERVER_ZBX_USER="$(printf '%s' "$vaultdata" | jq -r '.UserName')"
        DB_SERVER_ZBX_PASS="$(printf '%s' "$vaultdata" | jq -r '.Content')"

    else
        error "ZBX_VAULT has wrong value. HashiCorp or CyberArk are supported!"
    fi
}

check_db_connect() {
    local use_vault="${1:-false}"
    local wait_timeout=5

    info "********************"
    if [ -n "${DB_SERVER_HOST}" ]; then
        info "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
        info "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    else
        info "* DB_SERVER_HOST: Using DB socket"
        info "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    fi
    info "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    info "* DB_SERVER_SCHEMA: ${DB_SERVER_SCHEMA}"
    if [ "${DEBUG_MODE:-}" = "true" ]; then
        info "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
        info "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    fi
    info "********************"

    if [ -n "${ZBX_VAULT:-}" ] && [ "$use_vault" = "true" ]; then
        unset DB_SERVER_ZBX_USER
        unset DB_SERVER_ZBX_PASS

        info "***** Connecting to vault... ******"
        info "***** VAULT URL: $ZBX_VAULTURL"
        get_vault_secrets
    fi

    set_pg_env

    while :; do
        psql "${psql_connect_args[@]}" --username "${DB_SERVER_ROOT_USER}" --list --quiet >/dev/null 2>&1 && break
        psql "${psql_connect_args[@]}" --username "${DB_SERVER_ROOT_USER}" --list --dbname "${DB_SERVER_DBNAME}" --quiet >/dev/null 2>&1 && break

        info "**** PostgreSQL server is not available. Waiting ${wait_timeout} seconds..."
        sleep "${wait_timeout}"
    done

    clear_pg_env
}

psql_query() {
    local query="${1:-}"
    local db="${2:-}"
    local result=""

    set_pg_env

    result="$({
        psql --no-align --quiet --tuples-only \
            "${psql_connect_args[@]}" \
            --username "${DB_SERVER_ROOT_USER}" \
            --command "$query" \
            --dbname "$db" 2>/dev/null
    })"

    clear_pg_env
    printf '%s\n' "$result"
}

exec_sql_file() {
    local sql_script="${1:-}"
    local command="cat"

    set_pg_env

    if [ "${sql_script: -3}" = ".gz" ]; then
        command="zcat"
    fi

    "$command" "$sql_script" | psql --quiet \
        "${psql_connect_args[@]}" \
        --username "${DB_SERVER_ZBX_USER}" \
        --dbname "${DB_SERVER_DBNAME}" >/dev/null || exit 1

    clear_pg_env
}

create_db_database() {
    local db_exists

    db_exists="$(psql_query "SELECT 1 AS result FROM pg_database WHERE datname='${DB_SERVER_DBNAME}'" "${DB_SERVER_DBNAME}")"

    if [ -z "${db_exists}" ]; then
        info "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."

        set_pg_env
        createdb "${psql_connect_args[@]}" \
            --username "${DB_SERVER_ROOT_USER}" \
            --owner "${DB_SERVER_ZBX_USER}" \
            --lc-ctype "en_US.utf8" \
            --lc-collate "en_US.utf8" \
            "${DB_SERVER_DBNAME}"
        clear_pg_env
    else
        info "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database owner!"
    fi

    psql_query "CREATE SCHEMA IF NOT EXISTS ${DB_SERVER_SCHEMA}" "${DB_SERVER_DBNAME}" >/dev/null
}

apply_db_scripts() {
    local sql_script

    shopt -s nullglob
    for sql_script in "${ZABBIX_USER_HOME_DIR}"/dbscripts/*.sql; do
        info "** Processing additional '${sql_script}' SQL script"
        exec_sql_file "$sql_script"
    done
    shopt -u nullglob
}

create_db_schema() {
    local db_schema_file="${1:-}"
    local dbversion_table_exists

    dbversion_table_exists="$(psql_query "SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid =
                                         c.relnamespace WHERE  n.nspname = '$DB_SERVER_SCHEMA' AND c.relname = 'dbversion'" "${DB_SERVER_DBNAME}")"

    if [ -n "${dbversion_table_exists}" ]; then
        info "** Table '${DB_SERVER_DBNAME}.dbversion' already exists."
        ZBX_DB_VERSION="$(psql_query "SELECT mandatory FROM ${DB_SERVER_SCHEMA}.dbversion" "${DB_SERVER_DBNAME}")"
    fi

    if [ -z "${ZBX_DB_VERSION:-}" ]; then
        info "** Creating '${DB_SERVER_DBNAME}' schema in PostgreSQL"

        exec_sql_file "${db_schema_file}"

        if [ "${ENABLE_TIMESCALEDB,,}" = "true" ]; then
            psql_query "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;" "${DB_SERVER_DBNAME}" >/dev/null
            exec_sql_file "/usr/share/doc/zabbix-server-postgresql/timescaledb.sql"
        fi

        apply_db_scripts
    fi
}
