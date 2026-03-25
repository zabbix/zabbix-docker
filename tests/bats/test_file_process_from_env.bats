#!/usr/bin/env bats
# Tests for file_process_from_env() from the Zabbix Docker entrypoint scripts.
# This function writes a plain-text TLS value from an env var to an internal
# file so that the Zabbix daemon can reference a path instead of a raw secret.

load helpers/entrypoint_functions

setup() {
    TEST_TMP="$(mktemp -d)"
    export ZABBIX_INTERNAL_ENC_DIR="$TEST_TMP"
    unset ZBX_TLSCA ZBX_TLSCAFILE ZBX_TLSPSK ZBX_TLSPSKFILE
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ---------------------------------------------------------------------------
# When plain-text value is provided
# ---------------------------------------------------------------------------
@test "file_process_from_env: writes plain-text value to internal enc dir" {
    run bash -c "
        export ZABBIX_INTERNAL_ENC_DIR='$TEST_TMP'
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_process_from_env 'ZBX_TLSCAFILE' '' 'plain-cert-content'
        cat '$TEST_TMP/ZBX_TLSCAFILE'
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"plain-cert-content"* ]]
}

@test "file_process_from_env: exports the var pointing to the internal file path" {
    run bash -c "
        export ZABBIX_INTERNAL_ENC_DIR='$TEST_TMP'
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_process_from_env 'ZBX_TLSCAFILE' '/original/path/ca.crt' 'plain-cert-content'
        echo \"\$ZBX_TLSCAFILE\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"$TEST_TMP/ZBX_TLSCAFILE"* ]]
}

@test "file_process_from_env: unsets the plain-text companion variable" {
    run bash -c "
        export ZABBIX_INTERNAL_ENC_DIR='$TEST_TMP'
        export ZBX_TLSCA='plain-cert-content'
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_process_from_env 'ZBX_TLSCAFILE' '' 'plain-cert-content'
        echo \"TLSCA=[\${ZBX_TLSCA:-unset}]\"
    "
    [ "$status" -eq 0 ]
    [[ "$output" == *"TLSCA=[unset]"* ]]
}

# ---------------------------------------------------------------------------
# When no plain-text value is provided (value already in a file)
# ---------------------------------------------------------------------------
@test "file_process_from_env: does not create internal file when value is empty" {
    run bash -c "
        export ZABBIX_INTERNAL_ENC_DIR='$TEST_TMP'
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        file_process_from_env 'ZBX_TLSCAFILE' '/etc/ssl/ca.crt' ''
        ls '$TEST_TMP/' | wc -l
    "
    [ "$status" -eq 0 ]
    [ "$output" -eq 0 ]
}

@test "file_process_from_env: does not override the file_name var when value is empty" {
    run bash -c "
        export ZABBIX_INTERNAL_ENC_DIR='$TEST_TMP'
        source '${BATS_TEST_DIRNAME}/helpers/entrypoint_functions.bash'
        export ZBX_TLSCAFILE='/etc/ssl/ca.crt'
        file_process_from_env 'ZBX_TLSCAFILE' '/etc/ssl/ca.crt' ''
        echo \"\$ZBX_TLSCAFILE\"
    "
    [ "$status" -eq 0 ]
    # The variable should still hold the original path (not overwritten)
    [[ "$output" == *"/etc/ssl/ca.crt"* ]]
}
