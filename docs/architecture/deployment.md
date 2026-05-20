## Local stack:

- `local-compose.yaml` запускает backend, frontend, Keycloak, Keycloak PostgreSQL, Keycloak seed, OpenTelemetry Collector, Tempo, Loki, Alloy, Prometheus и Grafana.
- Keycloak доступен на `http://localhost:8081`.
- Frontend доступен на `http://localhost:3000`.
- Grafana доступна на `http://localhost:3001`.
- `keycloak-seed` применяет локальную тему входа, разрешает localhost redirect URIs и создает локального тестового пользователя `admin` / `admin`.

## Production stack:

- `production-compose.example.yaml` - шаблон первого деплоя на один VPS.
- Внешний reverse proxy терминирует HTTPS и проксирует:
  - frontend-домен на `127.0.0.1:3000`;
  - Keycloak-домен на `127.0.0.1:8081`.
- Backend доступен только внутри compose network.
- Keycloak импортирует `keycloak/pulse-check-realm.json` только если realm еще не существует.
- Backup для `pulse-check-postgres` и `keycloak-postgres` выполняется отдельным compose-сервисом `backup` из profile `ops`.

## Production backup:

- Backup запускается командой `COMPOSE_FILE=production-compose.example.yaml scripts/backup-production.sh`.
- Рекомендуемый schedule для VPS - один раз в сутки через host cron или systemd timer.
- Backup выполняется online: `backend` и `keycloak` не останавливаются, поэтому две базы могут отличаться на несколько секунд.
- Архив называется `pulse-check-backup-YYYYMMDDTHHMMSSZ.tar.gz`.
- Архив содержит `manifest.json`, `checksums.txt`, `pulse_check.dump` и `keycloak.dump`.
- Дампы создаются через `pg_dump -Fc` внутри backup-контейнера на базе PostgreSQL 17 client.
- Архив отправляется на удаленный сервер по SSH. Дополнительное шифрование архива не применяется.
- Retention выполняется на удаленном сервере: `pulse-check-backup-*.tar.gz` старше `BACKUP_RETENTION_DAYS` удаляются.
- Любая ошибка backup завершает скрипт с non-zero exit code.

Backup configuration:

- `BACKUP_REMOTE_HOST` - SSH host удаленного сервера backup.
- `BACKUP_REMOTE_USER` - SSH user на удаленном сервере.
- `BACKUP_REMOTE_DIR` - директория хранения backup-архивов на удаленном сервере.
- `BACKUP_SSH_KEY_FILE` - путь на production host к private SSH key для отправки backup.
- `BACKUP_RETENTION_DAYS` - срок хранения ежедневных backup, по умолчанию `30`.

## Production restore:

- Restore запускается из локального архива: `COMPOSE_FILE=production-compose.example.yaml scripts/restore-production.sh ./pulse-check-backup-YYYYMMDDTHHMMSSZ.tar.gz`.
- Restore может скачать архив с backup-сервера: `COMPOSE_FILE=production-compose.example.yaml scripts/restore-production.sh --remote pulse-check-backup-YYYYMMDDTHHMMSSZ.tar.gz`.
- Перед изменением баз скрипт проверяет `checksums.txt`, наличие обоих dump-файлов и требует вручную ввести `RESTORE`.
- Restore выполняется с downtime: скрипт останавливает `backend` и `keycloak`, но оставляет `pulse-check-postgres` и `keycloak-postgres` запущенными.
- Для каждой базы скрипт пересоздает database и применяет `pg_restore --exit-on-error --no-owner`.
- После восстановления скрипт запускает backend migrations, затем поднимает `backend` и `keycloak`.

## Production checklist:

- HTTPS обязателен для frontend и Keycloak.
- `KC_HOSTNAME` должен совпадать с публичным Keycloak URL.
- Redirect URIs, Web Origins и Post Logout Redirect URIs должны содержать только production-домены.
- Первый realm import должен выполняться уже с production URL.
- Self-registration должна быть выключена.
- Bootstrap admin password нельзя оставлять дефолтным.
- Секреты должны жить вне git: env-файл, secret manager или protected CI vars.
- Для баз данных нужны регулярные backup.
- Backup SSH private key должен жить вне git и быть доступен только production host.
- Restore из backup должен периодически проверяться на staging или отдельном production clone.
- Rotation секретов должен покрывать Keycloak admin, Keycloak DB password, reverse proxy secrets и CI secrets.
- При rotation URL/секретов frontend build args и backend OIDC env должны быть синхронизированы.
- Disabled user перестает получать новые refresh/access tokens; уже выданный access token живет до expiration.
