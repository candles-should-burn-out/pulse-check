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
- `GET /health/live` and `GET /livez` are liveness probes.
- `GET /health/ready` and `GET /readyz` are readiness probes.
- `GET /health/startup` and `GET /startupz` are startup probes.

The application handles `SIGINT` and `SIGTERM`, flips readiness to false, and gracefully shuts down the HTTP server.
