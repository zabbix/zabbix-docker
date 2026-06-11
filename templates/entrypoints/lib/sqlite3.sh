# shellcheck shell=bash

source "${ENTRYPOINT_LIBS}/logging.sh"

create_db_schema() {
    local db_schema_file="${1:-}"

    if ! command -v sqlite3 >/dev/null 2>&1; then
        return 0
    fi

    if [ -z "${ZBX_DB_NAME:-}" ]; then
        error "Missing variable! ZBX_DB_NAME must be set"
    fi

    local db_dir
    db_dir="$(dirname "${ZBX_DB_NAME}")"

    if [ ! -d "${db_dir}" ]; then
        info "** SQLite database directory '${db_dir}' does not exist. Creating..."
        mkdir -p "${db_dir}"
    fi

    if [ ! -f "${ZBX_DB_NAME}" ]; then
        info "** SQLite database '${ZBX_DB_NAME}' does not exist. Creating..."

        zcat "${db_schema_file}" | sqlite3 "${ZBX_DB_NAME}" >/dev/null
        sqlite3 "${ZBX_DB_NAME}" 'PRAGMA journal_mode=WAL;' >/dev/null
    else
        if ! sqlite3 "${ZBX_DB_NAME}" 'PRAGMA schema_version;' >/dev/null 2>&1; then
            error "File '${ZBX_DB_NAME}' exists but is not a valid SQLite database"
        fi

        info "** SQLite database '${ZBX_DB_NAME}' already exists."
    fi
}
