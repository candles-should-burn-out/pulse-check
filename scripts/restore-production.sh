#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat >&2 <<'USAGE'
Usage:
  scripts/restore-production.sh <local-backup-archive.tar.gz>
  scripts/restore-production.sh --remote <backup-archive-name.tar.gz>

Environment:
  COMPOSE_FILE            defaults to production-compose.example.yaml
  ENV_FILE                optional env file to source before remote download, defaults to .env when present
  BACKUP_REMOTE_HOST      required for --remote
  BACKUP_REMOTE_USER      required for --remote
  BACKUP_REMOTE_DIR       required for --remote
  BACKUP_SSH_KEY_FILE     required for --remote
USAGE
	exit 2
}

require_env() {
	local name="$1"
	if [[ -z "${!name:-}" ]]; then
		echo "$name is required" >&2
		exit 1
	fi
}

load_env_file() {
	local env_file="${ENV_FILE:-.env}"
	if [[ -f "$env_file" ]]; then
		set -a
		# shellcheck disable=SC1090
		source "$env_file"
		set +a
	fi
}

restore_database() {
	local service="$1"
	local dump_path="$2"

	"${compose[@]}" exec -T "$service" sh -ec 'dropdb --if-exists -U "$POSTGRES_USER" "$POSTGRES_DB"'
	"${compose[@]}" exec -T "$service" sh -ec 'createdb -U "$POSTGRES_USER" "$POSTGRES_DB"'
	"${compose[@]}" exec -T "$service" sh -ec 'pg_restore --exit-on-error --no-owner -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <"$dump_path"
}

if [[ $# -lt 1 ]]; then
	usage
fi

compose_file="${COMPOSE_FILE:-production-compose.example.yaml}"
compose=(docker compose -f "$compose_file")
tmpdir="$(mktemp -d)"

cleanup() {
	rm -rf "$tmpdir"
}
trap cleanup EXIT

if [[ "${1:-}" == "--remote" ]]; then
	if [[ $# -ne 2 ]]; then
		usage
	fi

	load_env_file
	require_env BACKUP_REMOTE_HOST
	require_env BACKUP_REMOTE_USER
	require_env BACKUP_REMOTE_DIR
	require_env BACKUP_SSH_KEY_FILE

	archive="$tmpdir/$2"
	scp \
		-i "$BACKUP_SSH_KEY_FILE" \
		-o BatchMode=yes \
		-o StrictHostKeyChecking=accept-new \
		"$BACKUP_REMOTE_USER@$BACKUP_REMOTE_HOST:$BACKUP_REMOTE_DIR/$2" \
		"$archive"
else
	if [[ $# -ne 1 ]]; then
		usage
	fi

	archive="$1"
fi

if [[ ! -f "$archive" ]]; then
	echo "Backup archive not found: $archive" >&2
	exit 1
fi

tar -xzf "$archive" -C "$tmpdir"

for file in manifest.json checksums.txt pulse_check.dump keycloak.dump; do
	if [[ ! -f "$tmpdir/$file" ]]; then
		echo "Backup archive is missing $file" >&2
		exit 1
	fi
done

(
	cd "$tmpdir"
	sha256sum -c checksums.txt
)

cat <<EOF
This will replace both production PostgreSQL databases from:
  $archive

Affected services:
  - pulse-check-postgres
  - keycloak-postgres

The script will stop backend and keycloak during restore.
EOF

read -r -p 'Type RESTORE to continue: ' confirmation
if [[ "$confirmation" != "RESTORE" ]]; then
	echo "Restore cancelled" >&2
	exit 1
fi

"${compose[@]}" stop backend keycloak

restore_database pulse-check-postgres "$tmpdir/pulse_check.dump"
restore_database keycloak-postgres "$tmpdir/keycloak.dump"

"${compose[@]}" run --rm --no-deps --entrypoint /app/pulse-check-migrate backend up
"${compose[@]}" up -d backend keycloak

echo "Restore completed"
