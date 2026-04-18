#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

readonly ZBX_WEB_SERVICE_CONFIG="${ZABBIX_CONF_DIR}/zabbix_web_service.conf"

# Default directories
# Internal directory for TLS related files, used when TLS*File specified as plain text values
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

update_config() {
    info "** Preparing Zabbix web service configuration file"

    [[ -f "$ZBX_WEB_SERVICE_CONFIG" ]] || error "Missing configuration file: $ZBX_WEB_SERVICE_CONFIG"

    : "${ZBX_ALLOWEDIP:=zabbix-server}"

    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "LogType" "console"
    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "LogFile"
    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "LogFileSize"
    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "DebugLevel" "${ZBX_DEBUGLEVEL:-}"

    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "AllowedIP" "${ZBX_ALLOWEDIP:-}"
    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "ListenPort" "${ZBX_LISTENPORT:-}"
    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "Timeout" "${ZBX_TIMEOUT:-}"

    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "TLSAccept" "${ZBX_TLSACCEPT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_WEB_SERVICE_CONFIG}" "TLSCAFile" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_WEB_SERVICE_CONFIG}" "TLSCertFile" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_WEB_SERVICE_CONFIG}" "TLSKeyFile" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"

    update_config_var "${ZBX_WEB_SERVICE_CONFIG}" "IgnoreURLCertErrors" "${ZBX_IGNOREURLCERTERRORS:-}"
}

prepare_service() {
    info "** Preparing Zabbix web service"

    update_config
    clear_zbx_env
}

#################################################

if [ $# -eq 0 ]; then
    set -- /usr/sbin/zabbix_web_service
elif [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_web_service "$@"
fi

if [ "${1:-}" = '/usr/sbin/zabbix_web_service' ]; then
    prepare_service
fi

exec "$@"

#################################################
