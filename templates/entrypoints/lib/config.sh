# shellcheck shell=bash

source "${ENTRYPOINT_LIBS}/logging.sh"
source "${ENTRYPOINT_LIBS}/format.sh"

# usage: file_env VAR [DEFAULT]
# as example: file_env 'MYSQL_PASSWORD' 'zabbix'
#    (will allow for "$MYSQL_PASSWORD_FILE" to fill in the value of "$MYSQL_PASSWORD" from a file)
# unsets the VAR_FILE afterwards and just leaving VAR
file_env() {
    local var="$1"
    local fileVar="${var}_FILE"
    local defaultValue="${2:-}"

    if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
        error "**** Both variables $var and $fileVar are set (but are exclusive)"
    fi

    local val="$defaultValue"

    if [ "${!var:-}" ]; then
        val="${!var}"
        info "** Using ${var} variable from ENV"
    elif [ "${!fileVar:-}" ]; then
        if [ ! -f "${!fileVar}" ]; then
            error "**** Secret file \"${!fileVar}\" is not found"
        fi
        val="$(< "${!fileVar}")"
        info "** Using ${var} variable from secret file"
    fi
    export "$var"="$val"
    unset "$fileVar"
}

is_masked_config_var() {
    local var_name="${1:-}"
    case "$var_name" in
        TLSPSKIdentity) return 0 ;;
        DBPassword) return 0 ;;
        *) return 1 ;;
    esac
}

update_config_var() {
    local config_path="${1:-}"
    local var_name="${2:-}"
    local var_value="${3:-}"
    local is_multiple="${4:-false}"
    local log_message

    [[ -f "$config_path" ]] || error "Missing configuration file: $config_path"

    if is_masked_config_var "$var_name" && [ -n "$var_value" ]; then
        log_message="** Updating $config_path parameter '$var_name': '****'. Enable DEBUG_MODE to view value..."
    else
        log_message="** Updating $config_path parameter '$var_name': '$var_value'..."
    fi

    # Remove configuration parameter definition in case of unset or empty parameter value
    if [ -z "$var_value" ]; then
        sed -i -e "/^${var_name}=/d" "$config_path"
        info "$log_message removed"
        return
    fi

    # Remove value from configuration parameter in case of set to double quoted parameter value
    if [[ "$var_value" == '""' ]]; then
        if grep -qE "^${var_name}=" "$config_path"; then
            sed -i -e "/^${var_name}=/s/=.*/=/" "$config_path"
        else
            sed -i -e "/^[#;] ${var_name}=/s/.*/&\n${var_name}=/" "$config_path"
        fi
        info "$log_message undefined"
        return
    fi

    # Use full path to a file for TLS related configuration parameters
    if [[ $var_name =~ ^TLS.*File$ ]] && [[ ! $var_value =~ ^/.+$ ]]; then
        var_value="${ZABBIX_USER_HOME_DIR}/enc/${var_value}"
    fi

    # Escaping characters in parameter value and name
    var_value_raw=$var_value
    var_name_raw=$var_name
    var_value="$(escape_special_chars "$var_value")"
    var_name="$(escape_special_chars "$var_name")"

    if grep -qE "^${var_name}=${var_value}$" "$config_path"; then
        log_message="$log_message exists"
    elif grep -qE "^${var_name}=" "$config_path" && [ "$is_multiple" != "true" ]; then
        sed -i -e "/^${var_name}=/s/=.*/=${var_value}/" "$config_path"
        log_message="$log_message updated"
    elif [ "$(grep -Ec "^# ${var_name}=" "$config_path")" -gt 1 ]; then
        sed -i -e "/^[#;] ${var_name}=$/i\\${var_name}=${var_value}" "$config_path"
        log_message="$log_message added first occurrence"
    elif [ "$(grep -Ec "^[#;] ${var_name}=" "$config_path")" -gt 0 ]; then
        sed -i -e "/^[#;] ${var_name}=/s/.*/&\n${var_name}=${var_value}/" "$config_path"
        log_message="$log_message added"
    else
        printf '\n%s=%s\n' "$var_name_raw" "$var_value_raw" >> "$config_path"
        log_message="$log_message added at the end"
    fi

    info "$log_message"
}

update_config_multiple_var() {
    local config_path="${1:-}"
    local var_name="${2:-}"
    local var_value="${3:-}"
    local value

    var_value="${var_value%\"}"
    var_value="${var_value#\"}"

    local IFS=,
    read -r -a opt_list <<< "$var_value"

    for value in "${opt_list[@]}"; do
        update_config_var "$config_path" "$var_name" "$value" true
    done
}

file_process_from_env() {
    local dir_name="${1:-}"
    local var_name="${2:-}"
    local file_name="${3:-}"
    local var_value="${4:-}"

    if [ -n "$var_value" ]; then
        file_name="${dir_name}/${var_name}"
        printf '%s' "$var_value" > "$file_name"
        export "$var_name=$file_name"
    fi

    # Remove variable with plain text data, for example ZBX_TLSCAFILE -> ZBX_TLSCA
    unset "${var_name%%FILE}"
}
