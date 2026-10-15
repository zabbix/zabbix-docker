#!/bin/sh

# Builds CycloneDX SBOMs for every component present in the output directory.
# Usage: sbom.sh <sources directory> <version> <output root>

set -eu

sources="$1"
version="$2"
output_root="$3"
script_dir="$(dirname "$0")"

fail() {
    echo "SBOM: $*" >&2
    exit 1
}

# The body runs in a subshell, so its variables stay local to one component
generate_component() (
    descriptor="$1"
    output_dir="$2"

    mkdir -p "$output_dir"
    output="$output_dir/$(jq -r '.component.name' "$descriptor").cdx.json"

    # Versions of vendored libraries are read from their files, so that the
    # SBOM follows upstream updates. Components built from separate sources,
    # such as agent 2 plugins, set the version in the descriptor.
    vendored='[]'
    count="$(jq '.vendored // [] | length' "$descriptor")"
    i=0
    while [ "$i" -lt "$count" ]; do
        entry="$(jq -c ".vendored[$i]" "$descriptor")"
        library_version="$(printf '%s' "$entry" | jq -r --arg version "$version" '.version // "" | split("{version}") | join($version)')"
        source_file="$(printf '%s' "$entry" | jq -r '.file // empty')"
        if [ -n "$source_file" ]; then
            source_file="$sources/$source_file"
            [ -f "$source_file" ] || fail "vendored file is missing: $source_file"

            pattern="$(printf '%s' "$entry" | jq -r '.version_pattern // empty')"
            if [ -n "$pattern" ]; then
                library_version="$(jq -rn --rawfile text "$source_file" --arg re "$pattern" '$text | capture($re).version')"
                [ -n "$library_version" ] || fail "version pattern does not match $source_file"
            fi
        fi

        vendored="$(printf '%s' "$vendored" | jq -c --argjson entry "$entry" --arg version "$library_version" '. + [$entry + {version: $version}]')"
        i=$((i + 1))
    done

    installed=/dev/null
    manifest=/dev/null
    if jq -e '.composer' "$descriptor" > /dev/null; then
        installed="$sources/$(jq -r '.composer.installed' "$descriptor")"
        manifest="$sources/$(jq -r '.composer.manifest' "$descriptor")"
        [ -f "$installed" ] || fail "composer file is missing: $installed"
        [ -f "$manifest" ] || fail "composer file is missing: $manifest"
    fi

    jq -n \
        --arg version "$version" \
        --argjson vendored "$vendored" \
        --slurpfile descriptor "$descriptor" \
        --slurpfile installed "$installed" \
        --slurpfile manifest "$manifest" \
        -f "$script_dir/cyclonedx.jq" > "$output"
)

for descriptor in "$script_dir"/*.json; do
    component="$(basename "$descriptor" .json)"
    [ -d "$output_root/$component" ] || continue
    generate_component "$descriptor" "$output_root/$component/sbom"
done
