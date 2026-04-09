if [[ "${DEBUG_MODE:-}" == "true" ]]; then
    export PS4='+ ${BASH_SOURCE##*/} : ${FUNCNAME[0]:-main} : ${LINENO}: '
    set -Ex
    trap 'rc=$?; printf "ERROR: exit code %s in %s:%s:%s: %s\n" \
         "$rc" "${BASH_SOURCE[0]##*/}" "${FUNCNAME[1]:-main}" "${BASH_LINENO[0]}" "$BASH_COMMAND" >&2; exit "$rc"' ERR
fi
