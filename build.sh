#!/usr/bin/env bash
set -euo pipefail

error() {
    printf 'Error: %s\n' "$*" >&2
    exit 1
}

info() {
    printf '%s\n' "$*"
}

require_file() {
    local file="${1:-}"
    [ -f "$file" ] || error "'$file' is missing"
}

get_dockerfile_arg() {
    local arg_name="${1:-}"
    grep -E "^ARG[[:space:]]+${arg_name}=" Dockerfile | head -n1 | cut -d= -f2- || true
}

resolve_vcs_ref() {
    local version="${1:-}"

    if [ "$version" != "local" ]; then
        git ls-remote https://git.zabbix.com/scm/zbx/zabbix.git "refs/tags/$version" | awk '{print substr($1,1,10); exit}'
        return
    fi

    local major_version zbx_version_raw
    major_version="$(get_dockerfile_arg "MAJOR_VERSION")"
    zbx_version_raw="$(get_dockerfile_arg "ZBX_VERSION")"

    [ -n "$major_version" ] || error "Unable to extract ARG MAJOR_VERSION from Dockerfile"
    [ -n "$zbx_version_raw" ] || error "Unable to extract ARG ZBX_VERSION from Dockerfile"

    if [ "$zbx_version_raw" = '${MAJOR_VERSION}' ]; then
        printf '%s\n' "$major_version"
    else
        printf '%s\n' "${major_version}.${zbx_version_raw%%.*}"
    fi
}

resolve_container_runtime() {
    if command -v docker >/dev/null 2>&1; then
        printf 'docker\n'
    elif command -v podman >/dev/null 2>&1; then
        printf 'podman\n'
    else
        error "Build command requires docker or podman"
    fi
}

validate_version() {
    local version="${1:-}"

    if [ "$version" = "local" ]; then
        return 0
    fi

    [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || error "Incorrect version syntax: '$version'"
}

build_image() {
    local runtime="${1:-}"
    local image_tag="${2:-}"
    local vcs_ref="${3:-}"

    DOCKER_BUILDKIT=1 "$runtime" build \
        -t "$image_tag" \
        --build-context sources="../../../sources" \
        --build-context config_templates="../../../templates/config" \
        --build-arg "VCS_REF=$vcs_ref" \
        --build-arg "BUILD_DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
        -f Dockerfile .
}

main() {
    require_file "Dockerfile"

    local os version app_component vcs_ref runtime image_name
    os="${PWD##*/}"
    version="${1:-local}"
    app_component="$(basename $(cd .. && pwd))"

    validate_version "$version"

    vcs_ref="$(resolve_vcs_ref "$version")"
    [ -n "$vcs_ref" ] || error "Unable to resolve VCS_REF for version '$version'"

    runtime="$(resolve_container_runtime)"
    image_name="zabbix-${app_component}:${os}-${vcs_ref}"

    info "Building image '$image_name' using $runtime"
    build_image "$runtime" "$image_name" "$vcs_ref"
}

main "$@"
