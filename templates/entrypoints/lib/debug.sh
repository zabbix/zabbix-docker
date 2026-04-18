# shellcheck shell=bash
debug_err_trap() {
    local rc=$?
    printf 'ERROR: exit code %s in %s:%s:%s: %s\n' \
        "$rc" "${BASH_SOURCE[1]##*/}" "${FUNCNAME[1]:-main}" "${BASH_LINENO[0]}" "$BASH_COMMAND" >&2
    exit "$rc"
}

enable_debug_mode() {
    [[ "${DEBUG_MODE:-}" == "true" ]] || return 0

    export PS4='+ ${BASH_SOURCE##*/} : ${FUNCNAME[0]:-main} : ${LINENO}: '
    set -Ex
    trap debug_err_trap ERR
}

enable_debug_mode
