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

- `GET /entities` returns a hardcoded entity list and requires a valid Keycloak Bearer token when OIDC env vars are set.
- `GET /metrics` exposes `pulse_check_entity_list_requests_total` in Prometheus text format.
- `GET /health/live` and `GET /livez` are liveness probes.
- `GET /health/ready` and `GET /readyz` are readiness probes.
- `GET /health/startup` and `GET /startupz` are startup probes.

The OpenAPI 3.0 schema is maintained in [`../docs/api/openapi.json`](../docs/api/openapi.json).

The application handles `SIGINT` and `SIGTERM`, flips readiness to false, and gracefully shuts down the HTTP server.

## Tracing

The service uses OpenTelemetry HTTP instrumentation and exports traces through OTLP/HTTP. It supports W3C Trace Context through the `Traceparent` header, and access logs include `trace_id` and `span_id`.

Useful environment variables:

- `OIDC_ISSUER` sets the expected token issuer. Example: `http://localhost:8081/realms/pulse-check`.
- `OIDC_JWKS_URL` sets the JWKS endpoint used by the backend. In Docker this can differ from `OIDC_ISSUER`. Example: `http://keycloak:8080/realms/pulse-check/protocol/openid-connect/certs`.
- `OIDC_AUDIENCE` sets the required token audience. Default local value: `pulse-check-api`.
- `OIDC_REQUIRED_ROLE` optionally requires a realm or resource role in access tokens.
- `OTEL_SERVICE_NAME` sets the service name. Default: `pulse-check-backend`.
- `OTEL_EXPORTER_OTLP_ENDPOINT` sets the OTLP endpoint. Example: `http://localhost:4318`.
- `OTEL_EXPORTER_OTLP_HEADERS` sets OTLP headers when your collector requires them.

## Docker

```sh
docker build -t pulse-check-backend:local .
```
