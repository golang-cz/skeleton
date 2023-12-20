#!/bin/bash
set -eu

function usage() {
    echo "Usage:"
    echo "  $0 import DATABASE FILES..."
    echo "  $0 dev-import DATABASE"
    echo "  $0 create DATABASE"
    echo "  $0 drop DATABASE"
    echo "  $0 terminate DATABASE"
    echo "  $0 run DATABASE QUERY"
    exit 1
}

test -z "${1-}" && usage
command="$1"
shift

test -z "${1-}" && usage
database="$1"
shift

create_user=$(cat <<EOF
    DO
    \$body\$
    BEGIN
      IF NOT EXISTS (
        SELECT * FROM pg_catalog.pg_user WHERE usename = 'devbox'
      ) THEN
        CREATE USER hubs WITH PASSWORD 'Password1';
      END IF;
    END
    \$body\$;
EOF
)

create_db=$(cat <<EOF
    CREATE DATABASE "$database" ENCODING 'UTF-8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE template0 OWNER devbox;
    \c $database

    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO devbox;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO devbox;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO devbox;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO devbox;
EOF
)

terminate=$(cat <<EOF
    SELECT pg_terminate_backend(pg_stat_activity.pid)
    FROM pg_stat_activity
    WHERE pg_stat_activity.datname = '$database' AND pid <> pg_backend_pid();
EOF
)

drop=$(cat <<EOF
    DROP DATABASE IF EXISTS "$database";
EOF
)

case "$command" in
  "create")
    echo "$terminate" "$drop" "$create_user" "$create_db" | psql -U devbox -v ON_ERROR_STOP=1
    ;;
  "drop")
    echo "$terminate" "$drop" | psql -U devbox -v ON_ERROR_STOP=1
    ;;
  "import")
    [ $# -eq 0 ] && usage
    for i in $*; do stat $i >/dev/null || exit 1; done
    echo "$terminate" "$drop" "$create_user" "$create_db" | cat - $* | psql -U devbox -v ON_ERROR_STOP=1
    ;;
  "restore")
    [ $# -eq 0 ] && usage
    for i in $*; do stat $i >/dev/null || exit 1; done
    echo "$terminate" "$drop" "$create_user" "$create_db" | psql -U devbox -v ON_ERROR_STOP=1
    set -x
    # We used to use "-j 8" arg for parallelism. But there's a bug in Postgres 12 causing
    # "pg_restore: error: could not find block ID 4682 in archive" error, see:
    # https://www.postgresql.org/message-id/flat/1582010626326-0.post%40n3.nabble.com#0891d77011cdb6ca3ad8ab7904a2ed63
    # https://www.postgresql.org/message-id/1582010626326-0.post%40n3.nabble.com
    pg_restore -U devbox -d "$database" --no-owner --no-privileges --no-acl $*
    ;;
  "run")
    psql -U devbox -v ON_ERROR_STOP=1 "$database" <&0
    ;;
  *)
    echo "$command: no such command"
    usage
    ;;
esac

exit $?
