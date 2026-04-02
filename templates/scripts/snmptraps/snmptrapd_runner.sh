#!/usr/bin/env bash

set -euo pipefail

# SNMP Trap output configuration
# S - Display the name of the MIB, as well as the object name (default OID output format)
# T - If values are printed as Hex strings, display a printable version as well
# t - Display TimeTicks values as raw numbers
# e - Removes the symbolic labels from enumeration values
#
: "${SNMPTRAP_OUTPUT_OPTIONS:=STte}"

conf_file_list="/etc/snmp/snmptrapd.conf"

if [ -f "${SNMP_PERSISTENT_DIR:-}/snmptrapd.conf" ]; then
    conf_file_list="${conf_file_list},${SNMP_PERSISTENT_DIR}/snmptrapd.conf"
fi

if [ -f "${SNMP_PERSISTENT_DIR:-}/snmptrapd_custom.conf" ]; then
    conf_file_list="${conf_file_list},${SNMP_PERSISTENT_DIR}/snmptrapd_custom.conf"
fi

args=( --doNotFork=yes -C -c "$conf_file_list" -n -t -X -Lo -A "-O${SNMPTRAP_OUTPUT_OPTIONS}" )

exec /usr/sbin/snmptrapd "${args[@]}"
