prepare_php_config() {
    local db_server_type="${1:-}"

    info "** Preparing PHP configuration"

    : "${PHP_FPM_PM:=dynamic}"
    : "${PHP_FPM_PM_MAX_CHILDREN:=50}"
    : "${PHP_FPM_PM_START_SERVERS:=5}"
    : "${PHP_FPM_PM_MIN_SPARE_SERVERS:=5}"
    : "${PHP_FPM_PM_MAX_SPARE_SERVERS:=35}"
    : "${PHP_FPM_PM_MAX_REQUESTS:=0}"

    export PHP_FPM_PM
    export PHP_FPM_PM_MAX_CHILDREN
    export PHP_FPM_PM_START_SERVERS
    export PHP_FPM_PM_MIN_SPARE_SERVERS
    export PHP_FPM_PM_MAX_SPARE_SERVERS
    export PHP_FPM_PM_MAX_REQUESTS

    if [ "$(id -u)" -eq 0 ]; then
        {
            echo "user = ${DAEMON_USER}"
            echo "group = ${DAEMON_GROUP}"
            echo "listen.owner = ${DAEMON_USER}"
            echo "listen.group = ${DAEMON_GROUP}"
        } >> "$PHP_CONFIG_FILE"
    fi

    : "${ZBX_DENY_GUI_ACCESS:=false}"
    : "${ZBX_GUI_ACCESS_IP_RANGE:=['127.0.0.1']}"
    : "${ZBX_GUI_WARNING_MSG:=Zabbix is under maintenance.}"

    : "${ZBX_MAXEXECUTIONTIME:=600}"
    : "${ZBX_MEMORYLIMIT:=128M}"
    : "${ZBX_POSTMAXSIZE:=16M}"
    : "${ZBX_UPLOADMAXFILESIZE:=2M}"
    : "${ZBX_MAXINPUTTIME:=300}"
    : "${PHP_TZ:=Europe/Riga}"

    export ZBX_DENY_GUI_ACCESS="${ZBX_DENY_GUI_ACCESS,,}"
    export ZBX_GUI_ACCESS_IP_RANGE
    export ZBX_GUI_WARNING_MSG

    export ZBX_MAXEXECUTIONTIME
    export ZBX_MEMORYLIMIT
    export ZBX_POSTMAXSIZE
    export ZBX_UPLOADMAXFILESIZE
    export ZBX_MAXINPUTTIME
    export PHP_TZ

    export DB_SERVER_TYPE="${db_server_type}"
    [ -n "${DB_SERVER_HOST:-}" ] && export DB_SERVER_HOST
    [ -n "${DB_SERVER_PORT:-}" ] && export DB_SERVER_PORT
    [ -n "${DB_SERVER_SOCKET:-}" ] && export DB_SERVER_SOCKET

    export DB_SERVER_DBNAME="${DB_SERVER_DBNAME}"
    export DB_SERVER_SCHEMA="${DB_SERVER_SCHEMA:-}"
    export DB_SERVER_USER="${DB_SERVER_ZBX_USER:-}"
    export DB_SERVER_PASS="${DB_SERVER_ZBX_PASS:-}"
    export ZBX_SERVER_HOST="${ZBX_SERVER_HOST}"
    export ZBX_SERVER_PORT="${ZBX_SERVER_PORT}"
    export ZBX_SERVER_NAME="${ZBX_SERVER_NAME}"

    : "${ZBX_DB_ENCRYPTION:=false}"
    : "${ZBX_DB_VERIFY_HOST:=false}"

    export ZBX_DB_ENCRYPTION="${ZBX_DB_ENCRYPTION,,}"
    export ZBX_DB_KEY_FILE="${ZBX_DB_KEY_FILE:-}"
    export ZBX_DB_CERT_FILE="${ZBX_DB_CERT_FILE:-}"
    export ZBX_DB_CA_FILE="${ZBX_DB_CA_FILE:-}"
    export ZBX_DB_VERIFY_HOST="${ZBX_DB_VERIFY_HOST,,}"

    export ZBX_VAULT="${ZBX_VAULT:-}"
    export ZBX_VAULTURL="${ZBX_VAULTURL:-}"
    export ZBX_VAULTPREFIX="${ZBX_VAULTPREFIX:-}"
    export ZBX_VAULTDBPATH="${ZBX_VAULTDBPATH:-}"
    export VAULT_TOKEN="${VAULT_TOKEN:-}"
    export ZBX_VAULTCERTFILE="${ZBX_VAULTCERTFILE:-}"
    export ZBX_VAULTKEYFILE="${ZBX_VAULTKEYFILE:-}"

    : "${DB_DOUBLE_IEEE754:=true}"
    export DB_DOUBLE_IEEE754="${DB_DOUBLE_IEEE754,,}"

    export ZBX_HISTORYSTORAGEURL="${ZBX_HISTORYSTORAGEURL:-}"
    export ZBX_HISTORYSTORAGETYPES="${ZBX_HISTORYSTORAGETYPES:-[]}"

    export ZBX_SSO_SETTINGS="${ZBX_SSO_SETTINGS:-}"
    export ZBX_SSO_SP_KEY="${ZBX_SSO_SP_KEY:-}"
    export ZBX_SSO_SP_CERT="${ZBX_SSO_SP_CERT:-}"
    export ZBX_SSO_IDP_CERT="${ZBX_SSO_IDP_CERT:-}"

    : "${ZBX_ALLOW_HTTP_AUTH:=true}"
    export ZBX_ALLOW_HTTP_AUTH

    : "${ZBX_SERVER_TLS_ACTIVE:=0}"
    export ZBX_SERVER_TLS_ACTIVE

    file_process_from_env "ZBX_SERVER_TLS_CAFILE" "${ZBX_SERVER_TLS_CAFILE:-}" "${ZBX_SERVER_TLS_CA:-}"
    file_process_from_env "ZBX_SERVER_TLS_KEYFILE" "${ZBX_SERVER_TLS_KEYFILE:-}" "${ZBX_SERVER_TLS_KEY:-}"
    file_process_from_env "ZBX_SERVER_TLS_CERTFILE" "${ZBX_SERVER_TLS_CERTFILE:-}" "${ZBX_SERVER_TLS_CERT:-}"

    export ZBX_SERVER_TLS_CERT_ISSUER="${ZBX_SERVER_TLS_CERT_ISSUER:-}"
    export ZBX_SERVER_TLS_CERT_SUBJECT="${ZBX_SERVER_TLS_CERT_SUBJECT:-}"
}
