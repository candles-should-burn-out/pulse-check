# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/pulse-check-backend ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/pulse-check-migrate ./cmd/migrate

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /bin/pulse-check-backend /app/pulse-check-backend
COPY --from=build /bin/pulse-check-migrate /app/pulse-check-migrate
COPY migrations /app/migrations

ENV HTTP_ADDR=:8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health/ready >/dev/null || exit 1

ENTRYPOINT ["/app/pulse-check-backend"]
