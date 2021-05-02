#!/usr/bin/env sh

force=false
pgcrypto=false

OPTIND=1

while getopts 'f;e' opt; do
    case $opt in
        f) force=true ;;
        e) pgcrypto=true;;
        *) echo 'Error in command line parsing' >&2
            exit 1
    esac
done
shift "$(( OPTIND - 1 ))"

if "$force"; then
    cockroach sql --insecure -e "DROP DATABASE IF EXISTS passwordless CASCADE"
fi

cockroach sql --insecure -e "CREATE DATABASE IF NOT EXISTS passwordless"

if "$pgcrypto"; then
    cockroach sql --insecure -e "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\""
fi

cat $(dirname $0)/000_schema.sql | cockroach sql --insecure -d passwordless
