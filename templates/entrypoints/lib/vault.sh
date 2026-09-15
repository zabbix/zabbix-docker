# shellcheck shell=bash

get_vault_secrets() {
    local wait_timeout=5
    local vaultdata errors
    local cyberark_opts
    local vault_base_url vault_db_path vault_prefix vault_url

    if [ -z "${ZBX_VAULTURL:-}" ] || [ -z "${ZBX_VAULTDBPATH:-}" ]; then
        error "Missing variables! If ZBX_VAULT is used then ZBX_VAULTURL and ZBX_VAULTDBPATH must be set"
    fi

    vault_base_url="${ZBX_VAULTURL%/}"
    vault_db_path="${ZBX_VAULTDBPATH#/}"
    vault_db_path="${vault_db_path%/}"
    vault_prefix="${ZBX_VAULTPREFIX:-}"

    if [ "${ZBX_VAULT:-}" = "HashiCorp" ]; then
        if [ -z "$vault_prefix" ]; then
            vault_prefix="v1/${vault_db_path%/*}/data"
            vault_db_path="${vault_db_path##*/}"
        else
            vault_prefix="${vault_prefix#/}"
            vault_prefix="${vault_prefix%/}"
        fi

        vault_url="${vault_base_url}/${vault_prefix}/${vault_db_path}"
        local curl_opts=(-s -m 10 -k)

        info "***** VAULT URL: $vault_url"

        while ! vaultdata="$(curl "${curl_opts[@]}" -H "X-Vault-Token: $VAULT_TOKEN" "$vault_url")"; do
            info "**** Vault is not available. Waiting ${wait_timeout} seconds... ****"
            sleep "$wait_timeout"
        done
        errors="$(printf '%s' "$vaultdata" | jq -r '.errors // empty')"
        if [ -n "${errors}" ]; then
            error "Error getting secrets from vault: $errors"
        fi
        DB_SERVER_ZBX_USER="$(printf '%s' "$vaultdata" | jq -r '.data.data.username')"
        DB_SERVER_ZBX_PASS="$(printf '%s' "$vaultdata" | jq -r '.data.data.password')"
    elif [ "${ZBX_VAULT:-}" = "CyberArk" ]; then
        if [ -z "${ZBX_VAULTCERTFILE:-}" ] ; then
            error "Missing variables! If CyberArk is used then ZBX_VAULTCERTFILE must be set"
        fi

        cyberark_opts=(-H "Content-type: application/json" --cert "$ZBX_VAULTCERTFILE")
        if [ -z "$vault_prefix" ]; then
            vault_prefix="/AIMWebService/api/Accounts?"
        else
            vault_prefix="/${vault_prefix#/}"
        fi
        vault_url="${vault_base_url}${vault_prefix}${vault_db_path}"
        local curl_opts=(-s -m 10)

        info "***** VAULT URL: $vault_url"

        if [ -n "${ZBX_VAULTKEYFILE:-}" ]; then
            cyberark_opts+=(--key "$ZBX_VAULTKEYFILE")
        fi
        while ! vaultdata="$(curl "${curl_opts[@]}" "${cyberark_opts[@]}" "$vault_url")"; do
            info "**** Vault is not available. Waiting ${wait_timeout} seconds... ****"
            sleep "$wait_timeout"
        done

        errors=$(printf '%s' "$vaultdata" | jq -r '.ErrorCode // empty')
        if [ -n "${errors}" ]; then
            error "Error getting secrets from vault: $errors"
        fi

        DB_SERVER_ZBX_USER="$(printf '%s' "$vaultdata" | jq -r '.UserName')"
        DB_SERVER_ZBX_PASS="$(printf '%s' "$vaultdata" | jq -r '.Content')"

    else
        error "ZBX_VAULT has wrong value. HashiCorp or CyberArk are supported!"
    fi
}
