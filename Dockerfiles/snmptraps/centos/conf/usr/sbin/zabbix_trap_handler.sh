#!/bin/bash

ZABBIX_TRAPS_FILE="/var/lib/zabbix/snmptraps/snmptraps.log"

ZBX_SNMP_TRAP_DATE_FORMAT=${ZBX_SNMP_TRAP_DATE_FORMAT:-"+%Y%m%d.%H%M%S"}

ZBX_SNMP_TRAP_FORMAT=${ZBX_SNMP_TRAP_FORMAT:-"\n"}

ZBX_SNMP_TRAP_USE_DNS=${ZBX_SNMP_TRAP_USE_DNS:-"false"}

date=$(date "$ZBX_SNMP_TRAP_DATE_FORMAT")

# The name of the host that sent the notification, as determined by gethostbyaddr(3).
# In fact this line is irrelevant and useless since snmptrapd basically attempts to
# perform reverse name lookup for the transport address (see below).
# In case of failure it will print "<UNKNOWN>"
read host
# The transport address, like "[UDP: [172.16.10.12]:23456->[10.150.0.8]]"
read sender
# The first OID should always be SNMPv2-MIB::sysUpTime.0
#read uptime
# the second should be SNMPv2-MIB::snmpTrapOID.0
#read trapoid

# The remaining lines will contain the payload varbind list. For SNMPv1 traps, the final OID will be SNMPv2-MIB::snmpTrapEnterprise.0.
vars=
while read oid val
do
    if [ "$vars" = "" ]
    then
        vars="$oid = $val"
    else
        vars="$vars$ZBX_SNMP_TRAP_FORMAT$oid = $val"
    fi

    if [[ "$oid" =~ snmpTrapAddress\.0 ]] || [[ "$oid" =~ 1\.3\.6\.1\.6\.3\.18\.1\.3\.0 ]]; then
        trap_address=$val
    fi
done

[[ ${sender} =~ \[(.*?)\].*\-\> ]] && sender_addr=${BASH_REMATCH[1]}

! [ -z $trap_address ] && sender_addr=$trap_address

[[ "$ZBX_SNMP_TRAP_USE_DNS" == "true" ]] && ! [[ ${host} =~ \[(.*?)\].*\-\> ]] && sender_addr=$host

# Header in Zabbix format shouldn't exist anywhere in vars, it is injection
# Must exit with 0
date_regex=$(echo "$ZBX_SNMP_TRAP_DATE_FORMAT" | sed -e 's/^+//g' \
        -e 's/%Y/[0-9]\{4\}/g' \
        -e 's/%m/[0-9]\{2\}/g' \
        -e 's/%d/[0-9]\{2\}/g' \
        -e 's/%T/[0-9]\{2\}:[0-9]\{2\}:[0-9]\{2\}/g' \
        -e 's/%z/[\+\-][0-9]\{4\}/g' \
        -e 's/%H/[0-9]\{2\}/g' \
        -e 's/%M/[0-9]\{2\}/g' \
        -e 's/%S/[0-9]\{2\}/g')

zbx_trap_regex="$date_regex ZBXTRAP"
echo "$vars" | grep -qE "$zbx_trap_regex" && exit 0

echo -e "$date ZBXTRAP $sender_addr$ZBX_SNMP_TRAP_FORMAT$sender$ZBX_SNMP_TRAP_FORMAT$vars" >> $ZABBIX_TRAPS_FILE
