# Pulse Check Backend

Minimal Go HTTP service stub.

## Run

```sh
go run ./cmd/app
```

The server listens on `:8080` by default. Override it with `HTTP_ADDR`, for example:

```sh
HTTP_ADDR=:9090 go run ./cmd/app
```

## Endpoints

- `GET /entities` returns a hardcoded entity list.
- `GET /metrics` exposes `pulse_check_entity_list_requests_total` in Prometheus text format.
- `GET /swagger/` returns the OpenAPI 3.0 schema.
- `GET /health/live` and `GET /livez` are liveness probes.
- `GET /health/ready` and `GET /readyz` are readiness probes.
- `GET /health/startup` and `GET /startupz` are startup probes.

The application handles `SIGINT` and `SIGTERM`, flips readiness to false, and gracefully shuts down the HTTP server.

## Tracing

The service uses OpenTelemetry HTTP instrumentation and exports traces through OTLP/HTTP. It supports W3C Trace Context through the `Traceparent` header, and access logs include `trace_id` and `span_id`.

Useful environment variables:

- `OTEL_SERVICE_NAME` sets the service name. Default: `pulse-check-backend`.
- `OTEL_EXPORTER_OTLP_ENDPOINT` sets the OTLP endpoint. Example: `http://localhost:4318`.
- `OTEL_EXPORTER_OTLP_HEADERS` sets OTLP headers when your collector requires them.

## Docker

```sh
docker build -t pulse-check-backend:local .
```
