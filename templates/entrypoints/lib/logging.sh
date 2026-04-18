# shellcheck shell=bash

timestamp() {
    date -u '+%Y-%m-%dT%H:%M:%SZ'
}
log() {
    local type="$1"
    shift

    printf '%s [%s]: %s\n' "$(timestamp)" "$type" "$*"
}
info() {
    log "info" "$@"
}
warn() {
    log "warning" "$@" >&2
}
error() {
    log "error" "$@" >&2
    exit 1
}
