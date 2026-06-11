# shellcheck shell=bash

source "${ENTRYPOINT_LIBS}/bootstrap.sh"
source "${ENTRYPOINT_LIBS}/openssl.sh"

# Internal directory for TLS related files, used when TLS*File specified as plain text values
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

proxy_config() {
    local default_host_name="${1:-}"

    : "${ZBX_SERVER_HOST:=zabbix-server}"
    export ZBX_SERVER_HOST

    if [ -z "${ZBX_HOSTNAME:-}" ] && [ -n "${ZBX_HOSTNAMEITEM:-}" ]; then
        export ZBX_HOSTNAME=""
    else
        export ZBX_HOSTNAME="${ZBX_HOSTNAME:-$default_host_name}"
    fi

    : "${ZBX_ENABLE_SNMP_TRAPS:=false}"
    [[ "${ZBX_ENABLE_SNMP_TRAPS,,}" == "true" ]] && export ZBX_STARTSNMPTRAPPER=1
    unset ZBX_ENABLE_SNMP_TRAPS

    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf" "LoadModule" "${ZBX_LOADMODULE:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCAFILE" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCRLFILE" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCERTFILE" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSKEYFILE" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSPSKFILE" "${ZBX_TLSPSKFILE:-}" "${ZBX_TLSPSK:-}"

    if [ "$(id -u)" -ne 0 ]; then
        ZBX_USER="$(id -un)"
        export ZBX_USER
    else
        export ZBX_ALLOWROOT=1
    fi

    openssl_rehash "${ZBX_SSLCALOCATION}"
}
