#!/usr/bin/env bash
# Shared helper: defines the common entrypoint functions used across all
# Zabbix Docker entrypoint scripts so they can be tested in isolation.

# ---------------------------------------------------------------------------
# escape_spec_char  – escapes special sed/regex characters inside a value
# ---------------------------------------------------------------------------
escape_spec_char() {
    local var_value=$1

    var_value="${var_value//\\/\\\\}"
    var_value="${var_value//[$'\n']/}"
    var_value="${var_value//\//\\/}"
    var_value="${var_value//./\\.}"
    var_value="${var_value//\*/\\*}"
    var_value="${var_value//^/\\^}"
    var_value="${var_value//\$/\\\$}"
    var_value="${var_value//\&/\\\&}"
    var_value="${var_value//\[/\\[}"
    var_value="${var_value//\]/\\]}"

    echo "$var_value"
}

# ---------------------------------------------------------------------------
# update_config_var  – sets / removes a key=value pair in a config file
# Masklist mirrors the agent entrypoint (TLSPSKIdentity only).
# ---------------------------------------------------------------------------
update_config_var() {
    local config_path=$1
    local var_name=$2
    local var_value=$3
    local is_multiple=$4

    local masklist=("TLSPSKIdentity")

    if [ ! -f "$config_path" ]; then
        echo "**** Configuration file '$config_path' does not exist"
        return
    fi

    if [[ " ${masklist[@]} " =~ " $var_name " ]] && [ -n "$var_value" ]; then
        echo -n "** Updating '$config_path' parameter \"$var_name\": '****'. Enable DEBUG_MODE to view value ..."
    else
        echo -n "** Updating '$config_path' parameter \"$var_name\": '$var_value'..."
    fi

    # Remove configuration parameter when value is empty/unset
    if [ -z "$var_value" ]; then
        sed -i -e "/^$var_name=/d" "$config_path"
        echo "removed"
        return
    fi

    # Clear the value when double-quoted empty string is provided
    if [[ "$var_value" == '""' ]]; then
        if [ "$(grep -E "^$var_name=" "$config_path")" ]; then
            sed -i -e "/^$var_name=/s/=.*/=/" "$config_path"
        else
            sed -i -e "/^[#;] $var_name=/s/.*/&\n$var_name=/" "$config_path"
        fi
        echo "undefined"
        return
    fi

    # Prepend home dir for relative TLS file paths
    if [[ $var_name =~ ^TLS.*File$ ]] && [[ ! $var_value =~ ^/.+$ ]]; then
        var_value=$ZABBIX_USER_HOME_DIR/enc/$var_value
    fi

    var_value=$(escape_spec_char "$var_value")
    var_name=$(escape_spec_char "$var_name")

    if [ "$(grep -E "^$var_name=$var_value$" "$config_path")" ]; then
        echo "exists"
    elif [ "$(grep -E "^$var_name=" "$config_path")" ] && [ "$is_multiple" != "true" ]; then
        sed -i -e "/^$var_name=/s/=.*/=$var_value/" "$config_path"
        echo "updated"
    elif [ "$(grep -Ec "^# $var_name=" "$config_path")" -gt 1 ]; then
        sed -i -e "/^[#;] $var_name=$/i\\$var_name=$var_value" "$config_path"
        echo "added first occurrence"
    elif [ "$(grep -Ec "^[#;] $var_name=" "$config_path")" -gt 0 ]; then
        sed -i -e "/^[#;] $var_name=/s/.*/&\n$var_name=$var_value/" "$config_path"
        echo "added"
    else
        sed -i -e '$a\' -e "$var_name=$var_value" "$config_path"
        echo "added at the end"
    fi
}

# ---------------------------------------------------------------------------
# update_config_multiple_var  – applies comma-separated values via update_config_var
# ---------------------------------------------------------------------------
update_config_multiple_var() {
    local config_path=$1
    local var_name=$2
    local var_value=$3

    var_value="${var_value%\"}"
    var_value="${var_value#\"}"

    local IFS=,
    local OPT_LIST=($var_value)

    for value in "${OPT_LIST[@]}"; do
        update_config_var "$config_path" "$var_name" "$value" true
    done
}

# ---------------------------------------------------------------------------
# file_env  – allows VAR or VAR_FILE (Docker secrets) to supply a value
# ---------------------------------------------------------------------------
file_env() {
    local var="$1"
    local fileVar="${var}_FILE"
    local defaultValue="${2:-}"

    if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
        echo "**** Both variables $var and $fileVar are set (but are exclusive)"
        exit 1
    fi

    local val="$defaultValue"

    if [ "${!var:-}" ]; then
        val="${!var}"
        echo "** Using ${var} variable from ENV"
    elif [ "${!fileVar:-}" ]; then
        if [ ! -f "${!fileVar}" ]; then
            echo "**** Secret file \"${!fileVar}\" is not found"
            exit 1
        fi
        val="$(< "${!fileVar}")"
        echo "** Using ${var} variable from secret file"
    fi
    export "$var"="$val"
    unset "$fileVar"
}

# ---------------------------------------------------------------------------
# file_process_from_env  – writes a plain-text env var to an internal file
# ---------------------------------------------------------------------------
file_process_from_env() {
    local var_name=$1
    local file_name=$2
    local var_value=$3

    if [ -n "$var_value" ]; then
        echo -n "$var_value" > "${ZABBIX_INTERNAL_ENC_DIR}/$var_name"
        file_name="${ZABBIX_INTERNAL_ENC_DIR}/${var_name}"
    fi

    if [ -n "$var_value" ]; then
        export "$var_name"="$file_name"
    fi
    unset "${var_name%%FILE}"
}

# ---------------------------------------------------------------------------
# clear_zbx_env  – removes ZABBIX_* environment variables (agent variant)
# ---------------------------------------------------------------------------
clear_zbx_env() {
    [[ "${ZBX_CLEAR_ENV}" == "false" ]] && return

    for env_var in $(env | grep -E "^ZABBIX_"); do
        unset "${env_var%%=*}"
    done
}

# ---------------------------------------------------------------------------
# db_tls_params  – builds TLS flags for the mariadb CLI
# ---------------------------------------------------------------------------
db_tls_params() {
    local result=""

    if [ -n "${ZBX_DBTLSCONNECT}" ]; then
        result="--ssl"

        if [ "${ZBX_DBTLSCONNECT}" != "required" ]; then
            result="${result} --ssl-verify-server-cert"
        fi

        if [ -n "${ZBX_DBTLSCAFILE}" ]; then
            result="${result} --ssl-ca=${ZBX_DBTLSCAFILE}"
        fi

        if [ -n "${ZBX_DBTLSKEYFILE}" ]; then
            result="${result} --ssl-key=${ZBX_DBTLSKEYFILE}"
        fi

        if [ -n "${ZBX_DBTLSCERTFILE}" ]; then
            result="${result} --ssl-cert=${ZBX_DBTLSCERTFILE}"
        fi
    fi

    echo "$result"
}
