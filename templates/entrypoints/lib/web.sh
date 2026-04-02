prepare_zbx_config() {
    [ -n "${ZBX_SESSION_NAME:-}" ] || return 0

    local defines_file="${ZABBIX_WWW_ROOT}/include/defines.inc.php"
    local tmp_file
    tmp_file="$(mktemp)"

    if [ ! -f "$defines_file" ]; then
        error "Missing file: $defines_file"
    fi

    sed "s/\(ZBX_SESSION_NAME',[[:space:]]*'\)[^']*\('.*\)/\1${ZBX_SESSION_NAME}\2/" \
        "$defines_file" > "$tmp_file"

    cat "$tmp_file" > "$defines_file"
}
