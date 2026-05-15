# Архитектура

- Full Local-First + Bring Your Own Backup
- Атрибуты могут быть разных типов: номер / почта / тег / заметка
- Для сущностей используются UUID (генерируются на стороне сервера)
  - В офлайн режиме действия превращаются в очередь эвентов ждущих подтверждение учета от сервера
- Можно создавать подчиненных (тех кто будет помогать собирать статистику) - они получают информацию о используемых статусах и их статистика отображается у руководителя (обобщенная по статусам, без конкретных сущностей)
- Возможность экспорта / импорта зашифрованного бэкапа
- Резервное копирование выполняет сам пользователь

## Обработка персональных данных

- Все пользовательские данные хранятся только на устройстве пользователя
- Сервер не получает и не хранит: Атрибуты / маппинг `ID → название / человек / заметка` / персональные данные третьих лиц
- Запрещено передавать пользовательские данные в: аналитику / crash reporting / поддержку / внешние сервисы без явного действия пользователя

## Frontend

- TODO: Доступ к приложению можно защищать паролем или биометрией
- TODO: Данные локально можно шифровать
- SPA / PWA: пользователь открывает ссылку → "Add to Home Screen" / "Установить приложение"
- Стек: React / TypeScript / Vite / MUI / React Router / React Hook Form / IndexedDB

### PWA

- vite-plugin-pwa, Workbox
- manifest.webmanifest
- service worker
- offline cache app shell
- cache busting при обновлениях
- install prompt для Android
- инструкция “Add to Home Screen” для iOS

### Типовая структура:

```
src/
    app/
    db/
    sync/
    features/
public/
    manifest.webmanifest
    icons/
service-worker.ts
```

### Persistent storage

```
if (navigator.storage?.persist) {
    const granted = await navigator.storage.persist();
    console.log("Persistent storage:", granted);
}
```

### Backup-файл

- Для ZIP: fflate
- Для валидации: Zod
- Для контрольных сумм: Web Crypto API / SHA-256
- Содержимое: manifest.json / data.json / attachments / checksums.json

## Backend

- Сервер хранит только идентификаторы сущностей (entity_id), статусы (state_id, name) и агрегированные счётчики (client_id, status_id, count)

## Авторизация

Выбран self-hosted Keycloak + PostgreSQL.

Причина выбора: нужна бесплатная open-source система управления пользователями с несколькими администраторами, блокировками, сбросом пароля и стандартным OIDC. Logto OSS рассматривался, но не подошёл из-за ограничения на одну админ-учётку в self-hosted OSS варианте.

Модель доступа:

- `/` — публичная стартовая страница без авторизации.
- `/app/*` — рабочая область, доступна только после входа.
- Пользовательская регистрация выключена.
- Базовый вход: email + password.
- Администраторы создают, отключают, блокируют пользователей и сбрасывают пароль через Keycloak Admin Console.
- Первый вход пользователя должен проходить через required actions Keycloak: `VERIFY_EMAIL` и `UPDATE_PASSWORD`.
- Pulse Check не хранит пароли и не реализует собственную регистрацию.
- Система пригласительных кодов будет отдельным слоем допуска и не заменяет Keycloak user management.

OIDC-клиенты:

- `pulse-check-frontend` — public SPA client, Authorization Code Flow + PKCE S256.
- `pulse-check-api` — backend audience/resource. Access token frontend-клиента должен содержать `aud=pulse-check-api`.

Backend защищает product API через Bearer JWT:

- публичные endpoints: health checks, `/metrics`, `/swagger/`;
- защищенные endpoints: `/entities` и будущие product API;
- проверяются подпись через JWKS, `iss`, `aud`, `exp`, `nbf`;
- Keycloak `issuer` и внутренний Docker `JWKS URL` настраиваются отдельно, потому что браузер и backend видят Keycloak по разным адресам.

## Локальный стенд

`local-compose.yaml` поднимает приложение и инфраструктуру для просмотра трейсов и логов бекенда:

- `backend` экспортирует трейсы по OTLP/HTTP в `otel-collector:4318`
- `frontend` экспортирует браузерные fetch-трейсы по OTLP/HTTP в `http://localhost:4318`
- `keycloak` поднимает локальную OIDC-авторизацию на `http://localhost:8081`
- `keycloak-seed` применяет локальную тему входа и создаёт тестового пользователя `admin` / `admin`
- `keycloak-postgres` хранит состояние Keycloak в named volume `keycloak-postgres-data`
- `otel-collector` принимает OTLP/gRPC на `4317` и OTLP/HTTP на `4318`
- `tempo-init` подготавливает права на volume `tempo-data`
- `tempo` хранит трейсы и открывает HTTP API на `3200`
- `loki` хранит логи бекенда и открывает HTTP API на `3100`
- `alloy` читает stdout/stderr Docker-контейнера `backend` и отправляет записи в Loki
- `prometheus` собирает метрики бекенда с `/metrics` и открывает HTTP API на `9090`
- `grafana` доступна на `http://localhost:3001` с заранее настроенными datasource `Prometheus`, `Tempo` и `Loki`

Поток трейсов: `backend/frontend -> OTel Collector -> Tempo -> Grafana`.
Поток логов: `backend stdout/stderr -> Alloy -> Loki -> Grafana`.
Поток метрик: `backend /metrics -> Prometheus -> Grafana`.

