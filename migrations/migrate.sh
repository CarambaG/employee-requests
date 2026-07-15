#!/bin/sh
set -eu

psql "$DATABASE_URL" -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
SQL

for migration in /migrations/*.up.sql; do
    version=$(basename "$migration" .up.sql)
    applied=$(psql "$DATABASE_URL" -tAc "SELECT 1 FROM schema_migrations WHERE version = '$version'")

    if [ "$applied" = "1" ]; then
        echo "migration $version already applied"
        continue
    fi

    {
        echo "BEGIN;"
        cat "$migration"
        printf "\nINSERT INTO schema_migrations (version) VALUES ('%s');\n" "$version"
        echo "COMMIT;"
    } | psql "$DATABASE_URL" -v ON_ERROR_STOP=1

    echo "migration $version applied"
done
