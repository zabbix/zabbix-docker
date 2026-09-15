# shellcheck shell=bash

openssl_rehash() {
    local ssl_location="$1"

    if ! openssl rehash "$ssl_location" >/dev/null; then
        warn "openssl rehash failed for '$ssl_location'"
    fi
}

openssl_prepare_ca() {
    local source_dir="$1"
    local target_dir="$2"

    rm -rf "$target_dir"/*
    cp -a "$source_dir/." "$target_dir/"
    openssl_rehash "$target_dir"
}