В Grafana автоматически создаётся папка `Pulse Check` с дашбордом `Backend entity list requests` для метрики `pulse_check_entity_list_requests_total`.
Состояние Grafana хранится в named volume `grafana-data`, поэтому сохранённые через UI изменения дашбордов переживают перезапуск контейнеров.

Логи бекенда можно смотреть в Grafana через `Explore -> Loki`. Базовый LogQL-запрос:

```logql
{compose_project="pulse-check", compose_service="backend"}
```

Access logs бекенда пишутся в JSON и содержат `trace_id`, поэтому из логов в Grafana можно переходить к связанным трейсам в Tempo.

Конфиги наблюдаемости лежат в `observability/`:

- `observability/alloy/config.alloy`
- `observability/loki.yaml`
- `observability/otel-collector.yaml`
- `observability/prometheus.yaml`
- `observability/tempo.yaml`
- `observability/grafana/datasources/tempo.yaml`
- `observability/grafana/dashboards/pulse-check.yaml`
- `observability/grafana/dashboard-definitions/entity-list-requests.json`

Запуск, когда понадобится:

```sh
task up
```

Перезапуск только приложения без перезапуска Keycloak, Grafana и остальной инфраструктуры:

```sh
task backend-restart
task frontend-restart
```

Локальная проверка авторизации:

1. Открыть frontend: `http://localhost:3000`.
2. Открыть Keycloak Admin Console: `http://localhost:8081/admin`.
3. Войти локальным bootstrap-админом из `.env` или дефолтом `admin` / `pulse-check-local-admin-password`.
4. Перейти в realm `pulse-check`.
5. Для быстрой проверки открыть `http://localhost:3000/app` и войти тестовым пользователем `admin` / `admin`.
6. Для проверки пользовательского сценария создать отдельного пользователя, указать email, включить `Email verified` для локального стенда или настроить SMTP.
7. В `Credentials` задать временный пароль.
8. Войти созданным пользователем и загрузить сущности.

Для локальной настройки можно скопировать `.env.example` в `.env` и заменить секреты. Файл `.env` не коммитится.

## Production: VPS + Docker Compose

Шаблон первого production-деплоя лежит в `production-compose.example.yaml`. Он рассчитан на один VPS, где внешний reverse proxy терминирует HTTPS и проксирует:

- публичный frontend домен на `127.0.0.1:3000`;
- публичный Keycloak домен на `127.0.0.1:8081`.

Минимальный порядок запуска:

1. Скопировать `.env.example` в защищённый env-файл на сервере.
2. Задать реальные секреты `KEYCLOAK_ADMIN_PASSWORD` и `KEYCLOAK_DB_PASSWORD`.
3. Задать публичные URL:
   - `APP_PUBLIC_URL=https://app.example.com`
   - `KEYCLOAK_PUBLIC_URL=https://auth.example.com`
   - `VITE_KEYCLOAK_URL=https://auth.example.com`
   - `OIDC_ISSUER=https://auth.example.com/realms/pulse-check`
   - `OIDC_JWKS_URL=http://keycloak:8080/realms/pulse-check/protocol/openid-connect/certs`
4. Настроить HTTPS reverse proxy и заголовки `X-Forwarded-*`.
5. Поднять стек через `docker compose --env-file <env-file> -f production-compose.example.yaml up -d --build`.
6. До первого запуска заменить localhost Redirect URIs, Web Origins и Post Logout Redirect URIs в `keycloak/pulse-check-realm.json` на production-домены.
7. Настроить SMTP в realm, иначе verify email и password reset не будут нормально работать.

`production-compose.example.yaml` запускает Keycloak с `--import-realm` и монтирует `keycloak/pulse-check-realm.json`. Keycloak импортирует realm при старте только если realm с таким именем еще не существует; при последующих рестартах существующий `pulse-check` realm не перезатирается.

Production checklist:

- HTTPS обязателен для frontend и Keycloak.
- `KC_HOSTNAME` должен совпадать с публичным Keycloak URL.
- Redirect URIs, Web Origins и Post Logout Redirect URIs должны содержать только реальные production-домены.
- Первый import realm должен выполняться уже с production URL, иначе localhost-настройки попадут в базу Keycloak и их придется править через Admin Console.
- Self-registration в realm должна быть выключена.
- Bootstrap admin password нельзя оставлять дефолтным.
- Секреты должны жить вне git: env-файл, secret manager или protected CI variables.
- Нужен регулярный backup базы `keycloak-postgres`.
- Нужен понятный порядок rotation секретов: Keycloak admin, Keycloak DB password и секреты reverse proxy/CI.
- Перед ротацией URL/секретов проверить, что frontend build args и backend OIDC env синхронизированы.
- Disabled user перестает получать новые refresh/access tokens; уже выданный access token живет до истечения срока.

## Test plan

Автоматическая проверка:

```sh
cd backend && go test ./...
cd frontend && npm run lint
cd frontend && npm run build
```

Ручная проверка:

- `/` открывается без авторизации.
- `/app` редиректит на Keycloak.
- После входа frontend отправляет `Authorization: Bearer <token>` и API работает.
- Disabled user теряет доступ после истечения текущего access token или при обновлении сессии.
- Logout возвращает пользователя на публичную страницу.
