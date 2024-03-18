#!/bin/sh
# wait-for-db.sh

set -e

# Extract the host from the connection string
host=$(echo "postgresql://postgres:secret@go-posts-db?sslmode=disable" | awk -F[/:] '{print $4}')

until ping -c 1 "$host" &> /dev/null; do
  >&2 echo "Host $host is unreachable - sleeping"
  sleep 1
done

>&2 echo "Host $host is reachable - executing command"
exec "$@"