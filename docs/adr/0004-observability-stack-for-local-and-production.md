# ADR-0004: Observability stack for local and production

Date: 2026-05-15

Status: Accepted

## Context

Разработчикам и операторам нужны health checks, metrics, traces и logs, но стек должен оставаться достаточно простым для local Docker Compose и первого VPS deployment.

Observability не должна нарушать privacy boundary из ADR-0001 и ADR-0003: telemetry data не должна становиться обходным каналом для пользовательских персональных данных.

## Decision

Использовать OpenTelemetry instrumentation в backend и frontend.

В local stack:

- OpenTelemetry Collector принимает OTLP/gRPC и OTLP/HTTP;
- Tempo хранит traces;
- Loki хранит backend logs;
- Alloy читает stdout/stderr backend container и отправляет logs в Loki;
- Prometheus собирает backend `/metrics`;
- Grafana предоставляет datasources и dashboards.

Backend access logs пишутся в JSON и содержат `trace_id` и `span_id` для correlation.

Источники истины:

- traces отправляются через OTLP endpoint;
- metrics backend-а отдаются в Prometheus text format через `/metrics`;
- logs пишутся в stdout/stderr и собираются Alloy;
- health state backend-а отдается через health endpoints.

Основные параметры конфигурации:

- `OTEL_SERVICE_NAME` - service name backend-а, по умолчанию `pulse-check-backend`;
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OTLP endpoint backend-а;
- `VITE_OTEL_EXPORTER_OTLP_ENDPOINT` - OTLP endpoint frontend-а;
- `VITE_OTEL_SERVICE_NAME` - service name frontend-а, по умолчанию `pulse-check-frontend`.

## Данные и База Данных

Pulse Check product schema не затрагивается.

Tempo, Loki, Prometheus и Grafana используют собственные локальные volumes в Docker Compose. Их внутренние схемы не являются product schema Pulse Check.

## API

### GET: /metrics

Тело запроса: отсутствует

Код ответа: `200 OK`

Ответ:

```text
# HELP pulse_check_entity_list_requests_total Total number of entity list requests.
# TYPE pulse_check_entity_list_requests_total counter
pulse_check_entity_list_requests_total 0
```

### GET: /health/live

Тело запроса: отсутствует

Код ответа: `200 OK`

Ответ:

```json
{
  "status": "ok"
}
```

### GET: /health/ready

Тело запроса: отсутствует

Код ответа: `200 OK`

Ответ:

```json
{
  "status": "ok"
}
```

Ошибка готовности:

Код ответа: `503 Service Unavailable`

```json
{
  "status": "not_ready"
}
```

### GET: /health/startup

Тело запроса: отсутствует

Код ответа: `200 OK`

Ответ:

```json
{
  "status": "ok"
}
```

Ошибка запуска:

Код ответа: `503 Service Unavailable`

```json
{
  "status": "starting"
}
```

## Ограничения

- Telemetry не должна содержать user-owned data, notes, attributes, персональные маппинги или содержимое backup.
- Access logs должны ограничиваться техническими полями: method, path, status, duration, `trace_id`, `span_id` и техническими ошибками.
- Production может стартовать меньшим observability stack, но app interfaces для metrics, traces и health checks должны сохраняться.
- OTLP endpoint должен настраиваться через environment variables и не быть жестко зашитым в код.
- Frontend tracing включается только при заданном `VITE_OTEL_EXPORTER_OTLP_ENDPOINT`.

## Consequences

- Local diagnosis можно выполнять в Grafana по metrics, logs и traces.
- Telemetry configuration остается явной в compose и environment variables.
- Backend имеет стабильные health endpoints для Docker healthcheck и production probes.
- Observability stack увеличивает эксплуатационную поверхность: нужно следить за retention, доступом к Grafana и объемом telemetry data.
- Privacy boundary распространяется на dashboards, logs, traces, metrics и alert payloads.

## Проработка

- Спроектировать production retention для Tempo, Loki и Prometheus.
- Определить минимальный production-набор dashboards и alerts.
- Спроектировать frontend error monitoring без передачи user-owned data.
