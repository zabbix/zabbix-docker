#!/bin/bash

DEFAULT_ARGS="-n -t -X -Lo -A"

CONF_FILE_LIST="/etc/snmp/snmptrapd.conf,$SNMP_PERSISTENT_DIR/snmptrapd.conf"

if [ -f "$SNMP_PERSISTENT_DIR/snmptrapd_custom.conf" ]; then
    CONF_FILE_LIST="$CONF_FILE_LIST,$SNMP_PERSISTENT_DIR/snmptrapd_custom.conf"
fi

/usr/sbin/snmptrapd --doNotFork=yes -C -c "$CONF_FILE_LIST" $DEFAULT_ARGS
