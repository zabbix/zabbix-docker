#!/usr/bin/env bats
# Tests for escape_spec_char() from the Zabbix Docker entrypoint scripts.

load helpers/entrypoint_functions

# ---------------------------------------------------------------------------
# Basic pass-through cases
# ---------------------------------------------------------------------------
@test "escape_spec_char: plain alphanumeric string is returned unchanged" {
    run escape_spec_char "hello123"
    [ "$status" -eq 0 ]
    [ "$output" = "hello123" ]
}

@test "escape_spec_char: empty string is returned as empty string" {
    run escape_spec_char ""
    [ "$status" -eq 0 ]
    [ "$output" = "" ]
}

# ---------------------------------------------------------------------------
# Individual special characters
# ---------------------------------------------------------------------------
@test "escape_spec_char: backslash is doubled" {
    run escape_spec_char 'a\b'
    [ "$status" -eq 0 ]
    [ "$output" = 'a\\b' ]
}

@test "escape_spec_char: forward slash is escaped" {
    run escape_spec_char "a/b"
    [ "$status" -eq 0 ]
    [ "$output" = 'a\/b' ]
}

@test "escape_spec_char: dot is escaped" {
    run escape_spec_char "a.b"
    [ "$status" -eq 0 ]
    [ "$output" = 'a\.b' ]
}

@test "escape_spec_char: asterisk is escaped" {
    run escape_spec_char "a*b"
    [ "$status" -eq 0 ]
    [ "$output" = 'a\*b' ]
}

@test "escape_spec_char: caret is escaped" {
    run escape_spec_char "a^b"
    [ "$status" -eq 0 ]
    [ "$output" = 'a\^b' ]
}

@test "escape_spec_char: dollar sign is escaped" {
    run escape_spec_char 'a$b'
    [ "$status" -eq 0 ]
    [ "$output" = 'a\$b' ]
}

@test "escape_spec_char: ampersand is escaped" {
    run escape_spec_char 'a&b'
    [ "$status" -eq 0 ]
    [ "$output" = 'a\&b' ]
}

@test "escape_spec_char: opening bracket is escaped" {
    run escape_spec_char 'a[b'
    [ "$status" -eq 0 ]
    [ "$output" = 'a\[b' ]
}

@test "escape_spec_char: closing bracket is escaped" {
    run escape_spec_char 'a]b'
    [ "$status" -eq 0 ]
    [ "$output" = 'a\]b' ]
}

@test "escape_spec_char: newlines are stripped" {
    run escape_spec_char $'a\nb'
    [ "$status" -eq 0 ]
    [ "$output" = "ab" ]
}

# ---------------------------------------------------------------------------
# Compound cases
# ---------------------------------------------------------------------------
@test "escape_spec_char: URL-like path is fully escaped" {
    run escape_spec_char "http://example.com/path"
    [ "$status" -eq 0 ]
    [ "$output" = 'http:\/\/example\.com\/path' ]
}

@test "escape_spec_char: value with multiple special chars is fully escaped" {
    run escape_spec_char 'foo.bar$baz&qux'
    [ "$status" -eq 0 ]
    [ "$output" = 'foo\.bar\$baz\&qux' ]
}
