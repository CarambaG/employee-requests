#!/bin/sh
set -eu

EMPLOYEE_COUNT=${EMPLOYEE_COUNT:-1000}
REQUEST_COUNT=${REQUEST_COUNT:-1000000}

validate_positive_integer() {
    name=$1
    value=$2

    case "$value" in
        ''|*[!0-9]*)
            echo "$name must be a positive integer" >&2
            exit 1
            ;;
    esac

    if [ "$value" -lt 1 ]; then
        echo "$name must be greater than zero" >&2
        exit 1
    fi
}

validate_positive_integer "EMPLOYEE_COUNT" "$EMPLOYEE_COUNT"
validate_positive_integer "REQUEST_COUNT" "$REQUEST_COUNT"

echo "seeding database: employees=$EMPLOYEE_COUNT requests=$REQUEST_COUNT"

psql "$DATABASE_URL" \
    -v ON_ERROR_STOP=1 \
    -v employee_count="$EMPLOYEE_COUNT" \
    -v request_count="$REQUEST_COUNT" \
    -f /seed/seed.sql
