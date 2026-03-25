#!/usr/bin/env bats
# Tests for file_env() from the Zabbix Docker server/proxy entrypoint scripts.
# file_env allows a variable to be supplied either directly (VAR) or via a
# Docker secret file (VAR_FILE).

load helpers/entrypoint_functions

setup() {
    # Create a temporary directory for secret files.
    TEST_TMP="$(mktemp -d)"
}

teardown() {
    rm -rf "$TEST_TMP"
    unset MYSQL_PASSWORD MYSQL_PASSWORD_FILE
}

# ---------------------------------------------------------------------------
# Direct environment variable
# ---------------------------------------------------------------------------
@test "file_env: exports value from direct env variable" {
    export MYSQL_PASSWORD="directpass"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
        echo \"\$MYSQL_PASSWORD\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"directpass"* ]]
}

@test "file_env: reports which source was used for direct env variable" {
    export MYSQL_PASSWORD="directpass"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"Using MYSQL_PASSWORD variable from ENV"* ]]
}

# ---------------------------------------------------------------------------
# File-based secret
# ---------------------------------------------------------------------------
@test "file_env: reads value from secret file when VAR_FILE is set" {
    echo -n "secretpass" > "$TEST_TMP/mysql_pass"
    export MYSQL_PASSWORD_FILE="$TEST_TMP/mysql_pass"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
        echo \"\$MYSQL_PASSWORD\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"secretpass"* ]]
}

@test "file_env: reports which source was used for secret file" {
    echo -n "secretpass" > "$TEST_TMP/mysql_pass"
    export MYSQL_PASSWORD_FILE="$TEST_TMP/mysql_pass"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"Using MYSQL_PASSWORD variable from secret file"* ]]
}

@test "file_env: unsets VAR_FILE after reading it" {
    echo -n "secretpass" > "$TEST_TMP/mysql_pass"
    export MYSQL_PASSWORD_FILE="$TEST_TMP/mysql_pass"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
        echo \"FILE_VAR=\${MYSQL_PASSWORD_FILE:-unset}\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"FILE_VAR=unset"* ]]
}

# ---------------------------------------------------------------------------
# Default value
# ---------------------------------------------------------------------------
@test "file_env: uses default value when neither VAR nor VAR_FILE is set" {
    unset MYSQL_PASSWORD MYSQL_PASSWORD_FILE
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD 'defaultpass'
        echo \"\$MYSQL_PASSWORD\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"defaultpass"* ]]
}

@test "file_env: sets empty string when no value and no default given" {
    unset MYSQL_PASSWORD MYSQL_PASSWORD_FILE
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
        echo \"VALUE=[\$MYSQL_PASSWORD]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"VALUE=[]"* ]]
}

# ---------------------------------------------------------------------------
# Error conditions
# ---------------------------------------------------------------------------
@test "file_env: exits with error when both VAR and VAR_FILE are set" {
    export MYSQL_PASSWORD="directpass"
    export MYSQL_PASSWORD_FILE="$TEST_TMP/mysql_pass"
    echo -n "secretpass" > "$TEST_TMP/mysql_pass"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
    "
    [ "$status" -eq 1 ]
    [[ "$output" == *"Both variables"* ]]
    [[ "$output" == *"exclusive"* ]]
}

@test "file_env: exits with error when VAR_FILE points to missing file" {
    export MYSQL_PASSWORD_FILE="/nonexistent/secret/file"
    run bash -c "
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_env MYSQL_PASSWORD
    "
    [ "$status" -eq 1 ]
    [[ "$output" == *"Secret file"* ]]
    [[ "$output" == *"not found"* ]]
}
