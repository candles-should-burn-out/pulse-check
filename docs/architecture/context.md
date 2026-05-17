# Наполнение системы (технологическое)

- Frontend - SPA/PWA на React, TypeScript, Vite, MUI, React Router, React Hook Form и IndexedDB
- Backend - Go, HTTP, PostgreSQL
- Для деплоя используем Docker Compose (При необходимости масштабирования перейдем на Docker Swarm)
- Self-hosted Keycloak, realm для OIDC-входа и администрирования, отдельный PostgreSQL для Keycloak
- Observability stack: OpenTelemetry Collector, Tempo, Loki, Alloy, Prometheus и Grafana
- Внешний reverse proxy, который отдает статику и маршрутизирует запросы к frontend, Keycloak, backend

## Авторизация:

- `pulse-check-frontend` - public SPA client с Authorization Code Flow и PKCE S256.
- `pulse-check-api` - backend audience/resource; access token frontend-клиента должен содержать `aud=pulse-check-api`.
- Первый вход пользователя должен проходить через required actions Keycloak: `VERIFY_EMAIL` и `UPDATE_PASSWORD`.
- Product APIs должны отклонять отсутствующие, невалидные, истекшие, wrong-issuer или wrong-audience токены.
- Production frontend и Keycloak должны работать через HTTPS.
- Secrets нельзя коммитить в git.

## Observability:

- Backend экспортирует трейсы по OTLP/HTTP в `otel-collector:4318`.
- Frontend экспортирует browser fetch traces в `http://localhost:4318`, когда tracing включен.
- Метрики: `backend /metrics -> Prometheus -> Grafana`
- Трейсы: `backend/frontend -> OTel Collector -> Tempo -> Grafana`
- Логи: `backend stdout/stderr -> Alloy -> Loki -> Grafana` (Alloy отправляет stdout/stderr backend-контейнера в Loki)
- Grafana получает provisioned datasources Prometheus, Tempo и Loki, а также папку dashboard-ов `Pulse Check`.