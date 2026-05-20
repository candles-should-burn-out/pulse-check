#!/usr/bin/env bash
set -euo pipefail

compose_file="${COMPOSE_FILE:-production-compose.example.yaml}"
command="${1:-up}"
if [[ $# -gt 0 ]]; then
	shift
fi

compose=(docker compose -f "$compose_file")

"${compose[@]}" build backend
"${compose[@]}" up -d pulse-check-postgres

for attempt in {1..30}; do
	if "${compose[@]}" exec -T pulse-check-postgres sh -ec 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >/dev/null 2>&1; then
		break
	fi

	if [[ "$attempt" -eq 30 ]]; then
		echo "PostgreSQL is not ready after 30 attempts" >&2
		exit 1
	fi

	sleep 2
done

"${compose[@]}" run --rm --no-deps --entrypoint /app/pulse-check-migrate backend "$command" "$@"
