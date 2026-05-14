# Pulse Check Frontend

Frontend stub for the Pulse Check backend.

The public landing page is served at `/`. The working application is under
`/app/*` and uses Keycloak OIDC login before calling protected backend APIs.

## Stack

- React
- TypeScript
- Vite
- MUI
- React Router
- React Hook Form is included as a dependency for upcoming forms, but is not used by the current entity-list stub.

## Run

Install dependencies:

```sh
npm install
```

Start the backend from `../backend`:

```sh
go run ./cmd/app
```

Start the frontend:

```sh
npm run dev
```

By default, the Vite dev server proxies `/api/*` to `http://localhost:8080/*`, so the entity list button calls `GET /api/entities` in the browser and reaches backend `GET /entities`.

To point the frontend at another API base URL without the dev proxy, create `.env.local`:

```sh
VITE_API_BASE_URL=http://localhost:8080
```

Keycloak settings are build-time Vite variables:

```sh
VITE_KEYCLOAK_URL=http://localhost:8081
VITE_KEYCLOAK_REALM=pulse-check
VITE_KEYCLOAK_CLIENT_ID=pulse-check-frontend
```

The local Docker stack provides these values automatically.

## Tracing

The frontend uses OpenTelemetry Web SDK and fetch instrumentation. Tracing is enabled when `VITE_OTEL_EXPORTER_OTLP_ENDPOINT` is set.

For the local compose stack, the browser sends OTLP/HTTP traces to the collector through the host-published port:

```sh
VITE_OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
VITE_OTEL_SERVICE_NAME=pulse-check-frontend
```

The exporter appends `/v1/traces` automatically when the endpoint does not include it.

## Docker

Build the nginx image with the static frontend bundle:

```sh
docker build -t pulse-check-frontend:local .
```

Run the container:

```sh
docker run --rm -p 8080:80 pulse-check-frontend:local
```

The frontend uses `/api` as the default API base URL. To bake another API URL into the production bundle, pass `VITE_API_BASE_URL` at build time:

```sh
docker build --build-arg VITE_API_BASE_URL=http://backend:8080 -t pulse-check-frontend:local .
```

To bake Keycloak settings into the production bundle, pass:

```sh
docker build \
  --build-arg VITE_KEYCLOAK_URL=https://auth.example.com \
  --build-arg VITE_KEYCLOAK_REALM=pulse-check \
  --build-arg VITE_KEYCLOAK_CLIENT_ID=pulse-check-frontend \
  -t pulse-check-frontend:local .
```

To bake browser trace export into the bundle, pass the collector endpoint too:

```sh
docker build --build-arg VITE_OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 -t pulse-check-frontend:local .
```
