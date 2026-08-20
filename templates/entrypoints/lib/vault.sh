# shellcheck shell=bash

vault_requested() {
    [ -n "${ZBX_VAULTURL:-}" ] || \
        [ -n "${ZBX_VAULTDBPATH:-}" ] || \
        [ -n "${VAULT_TOKEN:-}" ]
}

get_vault_secrets() {
    local wait_timeout=5
    local vaultdata errors
    local vault_base_url vault_db_path vault_mount vault_secret_path vault_url
    local curl_opts=(-sS -m 10)

    if [ -z "${ZBX_VAULTURL:-}" ] || [ -z "${ZBX_VAULTDBPATH:-}" ] || [ -z "${VAULT_TOKEN:-}" ]; then
        error "ZBX_VAULTURL, ZBX_VAULTDBPATH and VAULT_TOKEN must be set to use HashiCorp Vault"
    fi

    vault_base_url="${ZBX_VAULTURL%/}"
    vault_db_path="${ZBX_VAULTDBPATH#/}"
    vault_db_path="${vault_db_path%/}"

    if [[ "${vault_db_path}" != */* ]] || [ -z "${vault_db_path%%/*}" ] || [ -z "${vault_db_path#*/}" ]; then
        error "ZBX_VAULTDBPATH must contain a mount point and secret path"
    fi

    vault_mount="${vault_db_path%%/*}"
    vault_secret_path="${vault_db_path#*/}"
    vault_url="${vault_base_url}/v1/${vault_mount}/data/${vault_secret_path}"

    info "***** VAULT URL: ${vault_url}"
    while ! vaultdata="$(curl "${curl_opts[@]}" -H "X-Vault-Token: ${VAULT_TOKEN}" "${vault_url}")"; do
        info "**** Vault is not available. Waiting ${wait_timeout} seconds... ****"
        sleep "${wait_timeout}"
    done

    errors="$(printf '%s' "${vaultdata}" | jq -r '.errors // empty')"
    if [ -n "${errors}" ]; then
        error "Error getting secrets from vault: ${errors}"
    fi

    # Variables are consumed by the database library sourcing this file.
    # shellcheck disable=SC2034
    DB_SERVER_ZBX_USER="$(printf '%s' "${vaultdata}" | jq -r '.data.data.username')"
    # shellcheck disable=SC2034
    DB_SERVER_ZBX_PASS="$(printf '%s' "${vaultdata}" | jq -r '.data.data.password')"
}
