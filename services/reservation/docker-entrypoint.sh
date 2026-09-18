#!/bin/sh
set -eu

config=/app/config.yaml
if [ "${1:-}" = --config ]; then
    config=$2
fi
dsn=$(yq -er '.database.url' "$config")
directory=$(yq -er '.directory' /app/migrations.yaml)
lock=$(yq -er '.lock_timeout' /app/migrations.yaml)
duration=$(yq -er '.timeout' /app/migrations.yaml)
timeout "$duration" atlas migrate apply --dir "file://$directory" --url "$dsn" --lock-timeout "$lock"
exec reservation "$@"
