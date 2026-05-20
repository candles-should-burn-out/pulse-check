#!/usr/bin/env bash
set -euo pipefail

compose_file="${COMPOSE_FILE:-production-compose.example.yaml}"
compose=(docker compose -f "$compose_file")

"${compose[@]}" build backup
"${compose[@]}" run --rm backup
