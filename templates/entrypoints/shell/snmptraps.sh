#!/usr/bin/env bash

set -Eeuo pipefail

# shellcheck source=hooks.sh
source /usr/lib/docker-entrypoint/hooks.sh

if [[ $# -eq 0 ]]; then
    printf '%s\n' 'docker-entrypoint: no command was supplied' >&2
    exit 64
fi

if [[ $1 == -* ]]; then
    set -- /usr/sbin/snmptrapd "$@"
fi

if [[ $1 == /usr/sbin/snmptrapd ]]; then
    shift

    # SNMP trap output configuration:
    # S - display the MIB name as well as the object name;
    # T - display a printable version of hexadecimal strings;
    # t - display TimeTicks values as raw numbers;
    # e - remove symbolic labels from enumeration values.
    : "${SNMPTRAP_OUTPUT_OPTIONS:=STte}"

    run_entrypoint_hooks

    config_files=(/etc/snmp/snmptrapd.conf)
    if [[ -n ${SNMP_PERSISTENT_DIR:-} ]]; then
        [[ -f ${SNMP_PERSISTENT_DIR}/snmptrapd.conf ]] && config_files+=("${SNMP_PERSISTENT_DIR}/snmptrapd.conf")
        [[ -f ${SNMP_PERSISTENT_DIR}/snmptrapd_custom.conf ]] && config_files+=("${SNMP_PERSISTENT_DIR}/snmptrapd_custom.conf")
    fi

    config_list="$(IFS=,; printf '%s' "${config_files[*]}")"
    set -- /usr/sbin/snmptrapd --doNotFork=yes -C -c "$config_list" -n -t -X -Lo -A "-O${SNMPTRAP_OUTPUT_OPTIONS}" "$@"
fi

exec "$@"
