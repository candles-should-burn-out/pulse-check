# ADR-0002: Self-hosted Keycloak for OIDC

Status: Accepted

## Context

Pulse Check нужны authentication, administrator-managed users, блокировка пользователей, password reset и стандартная OIDC-интеграция. Проект не должен хранить пароли, реализовывать собственную регистрацию или lifecycle учетных записей.

Решение должно работать в local Docker Compose и переноситься в первый production deployment. Frontend и backend находятся в разных сетевых контекстах: browser-visible issuer отличается от backend JWKS URL внутри Docker network.

## Decision

Использовать self-hosted Keycloak с PostgreSQL как identity provider для OIDC.

Ключевые правила:

- `/` публичен;
- `/app/*` требует login;
- user self-registration выключена;
- базовый login - email и password;
- password reset выполняется через Keycloak;
- администраторы управляют пользователями через Keycloak Admin Console;
- первый вход требует `VERIFY_EMAIL` и `UPDATE_PASSWORD`;
- `pulse-check-frontend` - public SPA client с Authorization Code Flow и PKCE S256;
- `pulse-check-api` - backend audience/resource.

Keycloak является источником истины для пользователей, паролей, блокировок и required actions. Backend проверяет Bearer JWT через JWKS, issuer, audience, lifetime, `nbf` и optional role.

## User-flow

### Exist Flow: Открытие публичной страницы

Потребитель: anonymous user

Пользователь открывает `/` без login.

### Exist Flow: Вход в приложение

Потребитель: user

Пользователь открывает `/app/*`, проходит login через Keycloak и возвращается во frontend с OIDC-сессией.

### Exist Flow: Сброс пароля

Потребитель: user

Пользователь запускает password reset в Keycloak. Pulse Check не обрабатывает пароль напрямую.

### Exist Flow: Управление пользователями

Потребитель: admin

Администратор создает, блокирует и настраивает пользователей в Keycloak Admin Console.

## Configuration

Backend:

- `OIDC_ISSUER` - issuer, ожидаемый backend-ом;
- `OIDC_JWKS_URL` - JWKS endpoint, доступный backend-у;
- `OIDC_AUDIENCE` - expected audience, по умолчанию `pulse-check-api`;
- `OIDC_REQUIRED_ROLE` - optional роль для protected API.

Frontend:

- `VITE_KEYCLOAK_URL` - browser-visible Keycloak URL;
- `VITE_KEYCLOAK_REALM` - realm, по умолчанию `pulse-check`;
- `VITE_KEYCLOAK_CLIENT_ID` - frontend client, по умолчанию `pulse-check-frontend`.

`OIDC_ISSUER`, `OIDC_JWKS_URL` и `OIDC_AUDIENCE` должны быть заданы вместе. Frontend OIDC-параметры являются build-time Vite configuration.

## Database

Pulse Check product schema не затрагивается.

Keycloak использует отдельную PostgreSQL базу. Ее схема управляется миграциями Keycloak и не описывается таблицами Pulse Check.

## API

Новые продуктовые backend-ручки не добавляются.

Контракт доступа для protected backend API:

- клиент передает `Authorization: Bearer <access_token>`;
- backend возвращает `401 Unauthorized`, если токен отсутствует, недействителен, просрочен, выпущен не тем issuer или не содержит нужный audience;
- backend возвращает `403 Forbidden`, если задан `OIDC_REQUIRED_ROLE`, но токен не содержит эту роль.

Ответ ошибки:

```json
{
  "error": "invalid_token"
}
```

## Страницы

### /app/login

Назначение: frontend route, который запускает OIDC login и возвращает пользователя в `/app/`.

Доступ: anonymous user. Authenticated user перенаправляется в `/app/`.

Основные элементы:

- состояние проверки авторизации;
- сообщение об ошибке, если Keycloak initialization недоступен.

Действия:

- открыть страницу: frontend создает Keycloak login URL и делает browser redirect;
- вернуться browser back/forward navigation: frontend возвращает пользователя на `/`.

Состояния:

- загрузка авторизации;
- ошибка авторизации;
- пользователь уже authenticated.

### Keycloak /realms/{realm}/protocol/openid-connect/auth

Назначение: страница входа, которую рендерит Keycloak theme `pulse-check`.

Доступ: anonymous user в рамках OIDC Authorization Code Flow.

Основные элементы:

- поле login/email;
- поле password;
- ссылка на reset password, если `resetPasswordAllowed` включен;
- сообщения Keycloak об ошибках.

Действия:

- войти: Keycloak проверяет credentials и возвращает authorization response во frontend redirect URI;
- перейти к reset password: Keycloak открывает страницу сброса пароля.

Состояния:

- обычный ввод credentials;
- ошибка credentials или required action;
- успешный redirect обратно во frontend.

### Keycloak /realms/{realm}/login-actions/reset-credentials

Назначение: страница сброса пароля, которую рендерит Keycloak theme `pulse-check`.

Доступ: anonymous user, если reset password разрешен в realm.

Основные элементы:

- поле login/email;
- ссылка возврата ко входу;
- сообщения Keycloak об ошибках.

Действия:

- отправить форму: Keycloak запускает reset password flow;
- вернуться ко входу: Keycloak открывает login page.

Состояния:

- обычный ввод login/email;
- ошибка валидации или неизвестная учетная запись;
- успешная отправка reset-инструкций.

## Ограничения

- Production realm configuration должна бэкапиться вместе с PostgreSQL базой Keycloak.
- Redirect URIs и web origins должны быть явно настроены для каждого окружения.
- Если `OIDC_REQUIRED_ROLE` пустой, backend проверяет токен без дополнительного role gate.
- Доступность Keycloak обязательна для login и protected frontend routes.

## Consequences

- Администраторская ответственность за пользователей переносится в Keycloak Admin Console.
- Backend APIs получают единый authentication boundary через Bearer JWT.
- Browser-visible issuer и backend JWKS URL настраиваются отдельно из-за различий Docker и browser networking.
- Pulse Check зависит от operational maturity Keycloak deployment.

## Проработка

- Описать production backup/restore для Keycloak realm configuration и PostgreSQL.
- Зафиксировать правила выпуска и назначения ролей, если `OIDC_REQUIRED_ROLE` начнет использоваться в production.
