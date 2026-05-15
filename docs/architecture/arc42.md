# Архитектура Pulse Check

Документ использует arc42 как легкий Markdown-формат для архитектурной документации Pulse Check. ADR фиксируют причины ключевых решений, а этот документ дает актуальный обзор архитектуры.

## 1. Introduction and Goals

Pulse Check - local-first приложение для учета сущностей, их статусов и агрегированной статистики с минимизацией серверного знания о пользовательских данных.

Основные цели:

- Хранить пользовательские данные локально по умолчанию: названия сущностей, заметки, пользовательские поля, атрибуты и маппинги `ID -> название/человек/заметка` остаются на устройстве пользователя.
- Поддержать модель Bring Your Own Backup: пользователь сам экспортирует, хранит и восстанавливает зашифрованные бэкапы.
- Дать защищенный доступ к продукту без хранения паролей в Pulse Check и без собственной реализации user management.
- Поддержать будущую модель подчиненных/помощников: они помогают собирать статистику по статусам, а руководитель видит агрегаты без конкретных сущностей.
- Сохранить простой путь деплоя на VPS через Docker Compose.

Основные стейкхолдеры:

- Пользователи: им нужны локальная доступность, приватность, импорт/экспорт и понятная ответственность за данные, которые они вводят.
- Администраторы: им нужны управление пользователями, блокировки, сброс пароля и настройка realm через Keycloak.
- Разработчики и операторы: им нужны локальный стенд, health checks, метрики, трейсы, логи и повторяемый production deployment.

## 2. Constraints

- Ограничения по персональным данным:
  - Сервис не требует и не инициирует внесение персональных данных третьих лиц.
  - Пользователь самостоятельно определяет, какие сведения хранить локально.
  - Названия сущностей, заметки, пользовательские поля и пользовательские статусы по умолчанию хранятся локально.
  - Персональные данные нельзя передавать в аналитику, crash reporting, поддержку, логи или внешние сервисы без явного действия пользователя.
  - Пользователь не должен использовать в названиях статусов персональные, чувствительные, оскорбительные, дискриминационные или специальные категории персональных данных.
- Авторизация делегирована self-hosted Keycloak с PostgreSQL.
- Frontend - SPA/PWA на React, TypeScript, Vite, MUI, React Router, React Hook Form и IndexedDB.
- Backend - Go HTTP service.
- Локальный и первый production deployment используют Docker Compose.
- Production-цель - один VPS за внешним HTTPS reverse proxy.
- Frontend должен поддерживать offline-сценарии через очередь действий до подтверждения сервером.

## 3. Context and Scope

Pulse Check состоит из:

- Browser/PWA frontend для пользователей.
- Backend API для защищенных product operations.
- Keycloak realm для OIDC-входа и администрирования.
- PostgreSQL база Keycloak, используемая только Keycloak.
- Локальный observability stack: OpenTelemetry Collector, Tempo, Loki, Alloy, Prometheus и Grafana.
- Внешний reverse proxy в production, который терминирует HTTPS и маршрутизирует запросы к frontend и Keycloak.

Публичные и защищенные точки входа:

- `/` - публичная стартовая страница.
- `/app/*` - рабочая область, доступная после входа.
- Публичные backend endpoints: health checks, `/metrics`, `/swagger/`.
- Защищенные backend endpoints: `/entities` и будущие product API.

Текущий backend endpoint `/entities` пока является stub-реализацией, но граница уже задана: product API требуют валидный Bearer JWT, когда настроены OIDC env vars.

## 4. Solution Strategy

- Browser/PWA является владельцем пользовательских персональных данных.
- Для сущностей используются UUID, генерируемые на стороне сервера.
- Offline-действия превращаются в очередь событий, ожидающих подтверждения сервером.
- Backend хранит только технические идентификаторы, статусы и агрегированные счетчики, например `entity_id`, `state_id`, `name`, `client_id`, `status_id` и `count`.
- Backup реализуется через зашифрованный экспорт/импорт вместо серверного backup-хранилища.
- Авторизация и администрирование выполняются через Keycloak OIDC.
- Операционная видимость строится на OpenTelemetry instrumentation и локальном Grafana stack.

## 5. Building Block View

Frontend:

- React/TypeScript/Vite SPA с публичной страницей `/` и защищенной рабочей областью `/app/*`.
- Использует Keycloak login перед вызовом защищенных backend API.
- Планируемые PWA-возможности: `manifest.webmanifest`, service worker, Workbox cache, offline app shell, cache busting при обновлениях, install prompt для Android и инструкция "Add to Home Screen" для iOS.
- Планируемое локальное хранение: IndexedDB, опциональное локальное шифрование и `navigator.storage.persist()` там, где браузер это поддерживает.

