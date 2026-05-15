# ADR-0002: Self-hosted Keycloak for OIDC

Date: 2026-05-15

Status: Accepted

## Context

Pulse Check нужны authentication, administrator-managed users, отключение и блокировка пользователей, password reset и стандартная OIDC-интеграция. Проект не должен реализовывать хранение паролей или собственную регистрацию.

## Decision

Использовать self-hosted Keycloak с PostgreSQL.

Access model:

- `/` публичен.
- `/app/*` требует login.
- User self-registration выключена.
- Базовый login - email и password.
- Администраторы управляют пользователями через Keycloak Admin Console.
- Первый вход должен требовать `VERIFY_EMAIL` и `UPDATE_PASSWORD`.
- `pulse-check-frontend` - public SPA client с Authorization Code Flow и PKCE S256.
- `pulse-check-api` - backend audience/resource.

## Consequences

- Pulse Check не хранит пароли и не реализует регистрацию.
- Backend APIs проверяют Bearer JWT через JWKS, issuer, audience, lifetime и optional role.
- Browser-visible issuer и backend JWKS URL настраиваются отдельно, потому что Docker networking и browser networking различаются.
- Keycloak realm configuration становится production-critical и должна бэкапиться вместе с PostgreSQL базой Keycloak.
