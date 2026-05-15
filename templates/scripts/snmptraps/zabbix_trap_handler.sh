#!/usr/bin/env bash

set -eo pipefail

ZABBIX_TRAPS_FILE="${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"

ZBX_SNMP_TRAP_DATE_FORMAT="${ZBX_SNMP_TRAP_DATE_FORMAT:-+%Y-%m-%dT%T%z}"
ZBX_SNMP_TRAP_FORMAT="${ZBX_SNMP_TRAP_FORMAT:-\n}"
ZBX_SNMP_TRAP_USE_DNS="${ZBX_SNMP_TRAP_USE_DNS:-false}"

date_now="$(date "$ZBX_SNMP_TRAP_DATE_FORMAT")"

trap_address=""
sender_addr=""
vars=""

# The name of the host that sent the notification, as determined by gethostbyaddr(3).
# In fact this line is irrelevant and useless since snmptrapd basically attempts to
# perform reverse name lookup for the transport address (see below).
# In case of failure it will print "<UNKNOWN>"
IFS= read -r host
# The transport address, like "[UDP: [172.16.10.12]:23456->[10.150.0.8]]"
IFS= read -r sender
# The first OID should always be SNMPv2-MIB::sysUpTime.0
#IFS= read -r uptime
# The second should be SNMPv2-MIB::snmpTrapOID.0
#IFS= read -r trapoid

# The remaining lines will contain the payload varbind list. For SNMPv1 traps, the final OID will be SNMPv2-MIB::snmpTrapEnterprise.0.
while read -r oid val; do
    if [ -z "$vars" ]; then
        vars="$oid = $val"
    else
        vars="${vars}${ZBX_SNMP_TRAP_FORMAT}${oid} = $val"
    fi

    if [[ "$oid" =~ snmpTrapAddress\.0 ]] || [[ "$oid" =~ 1\.3\.6\.1\.6\.3\.18\.1\.3\.0 ]]; then
        trap_address="$val"
    fi
done

if [[ "${sender:-}" =~ \[(.*?)\].*-\> ]]; then
    sender_addr="${BASH_REMATCH[1]}"
fi

if [ -n "${trap_address:-}" ]; then
    sender_addr="$trap_address"
fi

if [[ "$ZBX_SNMP_TRAP_USE_DNS" == "true" ]] && ! [[ "${host:-}" =~ \[(.*?)\].*-\> ]]; then
    sender_addr="$host"
fi

# Header in Zabbix format shouldn't exist anywhere in vars, it is injection
# Must exit with 0
date_regex="$(
    printf '%s' "$ZBX_SNMP_TRAP_DATE_FORMAT" | sed \
        -e 's/^+//g' \
        -e 's/%Y/[0-9]\{4\}/g' \
        -e 's/%m/[0-9]\{2\}/g' \
        -e 's/%d/[0-9]\{2\}/g' \
        -e 's/%T/[0-9]\{2\}:[0-9]\{2\}:[0-9]\{2\}/g' \
        -e 's/%z/[\+\-][0-9]\{4\}/g' \
        -e 's/%H/[0-9]\{2\}/g' \
        -e 's/%M/[0-9]\{2\}/g' \
        -e 's/%S/[0-9]\{2\}/g'
)"

zbx_trap_regex="${date_regex} ZBXTRAP"
printf '%s\n' "$vars" | grep -qE "$zbx_trap_regex" && exit 0

printf '%b\n' "${date_now} ZBXTRAP ${sender_addr}${ZBX_SNMP_TRAP_FORMAT}${sender}${ZBX_SNMP_TRAP_FORMAT}${vars}" >> "$ZABBIX_TRAPS_FILE"
