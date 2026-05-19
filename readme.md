# Pulse Check

Pulse Check - privacy-oriented local-first приложение для учета сущностей, статусов и агрегированной статистики. Пользовательские названия, заметки, поля, атрибуты и маппинги по умолчанию остаются на устройстве пользователя; сервер хранит только технические идентификаторы, статусы, счетчики и служебные данные, необходимые для работы сервиса.

## Документация

- Архитектура:
  - [Introduction](docs/architecture/Introduction.md) - краткое описание продукта и архитектурного фокуса.
  - [Context](docs/architecture/context.md) - технологический контекст, авторизация и observability.
  - [Constraints](docs/architecture/constraints.md) - внешние, продуктовые и технологические ограничения системы.
  - [Deployment](docs/architecture/deployment.md) - local и production stack, reverse proxy и production checklist.
  - [Terms](docs/architecture/terms.md) - основные продуктовые и технические термины.
  - [User flow](docs/architecture/user_flow.md) - login flow и protected API flow.
- Architecture Decision Records:
  - [ADR-0001: Local-first and Bring Your Own Backup](docs/adr/0001-local-first-and-bring-your-own-backup.md) - граница локального хранения пользовательских данных и пользовательских backup/export/import.
  - [ADR-0002: Self-hosted Keycloak for OIDC](docs/adr/0002-self-hosted-keycloak-for-oidc.md) - выбор Keycloak для authentication и OIDC.
  - [ADR-0003: Server stores only technical aggregates](docs/adr/0003-server-stores-only-technical-aggregates.md) - правило хранения на сервере только технических агрегатов без пользовательских маппингов.
  - [ADR-0004: Observability stack for local and production](docs/adr/0004-observability-stack-for-local-and-production.md) - health checks, metrics, traces, logs и observability stack.
  - [ADR-0005: Серверно управляемые наборы статусов](docs/adr/0005-server-managed-status-sets.md) - server-managed status sets, роли владельца/участника и API статусов.
  - [ADR-0006: Приглашенные участники и агрегированная статистика](docs/adr/0006-invited-participants.md) - приглашения участников, membership graph и агрегированная статистика.
- [OpenAPI схема backend API](docs/api/openapi.json) - контракт backend API в формате OpenAPI 3.0.
- Database:
  - [status_sets](docs/database/status_sets.md) - наборы статусов и информация о владельцах.
  - [status_set_memberships](docs/database/status_set_memberships.md) - связи пользователей с наборами статусов.
  - [statuses](docs/database/statuses.md) - определения статусов внутри наборов: название, цвета и timestamps.
- Правила:
  - [adr_rules](docs/rules/adr_rules.md) - формат ADR, API, database-описаний и связанных разделов.
  - [feature_contribution_rules](docs/rules/feature_contribution_rules.md) - workflow разработки фичи.
- [TODO](docs/todo/todo.md) - рабочие заметки и открытые задачи.
- [Backend](backend/readme.md) - запуск, endpoints, tracing и backend-specific настройки.
- [Frontend](frontend/readme.md) - запуск, сборка и frontend-specific настройки.

### Markdown style

В Markdown-файлах не делаем ручной перенос строк только из-за достижения лимита длины строки. В среде разработки включен визуальный автоперенос, поэтому абзацы пишем одной строкой; переносы оставляем для смысловой структуры: новые абзацы, списки, таблицы и fenced code blocks.

## Локальная разработка

Если нужны локальные секреты или переопределения, скопируйте `.env.example` в `.env`. Файл `.env` не коммитится.

Запуск полного локального стенда:

```sh
task up
```

Перезапуск только application-сервисов:

```sh
task backend-restart
task frontend-restart
```

Полезные локальные URL:

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Keycloak Admin Console: `http://localhost:8081/admin`
- Grafana: `http://localhost:3001`

Локальный тестовый пользователь: `admin` / `admin`

## Проверки

```sh
cd backend && go test ./...
cd frontend && npm run lint
cd frontend && npm run build
```
