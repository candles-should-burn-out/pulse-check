# Pulse Check Frontend

Frontend stub for the Pulse Check backend.

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
