: "${DB_CHARACTER_SET:=utf8mb4}"
: "${DB_CHARACTER_COLLATE:=utf8mb4_bin}"

[ -n "${DB_ENGINE:-}" ] || error "DB_ENGINE is not set. Expected 'mysql' or 'mariadb'"

source "${ENTRYPOINT_LIBS}/logging.sh"
source "${ENTRYPOINT_LIBS}/config.sh"

set_mysql_cli() {
    case "${DB_ENGINE}" in
        mysql)
            MYSQL_CLI_BIN="mysql"
            MYSQL_ADMIN_BIN="mysqladmin"
            MYSQL_EXTRA_ARGS=()
            ;;
        mariadb)
            MYSQL_CLI_BIN="mariadb"
            MYSQL_ADMIN_BIN="mariadb-admin"
            MYSQL_EXTRA_ARGS=(--skip-ssl-verify-server-cert)
            ;;
        *)
            error "Unsupported DB_ENGINE: '${DB_ENGINE}'. Expected 'mysql' or 'mariadb'"
            ;;
    esac
}

set_mysql_tls_args() {
    MYSQL_TLS_ARGS=()

    if [ -n "${ZBX_DBTLSCONNECT:-}" ]; then
        if [ "${DB_ENGINE}" = "mariadb" ]; then
            MYSQL_TLS_ARGS+=(--ssl)

            if [ "${ZBX_DBTLSCONNECT}" != "required" ]; then
                MYSQL_TLS_ARGS+=(--ssl-verify-server-cert)
            fi
        else
            local ssl_mode="${ZBX_DBTLSCONNECT//verify_full/verify_identity}"
            MYSQL_TLS_ARGS+=("--ssl=${ssl_mode}")
        fi

        [ -n "${ZBX_DBTLSCAFILE:-}" ] && MYSQL_TLS_ARGS+=("--ssl-ca=${ZBX_DBTLSCAFILE}")
        [ -n "${ZBX_DBTLSKEYFILE:-}" ] && MYSQL_TLS_ARGS+=("--ssl-key=${ZBX_DBTLSKEYFILE}")
        [ -n "${ZBX_DBTLSCERTFILE:-}" ] && MYSQL_TLS_ARGS+=("--ssl-cert=${ZBX_DBTLSCERTFILE}")
    fi
}

set_mysql_auth_env() {
    export MYSQL_PWD="${DB_SERVER_ROOT_PASS:-}"
}

clear_mysql_auth_env() {
    unset MYSQL_PWD
}

# Check prerequisites for MySQL-compatible database
check_db_variables() {
    local default_db_name="${1:-}"

    if [ -n "${DB_SERVER_SOCKET:-}" ]; then
        mysql_connect_args=("-S" "${DB_SERVER_SOCKET}")
    else
        : "${DB_SERVER_HOST:=mysql-server}"
        : "${DB_SERVER_PORT:=3306}"
        mysql_connect_args=("-h" "${DB_SERVER_HOST}" "-P" "${DB_SERVER_PORT}")
    fi

    USE_DB_ROOT_USER=false
    CREATE_ZBX_DB_USER=false

    file_env MYSQL_USER
    file_env MYSQL_PASSWORD
    file_env MYSQL_ROOT_USER
    file_env MYSQL_ROOT_PASSWORD

    if [ -z "${MYSQL_USER:-}" ] && [ "${MYSQL_RANDOM_ROOT_PASSWORD:-}" = "true" ]; then
        error "**** Impossible to use MySQL server because of unknown Zabbix user and random 'root' password"
    fi

    if [ -z "${MYSQL_USER:-}" ] && [ -z "${MYSQL_ROOT_PASSWORD:-}" ] && [ "${MYSQL_ALLOW_EMPTY_PASSWORD:-}" != "true" ]; then
        error "*** Impossible to use MySQL server because 'root' password is not defined and it is not empty"
    fi

    if [ "${MYSQL_ALLOW_EMPTY_PASSWORD:-}" = "true" ] || [ -n "${MYSQL_ROOT_PASSWORD:-}" ]; then
        USE_DB_ROOT_USER=true
        DB_SERVER_ROOT_USER="${MYSQL_ROOT_USER:-root}"
        DB_SERVER_ROOT_PASS="${MYSQL_ROOT_PASSWORD:-}"
    fi

    [ -n "${MYSQL_USER:-}" ] && [ "${USE_DB_ROOT_USER}" = "true" ] && CREATE_ZBX_DB_USER=true

    # If root password is not specified use provided credentials
    : "${DB_SERVER_ROOT_USER:=${MYSQL_USER:-}}"
    if [ "${MYSQL_ALLOW_EMPTY_PASSWORD:-}" != "true" ]; then
        DB_SERVER_ROOT_PASS="${DB_SERVER_ROOT_PASS:-${MYSQL_PASSWORD:-}}"
    fi

    DB_SERVER_ZBX_USER="${MYSQL_USER:-zabbix}"
    DB_SERVER_ZBX_PASS="${MYSQL_PASSWORD:-zabbix}"
    DB_SERVER_DBNAME="${MYSQL_DATABASE:-$default_db_name}"

    set_mysql_cli
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
    if [ -n "${DB_SERVER_SOCKET:-}" ]; then
        info "* DB_SERVER_SOCKET: ${DB_SERVER_SOCKET}"
    else
        info "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
        info "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    fi
    info "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"

    if [ "${DEBUG_MODE:-}" = "true" ]; then
        if [ "${USE_DB_ROOT_USER:-}" = "true" ]; then
            info "* DB_SERVER_ROOT_USER: ${DB_SERVER_ROOT_USER}"
            info "* DB_SERVER_ROOT_PASS: ${DB_SERVER_ROOT_PASS}"
        fi
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

    set_mysql_tls_args
    set_mysql_auth_env

    while ! "$MYSQL_ADMIN_BIN" ping \
        "${mysql_connect_args[@]}" \
        -u "${DB_SERVER_ROOT_USER}" \
        --silent \
        "${MYSQL_EXTRA_ARGS[@]}" \
        --connect_timeout=10 \
        "${MYSQL_TLS_ARGS[@]}" >/dev/null 2>&1; do
        info "**** MySQL server is not available. Waiting ${wait_timeout} seconds..."
        sleep "$wait_timeout"
    done

    clear_mysql_auth_env
}

