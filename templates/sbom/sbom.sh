#!/bin/sh
# Builds a CycloneDX SBOM of one Zabbix component from its descriptor. The file is named after the component, for example zabbix-server.cdx.json.
# Usage: sbom.sh <descriptor.json> <sources directory> <version> <output directory>
set -eu

descriptor="$1"
sources="$2"
version="$3"
mkdir -p "$4"
output="$4/$(jq -r '.component.name' "$descriptor").cdx.json"

fail() {
    echo "SBOM: $*" >&2
    exit 1
}

# Versions of vendored libraries are read from their files, so that the SBOM
# follows upstream updates. A library without a version pattern is listed
# without a version.
vendored='[]'
count="$(jq '.vendored // [] | length' "$descriptor")"
i=0
while [ "$i" -lt "$count" ]; do
    entry="$(jq -c ".vendored[$i]" "$descriptor")"
    file="$sources/$(printf '%s' "$entry" | jq -r '.file')"
    [ -f "$file" ] || fail "vendored file is missing: $file"

    pattern="$(printf '%s' "$entry" | jq -r '.version_pattern // empty')"
    library_version=""
    if [ -n "$pattern" ]; then
        library_version="$(jq -rn --rawfile text "$file" --arg re "$pattern" '$text | capture($re).version')"
        [ -n "$library_version" ] || fail "version pattern does not match $file"
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
    -f "$(dirname "$0")/cyclonedx.jq" > "$output"
