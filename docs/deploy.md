# Deployment Guide

## Environment Variables

| Variable          | Required | Default              | Description                        |
|-------------------|----------|----------------------|------------------------------------|
| `CB_JWT_SECRET`   | **Yes**  |                      | Secret for signing JWT tokens. Must be changed from any default. |
| `CB_DB_PATH`      | No       | `clipboard.db`       | Path to the SQLite database file   |
| `CB_ADDR`         | No       | `:8080`              | Listen address                     |
| `CB_CORS_ORIGIN`  | No       | `*`                  | Allowed CORS origin                |
| `CB_TLS_CERT`     | No       |                      | Path to TLS certificate (PEM)      |
| `CB_TLS_KEY`      | No       |                      | Path to TLS private key (PEM)      |
| `CB_TLS_AUTO`     | No       | `false`              | Auto-generate self-signed TLS cert |
| `CB_TLS_DIR`      | No       | `.`                  | Directory for auto-generated certs |
| `CB_ADMIN_SECRET`  | No       |                      | Admin API secret                   |

## Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v cb-data:/data \
  -e CB_JWT_SECRET="change-me" \
  -e CB_DB_PATH=/data/clipboard.db \
  -e CB_CORS_ORIGIN="https://yourdomain.com" \
  clipboard
```

## Fly.io

Reference `fly.toml` in the project root:

```bash
fly launch
fly secrets set CB_JWT_SECRET="change-me"
fly deploy
```

The included `fly.toml` handles health checks, volume mounts, and port configuration.

## TLS Setup

### Manual Certificates

```bash
export CB_TLS_CERT=/etc/ssl/certs/clipboard.pem
export CB_TLS_KEY=/etc/ssl/private/clipboard.key
```

### Auto-Generated Self-Signed Certs

Set `CB_TLS_AUTO=true`. The server generates a self-signed certificate and stores it in `CB_TLS_DIR`. Useful for development and internal deployments behind a reverse proxy.

```bash
export CB_TLS_AUTO=true
export CB_TLS_DIR=/data/certs
```

## Reverse Proxy (nginx)

```nginx
server {
    listen 443 ssl;
    server_name clipboard.example.com;

    ssl_certificate     /etc/letsencrypt/live/clipboard.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/clipboard.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /v1/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

In this setup, terminate TLS at nginx and run the server with plain HTTP (`CB_ADDR=:8080`).

## Production Checklist

- [ ] Set `CB_JWT_SECRET` to a strong random value (not a default or placeholder)
- [ ] Set `CB_CORS_ORIGIN` to your exact domain (not `*`)
- [ ] Enable TLS (either at the application level or via a reverse proxy)
- [ ] Set `CB_ADMIN_SECRET` if using admin endpoints
- [ ] Use a persistent volume for `CB_DB_PATH`
- [ ] Set `CB_DB_PATH` to an absolute path inside the volume
