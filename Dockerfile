FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/package.json web/package-lock.json* ./
RUN npm ci --ignore-scripts 2>/dev/null || npm install
COPY web/ .
RUN node_modules/.bin/vite build

FROM golang:1.24-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./server/web/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /cb-server ./server/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /cb-server /usr/local/bin/cb-server

EXPOSE 8080
ENV CB_DB_PATH=/data/cb.db
VOLUME /data

ENTRYPOINT ["cb-server"]
