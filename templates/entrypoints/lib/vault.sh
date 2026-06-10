# shellcheck shell=bash

get_vault_secrets() {
    local wait_timeout=5
    local vaultdata errors
    local cyberark_opts
    if [ -z "${ZBX_VAULTURL:-}" ] || [ -z "${ZBX_VAULTDBPATH:-}" ]; then
        error "Missing variables! If ZBX_VAULT is used then ZBX_VAULTURL and ZBX_VAULTDBPATH must be set"
    fi

    # Sanitize input
    ZBX_VAULTURL="${ZBX_VAULTURL%/}"
    ZBX_VAULTDBPATH="${ZBX_VAULTDBPATH#/}"
    ZBX_VAULTDBPATH="${ZBX_VAULTDBPATH%/}"

    if [ "${ZBX_VAULT:-}" = "HashiCorp" ]; then
        if [ -z "${ZBX_VAULTPREFIX:-}" ]; then
            ZBX_VAULTPREFIX="v1/${ZBX_VAULTDBPATH%/*}/data"
            ZBX_VAULTDBPATH="${ZBX_VAULTDBPATH##*/}"
        else
            ZBX_VAULTPREFIX="${ZBX_VAULTPREFIX#/}"
            ZBX_VAULTPREFIX="${ZBX_VAULTPREFIX%/}"
        fi

        local vault_url="${ZBX_VAULTURL}/${ZBX_VAULTPREFIX}/${ZBX_VAULTDBPATH}"
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
        if [ -z "${ZBX_VAULTPREFIX:-}" ]; then
            ZBX_VAULTPREFIX="AIMWebService/api/Accounts?"
        else
            ZBX_VAULTPREFIX="${ZBX_VAULTPREFIX#/}"
            ZBX_VAULTPREFIX="${ZBX_VAULTPREFIX%/}/"
        fi
        local vault_url="${ZBX_VAULTURL}${ZBX_VAULTPREFIX}${ZBX_VAULTDBPATH}"
        local curl_opts=(-s -m 10)
        info "***** VAULT URL: $vault_url"
        # if key is defined use it
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
