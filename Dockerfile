FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/package.json web/package-lock.json* ./
RUN npm ci --ignore-scripts 2>/dev/null || npm install
COPY web/ .
RUN node_modules/.bin/vite build

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./server/web/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /cb-server ./server/main.go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /cb .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates sqlite
COPY --from=builder /cb-server /usr/local/bin/cb-server
COPY --from=builder /cb /usr/local/bin/cb

EXPOSE 8080
ENV CB_DB_PATH=/data/cb.db
VOLUME /data

ENTRYPOINT ["cb-server"]
