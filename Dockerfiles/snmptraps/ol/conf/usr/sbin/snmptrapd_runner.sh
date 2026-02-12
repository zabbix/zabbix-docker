#!/bin/bash

# SNMP Trap output configuration
# S - Display the name of the MIB, as well as the object name (This is the default OID output format)
# T - If values are printed as Hex strings, display a printable version as well
# t - Display TimeTicks values as raw numbers
# e - Removes the symbolic labels from enumeration values
#
: ${SNMPTRAP_OUTPUT_OPTIONS:="STte"}

DEFAULT_ARGS="-n -t -X -Lo -A -O${SNMPTRAP_OUTPUT_OPTIONS}"

CONF_FILE_LIST="/etc/snmp/snmptrapd.conf"

if [ -f "$SNMP_PERSISTENT_DIR/snmptrapd.conf" ]; then
    CONF_FILE_LIST="$CONF_FILE_LIST,$SNMP_PERSISTENT_DIR/snmptrapd.conf"
fi

if [ -f "$SNMP_PERSISTENT_DIR/snmptrapd_custom.conf" ]; then
    CONF_FILE_LIST="$CONF_FILE_LIST,$SNMP_PERSISTENT_DIR/snmptrapd_custom.conf"
fi

/usr/sbin/snmptrapd --doNotFork=yes -C -c "$CONF_FILE_LIST" $DEFAULT_ARGS