Backend:

- Go HTTP service с `/entities`, `/metrics`, `/swagger/` и health endpoints.
- Проверяет JWT-подпись через JWKS, а также `iss`, `aud`, `exp` и `nbf`.
- Поддерживает опциональную проверку realm/resource role через `OIDC_REQUIRED_ROLE`.
- Пишет access logs с `trace_id` и `span_id`.

Авторизация:

- `pulse-check-frontend` - public SPA client с Authorization Code Flow и PKCE S256.
- `pulse-check-api` - backend audience/resource; access token frontend-клиента должен содержать `aud=pulse-check-api`.
- Пользовательская регистрация выключена.
- Администраторы создают, отключают, блокируют пользователей и сбрасывают пароль через Keycloak Admin Console.
- Первый вход пользователя должен проходить через required actions Keycloak: `VERIFY_EMAIL` и `UPDATE_PASSWORD`.

Observability:

- Backend экспортирует трейсы по OTLP/HTTP в `otel-collector:4318`.
- Frontend экспортирует browser fetch traces в `http://localhost:4318`, когда tracing включен.
- Prometheus собирает backend `/metrics`.
- Alloy отправляет stdout/stderr backend-контейнера в Loki.
- Tempo хранит трейсы.
- Grafana получает provisioned datasources Prometheus, Tempo и Loki, а также папку dashboard-ов `Pulse Check`.

## 6. Runtime View

Login flow:

1. Пользователь открывает `/app`.
2. Frontend редиректит пользователя в Keycloak.
3. Keycloak аутентифицирует пользователя и возвращает OIDC authorization response.
4. Frontend получает токены через Authorization Code Flow with PKCE.
5. Frontend вызывает защищенные backend API с `Authorization: Bearer <token>`.
6. Backend проверяет issuer, audience, lifetime, signature и опциональную role.

Protected API flow:

1. Frontend в local development вызывает `/api/entities`.
2. Vite или nginx проксирует запрос на backend `/entities`.
3. Backend увеличивает `pulse_check_entity_list_requests_total`.
4. Backend возвращает список сущностей.

Offline and sync flow:

1. Пользователь выполняет действия offline.
2. Frontend сохраняет действия локально как queued events.
3. Когда сеть и авторизация доступны, frontend отправляет события на сервер для acknowledgement.
4. Conflict resolution и retry policy пока TBD.

Backup/import flow:

1. Пользователь явно запускает export backup.
2. Frontend создает зашифрованный архив с `manifest.json`, `data.json`, опциональным `attachments/` и `checksums.json`.
3. Для ZIP планируется `fflate`, для валидации - Zod, для checksums - Web Crypto API/SHA-256.
4. Пользователь самостоятельно хранит backup и позже явно импортирует его.

Telemetry flow:

- Трейсы: `backend/frontend -> OTel Collector -> Tempo -> Grafana`.
- Логи: `backend stdout/stderr -> Alloy -> Loki -> Grafana`.
- Метрики: `backend /metrics -> Prometheus -> Grafana`.

## 7. Deployment View

Local stack:

- `local-compose.yaml` запускает backend, frontend, Keycloak, Keycloak PostgreSQL, Keycloak seed, OpenTelemetry Collector, Tempo, Loki, Alloy, Prometheus и Grafana.
- Keycloak доступен на `http://localhost:8081`.
- Frontend доступен на `http://localhost:3000`.
- Grafana доступна на `http://localhost:3001`.
- `keycloak-seed` применяет локальную тему входа, разрешает localhost redirect URIs и создает локального тестового пользователя `admin` / `admin`.

Production stack:

- `production-compose.example.yaml` - шаблон первого деплоя на один VPS.
- Внешний reverse proxy терминирует HTTPS и проксирует:
  - frontend-домен на `127.0.0.1:3000`;
  - Keycloak-домен на `127.0.0.1:8081`.
- Backend доступен только внутри compose network.
- Keycloak импортирует `keycloak/pulse-check-realm.json` только если realm еще не существует.

Production checklist:

- HTTPS обязателен для frontend и Keycloak.
- `KC_HOSTNAME` должен совпадать с публичным Keycloak URL.
- Redirect URIs, Web Origins и Post Logout Redirect URIs должны содержать только production-домены.
- Первый realm import должен выполняться уже с production URL.
- Self-registration должна быть выключена.
- Bootstrap admin password нельзя оставлять дефолтным.
- Секреты должны жить вне git: env-файл, secret manager или protected CI vars.
- Для `keycloak-postgres` нужен регулярный backup.
- Rotation секретов должен покрывать Keycloak admin, Keycloak DB password, reverse proxy secrets и CI secrets.
- При rotation URL/секретов frontend build args и backend OIDC env должны быть синхронизированы.
- Disabled user перестает получать новые refresh/access tokens; уже выданный access token живет до expiration.

