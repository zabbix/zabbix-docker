openssl_rehash() {
    local ssl_location="$1"

    if command -v openssl >/dev/null 2>&1 && [ -n "$ssl_location:-}" ] && [ -d "$ssl_location" ]; then
        openssl rehash "$ssl_location" >/dev/null
    fi
}
