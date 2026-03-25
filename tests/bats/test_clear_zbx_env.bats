#!/usr/bin/env bats
# Tests for clear_zbx_env() from the Zabbix Docker agent entrypoint script.
# The function removes all ZABBIX_* environment variables after initialization.

load helpers/entrypoint_functions

teardown() {
    unset ZBX_CLEAR_ENV
    unset ZABBIX_CONF_DIR ZABBIX_USER_HOME_DIR ZABBIX_CUSTOM_VAR
}

# ---------------------------------------------------------------------------
# Default behaviour (clear is enabled unless explicitly set to "false")
# ---------------------------------------------------------------------------
@test "clear_zbx_env: removes ZABBIX_* variables by default" {
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        export ZABBIX_CONF_DIR='/etc/zabbix'
        export ZABBIX_USER_HOME_DIR='/var/lib/zabbix'
        clear_zbx_env
        echo \"CONF_DIR=[\${ZABBIX_CONF_DIR:-unset}]\"
        echo \"HOME_DIR=[\${ZABBIX_USER_HOME_DIR:-unset}]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"CONF_DIR=[unset]"* ]]
    [[ "$output" == *"HOME_DIR=[unset]"* ]]
}

@test "clear_zbx_env: removes all ZABBIX_* prefixed variables" {
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        export ZABBIX_CUSTOM_VAR='should_be_cleared'
        export ZABBIX_ANOTHER_VAR='also_cleared'
        clear_zbx_env
        echo \"VAR1=[\${ZABBIX_CUSTOM_VAR:-unset}]\"
        echo \"VAR2=[\${ZABBIX_ANOTHER_VAR:-unset}]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"VAR1=[unset]"* ]]
    [[ "$output" == *"VAR2=[unset]"* ]]
}

@test "clear_zbx_env: does not remove non-ZABBIX variables" {
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        export MY_CUSTOM_VAR='preserved'
        export ZBX_SERVER='zabbix-server'
        clear_zbx_env
        echo \"CUSTOM=[\${MY_CUSTOM_VAR:-unset}]\"
        echo \"ZBX=[\${ZBX_SERVER:-unset}]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"CUSTOM=[preserved]"* ]]
    [[ "$output" == *"ZBX=[zabbix-server]"* ]]
}

# ---------------------------------------------------------------------------
# ZBX_CLEAR_ENV=false disables clearing
# ---------------------------------------------------------------------------
@test "clear_zbx_env: skips clearing when ZBX_CLEAR_ENV is false" {
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        export ZBX_CLEAR_ENV='false'
        export ZABBIX_CONF_DIR='/etc/zabbix'
        clear_zbx_env
        echo \"CONF_DIR=[\${ZABBIX_CONF_DIR:-unset}]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"CONF_DIR=[/etc/zabbix]"* ]]
}

@test "clear_zbx_env: clears when ZBX_CLEAR_ENV is set to anything other than false" {
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        export ZBX_CLEAR_ENV='true'
        export ZABBIX_CONF_DIR='/etc/zabbix'
        clear_zbx_env
        echo \"CONF_DIR=[\${ZABBIX_CONF_DIR:-unset}]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"CONF_DIR=[unset]"* ]]
}
