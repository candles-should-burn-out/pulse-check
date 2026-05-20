#!/usr/bin/env bash
set -euo pipefail

require_env() {
	local name="$1"
	if [[ -z "${!name:-}" ]]; then
		echo "$name is required" >&2
		exit 1
	fi
}

require_env PULSE_CHECK_DB_HOST
require_env PULSE_CHECK_DB_NAME
require_env PULSE_CHECK_DB_USER
require_env PULSE_CHECK_DB_PASSWORD
require_env KEYCLOAK_DB_HOST
require_env KEYCLOAK_DB_NAME
require_env KEYCLOAK_DB_USER
require_env KEYCLOAK_DB_PASSWORD
require_env BACKUP_REMOTE_HOST
require_env BACKUP_REMOTE_USER
require_env BACKUP_REMOTE_DIR
require_env BACKUP_SSH_KEY_PATH

if [[ ! -r "$BACKUP_SSH_KEY_PATH" ]]; then
	echo "BACKUP_SSH_KEY_PATH is not readable: $BACKUP_SSH_KEY_PATH" >&2
	exit 1
fi

retention_days="${BACKUP_RETENTION_DAYS:-30}"
if [[ ! "$retention_days" =~ ^[0-9]+$ ]]; then
	echo "BACKUP_RETENTION_DAYS must be a positive integer" >&2
	exit 1
fi

if ((retention_days < 1)); then
	echo "BACKUP_RETENTION_DAYS must be a positive integer" >&2
	exit 1
fi

timestamp="$(date -u +"%Y%m%dT%H%M%SZ")"
archive_name="pulse-check-backup-${timestamp}.tar.gz"
workdir="$(mktemp -d)"

cleanup() {
	rm -rf "$workdir"
}
trap cleanup EXIT

ssh_opts=(
	-i "$BACKUP_SSH_KEY_PATH"
	-o BatchMode=yes
	-o StrictHostKeyChecking=accept-new
)
remote="${BACKUP_REMOTE_USER}@${BACKUP_REMOTE_HOST}"

wait_for_database() {
	local label="$1"
	local host="$2"
	local user="$3"
	local database="$4"
	local password="$5"

	for attempt in $(seq 1 30); do
		if PGPASSWORD="$password" pg_isready -h "$host" -U "$user" -d "$database" >/dev/null 2>&1; then
			return 0
		fi

		if [[ "$attempt" -eq 30 ]]; then
			echo "$label database is not ready after 30 attempts" >&2
			exit 1
		fi

		sleep 2
	done
}

wait_for_database "Pulse Check" "$PULSE_CHECK_DB_HOST" "$PULSE_CHECK_DB_USER" "$PULSE_CHECK_DB_NAME" "$PULSE_CHECK_DB_PASSWORD"
wait_for_database "Keycloak" "$KEYCLOAK_DB_HOST" "$KEYCLOAK_DB_USER" "$KEYCLOAK_DB_NAME" "$KEYCLOAK_DB_PASSWORD"

echo "Creating Pulse Check database dump"
PGPASSWORD="$PULSE_CHECK_DB_PASSWORD" pg_dump \
	-h "$PULSE_CHECK_DB_HOST" \
	-U "$PULSE_CHECK_DB_USER" \
	-d "$PULSE_CHECK_DB_NAME" \
	-Fc \
	-f "$workdir/pulse_check.dump"

echo "Creating Keycloak database dump"
PGPASSWORD="$KEYCLOAK_DB_PASSWORD" pg_dump \
	-h "$KEYCLOAK_DB_HOST" \
	-U "$KEYCLOAK_DB_USER" \
	-d "$KEYCLOAK_DB_NAME" \
	-Fc \
	-f "$workdir/keycloak.dump"

cat >"$workdir/manifest.json" <<EOF
{
  "format": "pulse-check-postgres-backup-v1",
  "created_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "backup_type": "online",
  "databases": [
    {
      "name": "$PULSE_CHECK_DB_NAME",
      "service": "pulse-check-postgres",
      "dump": "pulse_check.dump",
      "format": "pg_dump custom"
    },
    {
      "name": "$KEYCLOAK_DB_NAME",
      "service": "keycloak-postgres",
      "dump": "keycloak.dump",
      "format": "pg_dump custom"
    }
  ],
  "retention_days": $retention_days
}
EOF

(
	cd "$workdir"
	sha256sum manifest.json pulse_check.dump keycloak.dump >checksums.txt
	tar -czf "$archive_name" manifest.json checksums.txt pulse_check.dump keycloak.dump
)

echo "Uploading backup archive to $remote:$BACKUP_REMOTE_DIR/$archive_name"
ssh "${ssh_opts[@]}" "$remote" "mkdir -p '$BACKUP_REMOTE_DIR'"
scp "${ssh_opts[@]}" "$workdir/$archive_name" "$remote:$BACKUP_REMOTE_DIR/$archive_name"

echo "Applying remote retention: deleting backups older than $retention_days days"
ssh "${ssh_opts[@]}" "$remote" \
	"find '$BACKUP_REMOTE_DIR' -type f -name 'pulse-check-backup-*.tar.gz' -mtime +$retention_days -delete"

echo "Backup completed: $archive_name"
