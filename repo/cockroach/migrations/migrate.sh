#!/usr/bin/env sh

force=false
uuid_extension=false

OPTIND=1

while getopts 'f;e' opt; do
    case $opt in
        f) force=true ;;
        e) uuid_extension=true;;
        *) echo 'Error in command line parsing' >&2
            exit 1
    esac
done
shift "$(( OPTIND - 1 ))"

if "$force"; then
    cockroach sql --insecure -e "DROP DATABASE IF EXISTS passwordless CASCADE"
fi

cockroach sql --insecure -e "CREATE DATABASE IF NOT EXISTS passwordless"

if "$uuid_extension"; then
    cockroach sql --insecure -e "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\""
fi

cat $(dirname $0)/000_schema.sql | cockroach sql --insecure -d passwordless
