#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/logging.sh"

: "${JAVA:=/usr/bin/java}"
readonly ZBX_GATEWAY_CONFIG="${ZABBIX_CONF_DIR}/zabbix_java_gateway_logback.xml"
readonly ZABBIX_JAVA_DIR="/usr/sbin/zabbix_java"

build_classpath() {
    local classpath="lib"
    local jar

    while IFS= read -r -d '' jar; do
        classpath="${classpath}:$jar"
    done < <(find lib bin ext_lib -name '*.jar' -print0)

    printf '%s\n' "$classpath"
}

update_config() {
    info "** Preparing Zabbix Java Gateway log configuration file"

    : ${ZBX_DEBUGLEVEL:=info}

    info "Updating ${ZBX_GATEWAY_CONFIG} 'DebugLevel' parameter: '${ZBX_DEBUGLEVEL}'... updated"
    sed -i -e "/^.*<root level=/s/=.*/=\"${ZBX_DEBUGLEVEL}\">/" "$ZBX_GATEWAY_CONFIG"
}

run_service() {
    info "** Preparing Zabbix Java Gateway"
    [[ -f "$ZBX_GATEWAY_CONFIG" ]] || { error "Missing configuration file: $ZBX_GATEWAY_CONFIG" >&2; exit 1; }

    : ${ZBX_TIMEOUT:=3}

    update_config
    cd "$ZABBIX_JAVA_DIR"

    local classpath="$(build_classpath)"

    local -a java_opts=(
        -server
        "-Dlogback.configurationFile=${ZBX_GATEWAY_CONFIG}"
    )

    if [[ -n "${ZBX_JAVA_OPTS:-}" ]]; then
        read -r -a extra_java_opts <<< "${ZBX_JAVA_OPTS}"
        java_opts+=("${extra_java_opts[@]}")
    fi

    local -a zabbix_opts=(
        "-Dsun.rmi.transport.tcp.responseTimeout=${ZBX_TIMEOUT}000"
        "-Dzabbix.listenPort=${ZBX_LISTEN_PORT:-10052}"
        "-Dzabbix.timeout=${ZBX_TIMEOUT}"
        "-Dzabbix.pidFile=/tmp/java_gateway.pid"
    )

    [[ -n "${ZBX_LISTEN_IP:-}" ]] && zabbix_opts+=("-Dzabbix.listenIP=${ZBX_LISTEN_IP}")
    [[ -n "${ZBX_START_POLLERS:-}" ]] && zabbix_opts+=("-Dzabbix.startPollers=${ZBX_START_POLLERS}")
    [[ -n "${ZBX_PROPERTIES_FILE:-}" ]] && zabbix_opts+=("-Dzabbix.propertiesFile=${ZBX_PROPERTIES_FILE}")

    local -a cmd=(
        "$JAVA"
        "${java_opts[@]}"
        -classpath "$classpath"
        "${zabbix_opts[@]}"
        com.zabbix.gateway.JavaGateway
    )

    exec "${cmd[@]}"
}

if [[ $# -eq 0 || "${1:-}" == -* ]]; then
    run_service
fi

exec "$@"
