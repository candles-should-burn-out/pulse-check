# Frontend Error Monitoring через OTEL/Grafana

## Summary
Собираем runtime-ошибки фронта без SaaS: React/Vite ловит ошибки, отправляет обезличенный event в backend, backend пишет структурированный JSON в Loki и увеличивает Prometheus-счётчики. Grafana показывает дашборд и алерты по частоте ошибок.

## Key Changes
- Frontend:
  - Добавить `ErrorBoundary` вокруг `<App />` в `frontend/src/main.tsx`.
  - Добавить `frontend/src/error-reporting.ts` с обработчиками `window.error`, `unhandledrejection`, React boundary errors и ручной функцией `reportFrontendError`.
  - Отправлять события на `POST /api/client-errors` через `navigator.sendBeacon`, fallback `fetch(..., { keepalive: true })`.
  - Не отправлять PII: без access token, username, localStorage, request body, query/hash URL; message и stack sanitization обязательны.
- Backend:
  - Добавить публичный `POST /client-errors`, возвращающий `202 Accepted`.
  - Ограничить payload, например `MaxBytesReader` 8 KB; валидировать JSON и enum-поля.
  - Логировать событие через `slog` как `frontend error reported` с полями `kind`, `severity`, `route`, `release`, `fingerprint`, `message`, `top_frame`, `user_agent_family`, `trace_id`, `span_id`.
  - Добавить Prometheus counter: `pulse_check_frontend_errors_total{kind,severity,route}`.
- Observability:
  - Loki уже получает backend stdout, поэтому frontend error events появятся там автоматически.
  - Добавить Grafana dashboard panel: ошибки за 5/15 минут, топ fingerprint, топ route.
  - Добавить alert rule: warning при росте ошибок, critical при sustained spike или `severity="fatal"`.

## Public API
`POST /client-errors`

Минимальный payload:
```json
{
  "kind": "react_render|window_error|unhandled_rejection|manual",
  "severity": "error|fatal",
  "message": "sanitized short message",
  "fingerprint": "stable-client-computed-hash",
  "route": "/app/profile",
  "release": "0.1.0",
  "topFrame": "App.tsx:123",
  "stackFrames": ["App.tsx:123", "main.tsx:14"]
}
```

Backend не требует авторизации для этого endpoint, потому что часть ошибок может происходить до логина.

## Test Plan
- Frontend unit tests для sanitization, fingerprint и beacon/fetch fallback.
- React test: ErrorBoundary вызывает reporter и показывает fallback UI.
- Backend tests:
  - `POST /client-errors` валидный payload -> `202`, counter incremented.
  - слишком большой payload -> `413`.
  - неверный method -> `405`.
  - invalid JSON -> `400`.
- Smoke check в local compose: искусственная ошибка во фронте видна в Grafana Loki и Prometheus metric.

## Assumptions
- Целевой контур: существующий OTEL/Grafana, без Sentry на этом этапе.
- Ошибки мониторим как обезличенные технические события; пользовательские данные, токены, URL query/hash и произвольные form/request payloads не отправляем.
- Existing frontend OTEL traces остаются для request correlation, но primary alerting строится по backend logs + Prometheus metrics.
