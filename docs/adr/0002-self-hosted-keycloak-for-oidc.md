# ADR-0002: Self-hosted Keycloak for OIDC

Status: Accepted

## Context

Pulse Check нужны authentication, administrator-managed users, отключение и блокировка пользователей, password reset и стандартная OIDC-интеграция. Проект не должен реализовывать хранение паролей или собственную регистрацию.

Решение должно работать в local Docker Compose и переноситься в первый production deployment. Frontend и backend находятся в разных сетевых контекстах: browser-visible issuer отличается от backend JWKS URL внутри Docker network.

## Decision

Использовать self-hosted Keycloak с PostgreSQL как identity provider для OIDC.

Ключевые правила access model:

- `/` публичен;
- `/app/*` требует login;
- user self-registration выключена;
- базовый login - email и password;
- password reset выполняется через Keycloak;
- администраторы управляют пользователями через Keycloak Admin Console;
- первый вход должен требовать `VERIFY_EMAIL` и `UPDATE_PASSWORD`;
- `pulse-check-frontend` - public SPA client с Authorization Code Flow и PKCE S256;
- `pulse-check-api` - backend audience/resource.

Backend проверяет Bearer JWT через JWKS, issuer, audience, lifetime, `nbf` и optional role. Источником истины для пользователей, паролей, блокировок и required actions является Keycloak.

Основные параметры конфигурации:

- `OIDC_ISSUER` - issuer, ожидаемый backend-ом;
- `OIDC_JWKS_URL` - JWKS endpoint, доступный backend-у;
- `OIDC_AUDIENCE` - expected audience, по умолчанию `pulse-check-api`;
- `OIDC_REQUIRED_ROLE` - optional роль для доступа к protected API;
- `VITE_KEYCLOAK_URL` - browser-visible Keycloak URL;
- `VITE_KEYCLOAK_REALM` - realm, по умолчанию `pulse-check`;
- `VITE_KEYCLOAK_CLIENT_ID` - frontend client, по умолчанию `pulse-check-frontend`.

## Данные и База Данных

Pulse Check product schema не затрагивается.

Keycloak использует отдельную PostgreSQL базу. Ее схема управляется миграциями Keycloak и не описывается таблицами Pulse Check.

## API

Новые продуктовые backend-ручки не добавляются.

Контракт доступа для protected backend API:

- клиент передает `Authorization: Bearer <access_token>`;
- backend возвращает `401 Unauthorized`, если токен отсутствует, недействителен, просрочен, выпущен не тем issuer или не содержит нужный audience;
- backend возвращает `403 Forbidden`, если настроен `OIDC_REQUIRED_ROLE`, но токен не содержит эту роль;
- ответ ошибок имеет JSON-формат:

```json
{
  "error": "invalid_token"
}
```

## Ограничения

- Pulse Check не хранит пароли и не реализует регистрацию пользователей.
- Production realm configuration является production-critical и должна бэкапиться вместе с PostgreSQL базой Keycloak.
- `OIDC_ISSUER`, `OIDC_JWKS_URL` и `OIDC_AUDIENCE` должны быть заданы вместе; частичная настройка считается ошибкой конфигурации.
- Frontend OIDC-параметры являются build-time Vite configuration.
- Redirect URIs и web origins должны быть явно настроены для каждого окружения.
- Если `OIDC_REQUIRED_ROLE` пустой, backend проверяет токен без дополнительного role gate.

## Consequences

- Pulse Check не хранит пароли и не реализует lifecycle учетных записей.
- Администраторская ответственность за пользователей переносится в Keycloak Admin Console.
- Backend APIs получают единый authentication boundary через Bearer JWT.
- Browser-visible issuer и backend JWKS URL настраиваются отдельно, потому что Docker networking и browser networking различаются.
- Доступность Keycloak становится обязательной для login и protected frontend routes.

## Проработка

- Описать production-процесс backup/restore Keycloak realm configuration и Keycloak PostgreSQL.
- Зафиксировать правила выпуска и назначения ролей, если `OIDC_REQUIRED_ROLE` начнет использоваться в production.
