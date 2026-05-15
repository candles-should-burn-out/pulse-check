# ADR-0004: Observability stack for local and production

Date: 2026-05-15

Status: Accepted

## Context

Разработчикам и операторам нужны health checks, metrics, traces и logs, но стек должен оставаться достаточно простым для local Docker Compose и первого VPS deployment.

## Decision

Использовать OpenTelemetry instrumentation в backend и frontend. В local stack:

- OpenTelemetry Collector принимает OTLP/gRPC и OTLP/HTTP.
- Tempo хранит traces.
- Loki хранит backend logs.
- Alloy читает stdout/stderr backend container и отправляет logs в Loki.
- Prometheus собирает backend `/metrics`.
- Grafana предоставляет datasources и dashboards.

Backend access logs пишутся в JSON и содержат `trace_id` и `span_id` для correlation.

## Consequences

- Local diagnosis можно выполнять в Grafana по metrics, logs и traces.
- Telemetry configuration остается явной в compose и environment variables.
- Telemetry не должна становиться каналом передачи personal data.
- Production может стартовать меньшим observability stack, но app interfaces для metrics, traces и health checks уже определены.