mysql_query() {
    local query="${1:-}"
    local result=""

    set_mysql_tls_args
    set_mysql_auth_env

    result="$(
        {
            "$MYSQL_CLI_BIN" \
                --silent \
                --skip-column-names \
                "${MYSQL_EXTRA_ARGS[@]}" \
                "${mysql_connect_args[@]}" \
                -u "${DB_SERVER_ROOT_USER}" \
                -e "$query" \
                "${MYSQL_TLS_ARGS[@]}"
        } 2>/dev/null
    )"

    clear_mysql_auth_env
    printf '%s\n' "$result"
}

exec_sql_file() {
    local sql_script="${1:-}"
    local command="cat"

    set_mysql_tls_args
    set_mysql_auth_env

    [ "${sql_script: -3}" = ".gz" ] && command="zcat"

    "$command" "$sql_script" | "$MYSQL_CLI_BIN" \
        --silent \
        --skip-column-names \
        "${MYSQL_EXTRA_ARGS[@]}" \
        --default-character-set="${DB_CHARACTER_SET}" \
        "${mysql_connect_args[@]}" \
        -u "${DB_SERVER_ROOT_USER}" \
        "${MYSQL_TLS_ARGS[@]}" \
        "${DB_SERVER_DBNAME}" >/dev/null

    clear_mysql_auth_env
}

create_db_user() {
    [ "${CREATE_ZBX_DB_USER}" = "true" ] || return 0

    info "** Creating '${DB_SERVER_ZBX_USER}' user in MySQL database"

    local user_exists
    user_exists="$(mysql_query "SELECT 1 FROM mysql.user WHERE user = '${DB_SERVER_ZBX_USER}' AND host = '%'")"

    if [ -z "$user_exists" ]; then
        mysql_query "CREATE USER '${DB_SERVER_ZBX_USER}'@'%' IDENTIFIED BY '${DB_SERVER_ZBX_PASS}'" >/dev/null
    else
        mysql_query "ALTER USER '${DB_SERVER_ZBX_USER}'@'%' IDENTIFIED BY '${DB_SERVER_ZBX_PASS}'" >/dev/null
    fi

    mysql_query "GRANT ALL PRIVILEGES ON ${DB_SERVER_DBNAME}.* TO '${DB_SERVER_ZBX_USER}'@'%'" >/dev/null
}

create_db_database() {
    local db_exists
    db_exists="$(mysql_query "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='${DB_SERVER_DBNAME}'")"

    if [ -z "${db_exists}" ]; then
        info "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."
        mysql_query "CREATE DATABASE ${DB_SERVER_DBNAME} CHARACTER SET ${DB_CHARACTER_SET} COLLATE ${DB_CHARACTER_COLLATE}" >/dev/null
        mysql_query "GRANT ALL PRIVILEGES ON ${DB_SERVER_DBNAME}.* TO '${DB_SERVER_ZBX_USER}'@'%'" >/dev/null
    else
        info "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database COLLATE!"
    fi
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

    dbversion_table_exists="$(mysql_query "SELECT 1 FROM information_schema.tables WHERE table_schema='${DB_SERVER_DBNAME}' and table_name='dbversion'")"

    if [ -n "${dbversion_table_exists}" ]; then
        warn "** Table '${DB_SERVER_DBNAME}.dbversion' already exists."
        ZBX_DB_VERSION="$(mysql_query "SELECT mandatory FROM ${DB_SERVER_DBNAME}.dbversion")"
    fi

    if [ -z "${ZBX_DB_VERSION:-}" ]; then
        info "** Creating '${DB_SERVER_DBNAME}' schema in MySQL"
        exec_sql_file "${db_schema_file}"
        info "** Database schema successfully created!"

        apply_db_scripts
    fi
}
