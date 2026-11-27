#!/usr/bin/env bash
# The j-cicd backend-test step (.cicd.yaml): the Go tests against a throwaway Postgres container,
# so the Postgres-only tests (row-level security, two-user tests, migrations, concurrency) run.
# TEST_DATABASE_REQUIRED=1 makes them fail instead of skip if Postgres isn't reachable
# (internal/database/pgtest). Needs Docker. Run from backend/: ../dev/ci/backend-test.sh
set -euo pipefail

IMAGE=postgres:16-alpine # same as docker-compose.yml, prod and the e2e container
name="st-backend-test-$$-$RANDOM"
cleanup() { docker rm -f "$name" >/dev/null 2>&1 || true; }
trap cleanup EXIT

# tmpfs and no durability: fast, and thrown away with the container.
docker run -d --name "$name" -p 127.0.0.1::5432 \
  --tmpfs /var/lib/postgresql/data:rw,size=1g --shm-size 256m \
  -e POSTGRES_PASSWORD=postgres "$IMAGE" \
  -c fsync=off -c synchronous_commit=off -c full_page_writes=off -c max_connections=200 >/dev/null
port="$(docker port "$name" 5432/tcp | head -n1 | sed 's/.*://')"

# The image's init runs a socket-only server first, so a TCP check passes only once the real
# server is up.
for _ in $(seq 1 120); do
  if docker exec "$name" pg_isready -q -h 127.0.0.1 -U postgres; then break; fi
  sleep 0.5
done
docker exec "$name" pg_isready -q -h 127.0.0.1 -U postgres || { echo "backend-test: Postgres did not start"; docker logs "$name"; exit 1; }

export TEST_DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:${port}/postgres?sslmode=disable"
export TEST_DATABASE_REQUIRED=1
go test ./...
go test -tags e2e ./internal/e2e/...
