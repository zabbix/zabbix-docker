#!/usr/bin/env bats
# Tests for update_config_var() and update_config_multiple_var() from the
# Zabbix Docker entrypoint scripts.

load helpers/entrypoint_functions

# Temporary config file used by all tests in this file.
CONF_FILE=""

setup() {
    CONF_FILE="$(mktemp)"
    export ZABBIX_USER_HOME_DIR="/var/lib/zabbix"
}

teardown() {
    rm -f "$CONF_FILE"
}

# ---------------------------------------------------------------------------
# Missing config file
# ---------------------------------------------------------------------------
@test "update_config_var: warns when config file does not exist" {
    run update_config_var "/nonexistent/path/zabbix.conf" "LogFile" "/tmp/zabbix.log"
    [ "$status" -eq 0 ]
    [[ "$output" == *"does not exist"* ]]
}

# ---------------------------------------------------------------------------
# Removing a parameter (empty value)
# ---------------------------------------------------------------------------
@test "update_config_var: removes existing parameter when value is empty" {
    echo "LogFile=/var/log/zabbix.log" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" ""
    [ "$status" -eq 0 ]
    [[ "$output" == *"removed"* ]]
    # Key must no longer appear in the file
    run grep -c "^LogFile=" "$CONF_FILE"
    [ "$output" -eq 0 ]
}

@test "update_config_var: silently skips removal when parameter is not in file" {
    echo "# some config" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" ""
    [ "$status" -eq 0 ]
    [[ "$output" == *"removed"* ]]
}

# ---------------------------------------------------------------------------
# Undefined value ("")
# ---------------------------------------------------------------------------
@test "update_config_var: clears existing parameter when value is double-quoted empty" {
    echo "LogFile=/var/log/zabbix.log" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" '""'
    [ "$status" -eq 0 ]
    [[ "$output" == *"undefined"* ]]
    run grep "^LogFile=$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

@test "update_config_var: adds parameter with empty value when comment exists" {
    echo "# LogFile=/var/log/zabbix.log" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" '""'
    [ "$status" -eq 0 ]
    [[ "$output" == *"undefined"* ]]
    run grep "^LogFile=$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# Adding a new parameter
# ---------------------------------------------------------------------------
@test "update_config_var: adds parameter after comment line" {
    echo "# LogFile=/var/log/zabbix.log" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" "/tmp/test.log"
    [ "$status" -eq 0 ]
    [[ "$output" == *"added"* ]]
    run grep "^LogFile=/tmp/test\.log$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

@test "update_config_var: appends parameter at end when no comment exists" {
    echo "OtherParam=value" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" "/tmp/test.log"
    [ "$status" -eq 0 ]
    [[ "$output" == *"added at the end"* ]]
    run grep "^LogFile=/tmp/test\.log$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# Updating an existing parameter
# ---------------------------------------------------------------------------
@test "update_config_var: updates an existing parameter value" {
    echo "LogFile=/old/path.log" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" "/new/path.log"
    [ "$status" -eq 0 ]
    [[ "$output" == *"updated"* ]]
    run grep "^LogFile=\/new\/path\.log$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# Existing value already matches
# ---------------------------------------------------------------------------
@test "update_config_var: reports 'exists' when value is already set" {
    echo "LogFile=/tmp/test.log" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LogFile" "/tmp/test.log"
    [ "$status" -eq 0 ]
    [[ "$output" == *"exists"* ]]
}

# ---------------------------------------------------------------------------
# Multiple-values mode (is_multiple=true)
# ---------------------------------------------------------------------------
@test "update_config_var: allows duplicate key when is_multiple=true" {
    printf "LoadModule=mod_a.so\n" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "LoadModule" "mod_b.so" "true"
    [ "$status" -eq 0 ]
    # Both values must be in the file
    run grep -c "^LoadModule=" "$CONF_FILE"
    [ "$output" -eq 2 ]
}

# ---------------------------------------------------------------------------
# Masked parameter (TLSPSKIdentity)
# ---------------------------------------------------------------------------
@test "update_config_var: masks TLSPSKIdentity value in output" {
    echo "# TLSPSKIdentity=myid" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "TLSPSKIdentity" "supersecret"
    [ "$status" -eq 0 ]
    [[ "$output" == *"****"* ]]
    [[ "$output" != *"supersecret"* ]]
}

# ---------------------------------------------------------------------------
# TLS file path expansion
# ---------------------------------------------------------------------------
@test "update_config_var: prepends home dir for relative TLSCAFile path" {
    echo "# TLSCAFile=ca.crt" > "$CONF_FILE"
    ZABBIX_USER_HOME_DIR="/home/zabbix"
    run update_config_var "$CONF_FILE" "TLSCAFile" "ca.crt"
    [ "$status" -eq 0 ]
    run grep "^TLSCAFile=/home/zabbix/enc/ca\.crt$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

@test "update_config_var: keeps absolute TLSCAFile path unchanged" {
    echo "# TLSCAFile=/etc/ssl/ca.crt" > "$CONF_FILE"
    run update_config_var "$CONF_FILE" "TLSCAFile" "/etc/ssl/ca.crt"
    [ "$status" -eq 0 ]
    run grep "^TLSCAFile=\/etc\/ssl\/ca\.crt$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# update_config_multiple_var
# ---------------------------------------------------------------------------
@test "update_config_multiple_var: adds each comma-separated value" {
    printf "# DenyKey=\n" > "$CONF_FILE"
    run update_config_multiple_var "$CONF_FILE" "DenyKey" "system.run[*],vfs.file.contents[*]"
    [ "$status" -eq 0 ]
    run grep -c "^DenyKey=" "$CONF_FILE"
    [ "$output" -eq 2 ]
}

@test "update_config_multiple_var: strips surrounding double-quotes from value list" {
    printf "# AllowKey=\n# AllowKey=\n" > "$CONF_FILE"
    run update_config_multiple_var "$CONF_FILE" "AllowKey" '"net.if.*,system.cpu.load"'
    [ "$status" -eq 0 ]
    run grep "^AllowKey=net\.if\.\*$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}

@test "update_config_multiple_var: single value without commas is handled" {
    printf "# LoadModule=\n" > "$CONF_FILE"
    run update_config_multiple_var "$CONF_FILE" "LoadModule" "dummy.so"
    [ "$status" -eq 0 ]
    run grep "^LoadModule=dummy\.so$" "$CONF_FILE"
    [ "$status" -eq 0 ]
}
