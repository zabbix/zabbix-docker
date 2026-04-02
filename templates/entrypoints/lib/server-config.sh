source "${ENTRYPOINT_LIBS}/bootstrap.sh"
source "${ENTRYPOINT_LIBS}/openssl.sh"

# Internal directory for TLS related files, used when TLS*File specified as plain text values
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

server_config() {
    : "${ZBX_ENABLE_SNMP_TRAPS:=false}"
    [[ "${ZBX_ENABLE_SNMP_TRAPS,,}" == "true" ]] && export ZBX_STARTSNMPTRAPPER=1
    unset ZBX_ENABLE_SNMP_TRAPS

    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_server_modules.conf" "LoadModule" "${ZBX_LOADMODULE:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCAFILE" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCRLFILE" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCERTFILE" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSKEYFILE" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"

    if [ "${ZBX_AUTOHANODENAME:-}" == 'fqdn' ] && [ ! -n "${ZBX_HANODENAME:-}" ]; then
        export ZBX_HANODENAME="$(hostname -f)"
    elif [ "${ZBX_AUTOHANODENAME:-}" == 'hostname' ] && [ ! -n "${ZBX_HANODENAME:-}" ]; then
        export ZBX_HANODENAME="$(hostname)"
    fi
    unset ZBX_AUTOHANODENAME

    : "${ZBX_NODEADDRESSPORT:=10051}"
    if [ "${ZBX_AUTONODEADDRESS:-}" == 'fqdn' ] && [ ! -n "${ZBX_NODEADDRESS:-}" ]; then
        export ZBX_NODEADDRESS="$(hostname -f):${ZBX_NODEADDRESSPORT}"
    elif [ "${ZBX_AUTONODEADDRESS:-}" == 'hostname' ] && [ ! -n "${ZBX_NODEADDRESS:-}" ]; then
        export ZBX_NODEADDRESS="$(hostname):${ZBX_NODEADDRESSPORT}"
    fi
    unset ZBX_AUTONODEADDRESS

    if [ "$(id -u)" -ne 0 ]; then
        export ZBX_USER="$(id -un)"
    else
        export ZBX_ALLOWROOT=1
    fi

    openssl_rehash "${ZBX_SSLCALOCATION}"
}
