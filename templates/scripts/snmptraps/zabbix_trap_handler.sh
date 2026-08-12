#!/usr/bin/env bash

set -eo pipefail

ZABBIX_USER_HOME_DIR="${ZABBIX_USER_HOME_DIR:-/var/lib/zabbix}"
ZABBIX_TRAPS_FILE="${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"

ZBX_SNMP_TRAP_DATE_FORMAT="${ZBX_SNMP_TRAP_DATE_FORMAT:-+%Y-%m-%dT%T%z}"
ZBX_SNMP_TRAP_FORMAT="${ZBX_SNMP_TRAP_FORMAT:-\n}"
ZBX_SNMP_TRAP_USE_DNS="${ZBX_SNMP_TRAP_USE_DNS:-false}"

date_now="$(date "$ZBX_SNMP_TRAP_DATE_FORMAT")"

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

trap_address=""
sender_addr=""
sender_regex='\[([^]]+)\].*->'
vars=""

# The name of the host that sent the notification, as determined by gethostbyaddr(3).
# In fact this line is irrelevant and useless since snmptrapd basically attempts to
# perform reverse name lookup for the transport address (see below).
# In case of failure it will print "<UNKNOWN>"
IFS= read -r host
# The transport address, like "UDP: [172.16.10.12]:23456->[10.150.0.8]:1162"
IFS= read -r sender
# The first OID should always be SNMPv2-MIB::sysUpTime.0
#IFS= read -r uptime
# The second should be SNMPv2-MIB::snmpTrapOID.0
#IFS= read -r trapoid

# The remaining lines will contain the payload varbind list. For SNMPv1 traps, the final OID will be SNMPv2-MIB::snmpTrapEnterprise.0.
while read -r oid val || [ -n "$oid" ]; do
    # Header in Zabbix format shouldn't exist anywhere in vars, it is injection
    # Must exit with 0
    [[ "$val" =~ $zbx_trap_regex ]] && exit 0
    sanitized_val=${val//ZBXTRAP/\'ZBXTRAP\'}

    if [ -z "$vars" ]; then
        vars="$oid = $sanitized_val"
    else
        vars="${vars}${ZBX_SNMP_TRAP_FORMAT}${oid} = $sanitized_val"
    fi

    if [[ "$oid" =~ (^|::)snmpTrapAddress\.0$ ]] || [[ "$oid" =~ ^\.?1\.3\.6\.1\.6\.3\.18\.1\.3\.0$ ]]; then
        trap_address="$sanitized_val"
    fi
done

# Choose the sender identity written immediately after ZBXTRAP.
# The precedence is: snmpTrapAddress.0 (the originating agent carried inside the notification),
# the reverse-resolved transport host when DNS is enabled, the UDP transport
# source observed by snmptrapd, and finally the host line as a structural
# fallback. If a relay or NAT hides the source and does not supply
# snmpTrapAddress.0, the original agent address cannot be recovered here
if [[ "${sender:-}" =~ $sender_regex ]]; then
    sender_addr="${BASH_REMATCH[1]}"
fi

if [ -z "${trap_address:-}" ] && [[ "$ZBX_SNMP_TRAP_USE_DNS" == "true" ]] && [ -n "${host:-}" ] && [ "$host" != "<UNKNOWN>" ] && ! [[ "$host" =~ $sender_regex ]]; then
    sender_addr="$host"
fi

if [ -n "${trap_address:-}" ]; then
    sender_addr="$trap_address"
fi

# Never emit an empty address: fall back to the host line snmptrapd provided (a resolved name or "<UNKNOWN>")
sender_addr="${sender_addr:-$host}"

printf '%b\n' "${date_now} ZBXTRAP ${sender_addr}${ZBX_SNMP_TRAP_FORMAT}${sender}${ZBX_SNMP_TRAP_FORMAT}${vars}" >> "$ZABBIX_TRAPS_FILE"
