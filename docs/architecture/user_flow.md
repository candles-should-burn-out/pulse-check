## Login flow:

1. Пользователь открывает `/app`.
2. Frontend редиректит пользователя в Keycloak.
3. Keycloak аутентифицирует пользователя и возвращает OIDC authorization response.
4. Frontend получает токены через Authorization Code Flow with PKCE.
5. Frontend вызывает защищенные backend API с `Authorization: Bearer <token>`.
6. Backend проверяет issuer, audience, lifetime, signature и опциональную role.

## Protected API flow:

1. Frontend в local development вызывает `/api/entities`.
2. Vite или nginx проксирует запрос на backend `/entities`.
3. Backend увеличивает `pulse_check_entity_list_requests_total`.
4. Backend возвращает список сущностей.