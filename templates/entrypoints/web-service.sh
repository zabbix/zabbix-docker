#!/bin/bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

# Default directories
# Internal directory for TLS related files, used when TLS*File specified as plain text values
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

update_config() {
    export ZBX_ALLOWEDIP=${ZBX_ALLOWEDIP:="zabbix-server"}

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCAFILE" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCERTFILE" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSKEYFILE" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"
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
