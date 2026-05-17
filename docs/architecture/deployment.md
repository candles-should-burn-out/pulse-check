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

## Production checklist:

- HTTPS обязателен для frontend и Keycloak.
- `KC_HOSTNAME` должен совпадать с публичным Keycloak URL.
- Redirect URIs, Web Origins и Post Logout Redirect URIs должны содержать только production-домены.
- Первый realm import должен выполняться уже с production URL.
- Self-registration должна быть выключена.
- Bootstrap admin password нельзя оставлять дефолтным.
- Секреты должны жить вне git: env-файл, secret manager или protected CI vars.
- Для баз данных нужны регулярные backup.
- Rotation секретов должен покрывать Keycloak admin, Keycloak DB password, reverse proxy secrets и CI secrets.
- При rotation URL/секретов frontend build args и backend OIDC env должны быть синхронизированы.
- Disabled user перестает получать новые refresh/access tokens; уже выданный access token живет до expiration.