## 8. Crosscutting Concepts

Authentication and authorization:

- Pulse Check не хранит пароли.
- Pulse Check не реализует собственную регистрацию.
- Invite codes, если появятся, будут отдельным admission layer и не заменят Keycloak user management.
- Browser и backend могут видеть Keycloak по разным адресам, поэтому `OIDC_ISSUER` и `OIDC_JWKS_URL` настраиваются отдельно.

Privacy and data minimization:

- User-owned names, notes, attributes и personal mappings остаются локальными по умолчанию.
- Server APIs не должны получать пользовательские персональные данные, если будущая явно задокументированная функция не изменит эту границу.
- Logs, traces, metrics, support channels и analytics не должны получать персональные данные без явного действия пользователя.

Backup and local persistence:

- Backup запускается пользователем и принадлежит пользователю.
- Backup files шифруются.
- Целостность данных проверяется checksums.
- Browser persistent storage нужно запрашивать там, где он доступен, но это не
  гарантия backup.

Identifiers and events:

- Entity использует UUID, сгенерированный сервером.
- Offline actions представлены как events, ожидающие server acknowledgement.
- Aggregation должна строиться на technical identifiers и counters, а не на
  personal labels.

Observability:

- Access logs пишутся в JSON и содержат trace correlation IDs.
- Metrics должны описывать поведение сервиса без утечки пользовательских данных.
- Local Grafana используется для development/diagnosis, а не как product analytics sink.

## 9. Architectural Decisions

- [ADR-0001: Local-first and Bring Your Own Backup](../adr/0001-local-first-and-bring-your-own-backup.md)
- [ADR-0002: Self-hosted Keycloak for OIDC](../adr/0002-self-hosted-keycloak-for-oidc.md)
- [ADR-0003: Server stores only technical aggregates](../adr/0003-server-stores-only-technical-aggregates.md)
- [ADR-0004: Observability stack for local and production](../adr/0004-observability-stack-for-local-and-production.md)

## 10. Quality Requirements

Privacy:

- Backend не должен получать entity names, notes, custom fields, attributes или personal mappings в default architecture.
- Telemetry и support workflows не должны передавать персональные данные.

Offline availability:

- Установленный frontend должен оставаться пригодным для локальной работы без сети.
- Offline actions должны локально сохраняться до server acknowledgement.

Security:

- Product APIs должны отклонять отсутствующие, невалидные, истекшие, wrong-issuer или wrong-audience токены.
- Production frontend и Keycloak должны работать через HTTPS.
- Secrets нельзя коммитить в git.

Recoverability:

- Пользователь должен иметь возможность export/import encrypted backups.
- Operators должны делать backup Keycloak PostgreSQL в production.

Operability:

- Backend должен отдавать health endpoints.
- Backend должен отдавать Prometheus metrics.
- Logs и traces должны коррелироваться через trace IDs.

## 11. Risks and Technical Debt

- Потеря устройства может привести к потере пользовательских данных, если пользователь не сделал и не сохранил backup.
- Browser storage может быть очищен; persistent storage повышает шансы, но не является backup guarantee.
- Offline conflict resolution пока не специфицирован.
- Backup UX и encryption/key handling пока не реализованы.
- Keycloak realm drift возможен после первого import, потому что существующий realm не перезатирается последующими `--import-realm` запусками.
- Ошибка production URL при первом realm import может сохраниться в базе Keycloak и потребовать ручного исправления через Admin Console.
- Access token остается валидным до expiration после disabled user.
- Текущий backend `/entities` - stub и должен эволюционировать до реальной product data модели.

## 12. Glossary

| Term                     | Meaning                                                                                                             |
|--------------------------|---------------------------------------------------------------------------------------------------------------------|
| Entity                   | Отслеживаемая сущность с server-generated UUID. User-visible names или notes локальны по умолчанию.                 |
| Status / State           | Статус сущности, используемый для локального отображения и агрегированной статистики.                               |
| Subordinate              | Помощник, который может участвовать в сборе статистики по статусам без раскрытия конкретных сущностей руководителю. |
| Backup                   | Пользовательский зашифрованный export, который позже можно import.                                                  |
| `client_id`              | Технический идентификатор для aggregate counters.                                                                   |
| `entity_id`              | Технический UUID сущности.                                                                                          |
| `state_id` / `status_id` | Технический идентификатор state/status для aggregates.                                                              |
| `count`                  | Агрегированное количество для пары client/status.                                                                   |
| OIDC                     | OpenID Connect protocol, используемый через Keycloak.                                                               |
| JWKS                     | JSON Web Key Set для проверки подписи tokens backend-ом.                                                            |
