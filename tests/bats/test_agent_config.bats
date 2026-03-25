#!/usr/bin/env bats
# Tests for the prepare_zbx_agent_config() logic extracted from
# Dockerfiles/agent/*/docker-entrypoint.sh.
#
# Rather than sourcing the full entrypoint script (which executes the agent),
# the relevant logic is reproduced here via a thin wrapper so the server-
# connection and passive/active list derivation can be tested in isolation.

load helpers/entrypoint_functions

# ---------------------------------------------------------------------------
# Helper: simulates prepare_zbx_agent_config() with the provided env vars
# and echoes the final PASSIVESERVERS / ACTIVESERVERS values.
# ---------------------------------------------------------------------------
run_prepare() {
    bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'

        # Accept env from the outer environment
        ZBX_SERVER_HOST=\"\${ZBX_SERVER_HOST:-}\"
        ZBX_SERVER_PORT=\"\${ZBX_SERVER_PORT:-10051}\"
        ZBX_PASSIVESERVERS=\"\${ZBX_PASSIVESERVERS:-}\"
        ZBX_ACTIVESERVERS=\"\${ZBX_ACTIVESERVERS:-}\"
        ZBX_PASSIVE_ALLOW=\"\${ZBX_PASSIVE_ALLOW:-true}\"
        ZBX_ACTIVE_ALLOW=\"\${ZBX_ACTIVE_ALLOW:-true}\"

        # Replicate the prepare_zbx_agent_config() logic
        if [ -n \"\$ZBX_SERVER_HOST\" ] && [ -n \"\$ZBX_PASSIVESERVERS\" ]; then
            ZBX_PASSIVESERVERS=\"\$ZBX_SERVER_HOST,\$ZBX_PASSIVESERVERS\"
        elif [ -n \"\$ZBX_SERVER_HOST\" ]; then
            ZBX_PASSIVESERVERS=\"\$ZBX_SERVER_HOST\"
        fi

        if [ -n \"\$ZBX_SERVER_HOST\" ]; then
            if [ -n \"\$ZBX_SERVER_PORT\" ] && [ \"\$ZBX_SERVER_PORT\" != '10051' ]; then
                ZBX_SERVER_HOST=\"\$ZBX_SERVER_HOST:\$ZBX_SERVER_PORT\"
            fi
            if [ -n \"\$ZBX_ACTIVESERVERS\" ]; then
                ZBX_ACTIVESERVERS=\"\$ZBX_SERVER_HOST,\$ZBX_ACTIVESERVERS\"
            else
                ZBX_ACTIVESERVERS=\"\$ZBX_SERVER_HOST\"
            fi
        fi

        if [ \"\${ZBX_PASSIVE_ALLOW,,}\" == 'true' ] && [ -n \"\$ZBX_PASSIVESERVERS\" ]; then
            export ZBX_PASSIVESERVERS
        else
            unset ZBX_PASSIVESERVERS
        fi

        if [ \"\${ZBX_ACTIVE_ALLOW,,}\" == 'true' ] && [ -n \"\$ZBX_ACTIVESERVERS\" ]; then
            export ZBX_ACTIVESERVERS
        else
            unset ZBX_ACTIVESERVERS
        fi

        echo \"PASSIVE=[\${ZBX_PASSIVESERVERS:-unset}]\"
        echo \"ACTIVE=[\${ZBX_ACTIVESERVERS:-unset}]\"
    "
}

teardown() {
    unset ZBX_SERVER_HOST ZBX_SERVER_PORT ZBX_PASSIVESERVERS ZBX_ACTIVESERVERS
    unset ZBX_PASSIVE_ALLOW ZBX_ACTIVE_ALLOW
}

# ---------------------------------------------------------------------------
# Default server host is used for both passive and active lists
# ---------------------------------------------------------------------------
@test "agent_config: ZBX_SERVER_HOST populates both passive and active server lists" {
    export ZBX_SERVER_HOST="zabbix-server"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASSIVE=[zabbix-server]"* ]]
    [[ "$output" == *"ACTIVE=[zabbix-server]"* ]]
}

# ---------------------------------------------------------------------------
# Non-default port is appended for active checks only
# ---------------------------------------------------------------------------
@test "agent_config: custom port is appended to active server host" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_SERVER_PORT="10061"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"ACTIVE=[zabbix-server:10061]"* ]]
}

@test "agent_config: default port (10051) is NOT appended to active server host" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_SERVER_PORT="10051"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"ACTIVE=[zabbix-server]"* ]]
    [[ "$output" != *"ACTIVE=[zabbix-server:10051]"* ]]
}

# ---------------------------------------------------------------------------
# Additional passive/active server merging
# ---------------------------------------------------------------------------
@test "agent_config: additional passive servers are appended after ZBX_SERVER_HOST" {
    export ZBX_SERVER_HOST="primary-server"
    export ZBX_PASSIVESERVERS="secondary-server"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASSIVE=[primary-server,secondary-server]"* ]]
}

@test "agent_config: additional active servers are appended after ZBX_SERVER_HOST" {
    export ZBX_SERVER_HOST="primary-server"
    export ZBX_ACTIVESERVERS="secondary-server"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"ACTIVE=[primary-server,secondary-server]"* ]]
}

# ---------------------------------------------------------------------------
# Disabling passive / active checks
# ---------------------------------------------------------------------------
@test "agent_config: passive checks are disabled when ZBX_PASSIVE_ALLOW=false" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_PASSIVE_ALLOW="false"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASSIVE=[unset]"* ]]
}

@test "agent_config: active checks are disabled when ZBX_ACTIVE_ALLOW=false" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_ACTIVE_ALLOW="false"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"ACTIVE=[unset]"* ]]
}

@test "agent_config: both checks are disabled when both ALLOW flags are false" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_PASSIVE_ALLOW="false"
    export ZBX_ACTIVE_ALLOW="false"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASSIVE=[unset]"* ]]
    [[ "$output" == *"ACTIVE=[unset]"* ]]
}

# ---------------------------------------------------------------------------
# Case-insensitive ALLOW flags
# ---------------------------------------------------------------------------
@test "agent_config: ZBX_PASSIVE_ALLOW=False (mixed case) is treated as false" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_PASSIVE_ALLOW="False"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASSIVE=[unset]"* ]]
}

@test "agent_config: ZBX_ACTIVE_ALLOW=TRUE (uppercase) is treated as true" {
    export ZBX_SERVER_HOST="zabbix-server"
    export ZBX_ACTIVE_ALLOW="TRUE"
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"ACTIVE=[zabbix-server]"* ]]
}

# ---------------------------------------------------------------------------
# No server host set
# ---------------------------------------------------------------------------
@test "agent_config: passive and active lists are unset when ZBX_SERVER_HOST is empty" {
    unset ZBX_SERVER_HOST
    run run_prepare
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASSIVE=[unset]"* ]]
    [[ "$output" == *"ACTIVE=[unset]"* ]]
}
