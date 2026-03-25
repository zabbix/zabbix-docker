#!/usr/bin/env bats
# Tests for db_tls_params() from the Zabbix Docker server/proxy entrypoint scripts.
# The function converts ZBX_DBTLS* environment variables into mariadb CLI flags.

load helpers/entrypoint_functions

teardown() {
    unset ZBX_DBTLSCONNECT ZBX_DBTLSCAFILE ZBX_DBTLSKEYFILE ZBX_DBTLSCERTFILE
}

# ---------------------------------------------------------------------------
# TLS disabled (ZBX_DBTLSCONNECT is unset)
# ---------------------------------------------------------------------------
@test "db_tls_params: returns empty string when ZBX_DBTLSCONNECT is not set" {
    unset ZBX_DBTLSCONNECT
    run db_tls_params
    [ "$status" -eq 0 ]
    [ "$output" = "" ]
}

# ---------------------------------------------------------------------------
# TLS required (no certificate verification)
# ---------------------------------------------------------------------------
@test "db_tls_params: returns --ssl when ZBX_DBTLSCONNECT=required" {
    export ZBX_DBTLSCONNECT="required"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl"* ]]
    [[ "$output" != *"--ssl-verify-server-cert"* ]]
}

# ---------------------------------------------------------------------------
# TLS verify_ca / verify_full (adds server cert verification)
# ---------------------------------------------------------------------------
@test "db_tls_params: adds --ssl-verify-server-cert when ZBX_DBTLSCONNECT=verify_ca" {
    export ZBX_DBTLSCONNECT="verify_ca"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl"* ]]
    [[ "$output" == *"--ssl-verify-server-cert"* ]]
}

@test "db_tls_params: adds --ssl-verify-server-cert when ZBX_DBTLSCONNECT=verify_full" {
    export ZBX_DBTLSCONNECT="verify_full"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl-verify-server-cert"* ]]
}

# ---------------------------------------------------------------------------
# CA file
# ---------------------------------------------------------------------------
@test "db_tls_params: appends --ssl-ca when ZBX_DBTLSCAFILE is set" {
    export ZBX_DBTLSCONNECT="required"
    export ZBX_DBTLSCAFILE="/etc/ssl/ca.crt"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl-ca=/etc/ssl/ca.crt"* ]]
}

@test "db_tls_params: does not add --ssl-ca when ZBX_DBTLSCAFILE is empty" {
    export ZBX_DBTLSCONNECT="required"
    unset ZBX_DBTLSCAFILE
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" != *"--ssl-ca"* ]]
}

# ---------------------------------------------------------------------------
# Key file
# ---------------------------------------------------------------------------
@test "db_tls_params: appends --ssl-key when ZBX_DBTLSKEYFILE is set" {
    export ZBX_DBTLSCONNECT="required"
    export ZBX_DBTLSKEYFILE="/etc/ssl/client.key"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl-key=/etc/ssl/client.key"* ]]
}

# ---------------------------------------------------------------------------
# Cert file
# ---------------------------------------------------------------------------
@test "db_tls_params: appends --ssl-cert when ZBX_DBTLSCERTFILE is set" {
    export ZBX_DBTLSCONNECT="required"
    export ZBX_DBTLSCERTFILE="/etc/ssl/client.crt"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl-cert=/etc/ssl/client.crt"* ]]
}

# ---------------------------------------------------------------------------
# Full mutual-TLS combination
# ---------------------------------------------------------------------------
@test "db_tls_params: includes all flags when all TLS env vars are set" {
    export ZBX_DBTLSCONNECT="verify_full"
    export ZBX_DBTLSCAFILE="/etc/ssl/ca.crt"
    export ZBX_DBTLSKEYFILE="/etc/ssl/client.key"
    export ZBX_DBTLSCERTFILE="/etc/ssl/client.crt"
    run db_tls_params
    [ "$status" -eq 0 ]
    [[ "$output" == *"--ssl"* ]]
    [[ "$output" == *"--ssl-verify-server-cert"* ]]
    [[ "$output" == *"--ssl-ca=/etc/ssl/ca.crt"* ]]
    [[ "$output" == *"--ssl-key=/etc/ssl/client.key"* ]]
    [[ "$output" == *"--ssl-cert=/etc/ssl/client.crt"* ]]
}
