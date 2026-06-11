# Deployment Guide

## Architecture

```
Internet → cb Server (HTTP :8080, optional HTTPS :443)
```

cb runs as an HTTP server by default. You can enable HTTPS directly (self-signed certificate) or use a reverse proxy (recommended for production with a domain).

---

## Standalone (Direct Binary)

### Install

```bash
# Debian/Ubuntu
sudo dpkg -i cb_0.1.0_amd64.deb

# Or download binary from GitHub Releases and add to PATH
```

### Run

```bash
cb-server
```

The server starts on HTTP :8080. Open `http://localhost:8080` in your browser.

### Enable HTTPS (Self-Signed)

1. Open the web UI → Settings → Server Config
2. Click "Enable HTTPS"
3. The server generates a self-signed certificate in `~/.cb/certs/` and starts HTTPS on :443
4. HTTP requests are automatically redirected to HTTPS

```bash
# Or enable via environment variable
CB_TLS_AUTO=true cb-server
```

### Run as systemd Service

```bash
sudo tee /etc/systemd/system/cb.service << 'EOF'
[Unit]
Description=cb server
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/cb-server
WorkingDirectory=/var/lib/cb
Environment=CB_DB_PATH=/var/lib/cb/cb.db
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo mkdir -p /var/lib/cb
sudo systemctl enable cb
sudo systemctl start cb
```

---

## Docker

### HTTP only

```bash
docker run -d \
  -p 8080:8080 \
  -v cb-data:/data \
  --name cb \
  --restart always \
  ghcr.io/morolis/cb:latest
```

### With HTTPS support (enable from web UI when ready)

```bash
docker run -d \
  -p 8080:8080 \
  -p 443:443 \
  -v cb-data:/data \
  --name cb \
  --restart always \
  ghcr.io/morolis/cb:latest
```

Then open Settings → Server Config → click "Enable HTTPS".

### With reverse proxy (recommended for production)

```bash
docker run -d \
  -p 8080:8080 \
  -v cb-data:/data \
  --name cb \
  --restart always \
  ghcr.io/morolis/cb:latest
```

Then configure Nginx or Caddy on the host (see below).

---

## Reverse Proxy: Caddy (Recommended)

Caddy automatically obtains and renews Let's Encrypt certificates. Requires a domain.

```bash
# Install Caddy (Ubuntu)
sudo snap install caddy

# Or Debian
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update && apt install caddy
```

**/etc/caddy/Caddyfile:**
```
cb.yourdomain.com {
    reverse_proxy localhost:8080
}
```

```bash
sudo systemctl restart caddy
```

Caddy handles TLS, WebSocket, HTTP/2 automatically. No extra configuration needed.

---

## Reverse Proxy: Nginx

**/etc/nginx/conf.d/cb.conf:**
```nginx
server {
    listen 80;
    server_name cb.yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name cb.yourdomain.com;

    ssl_certificate     /etc/letsencrypt/live/cb.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/cb.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket (required for real-time sync)
    location /v1/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }
}
```

Get a certificate:
```bash
apt install certbot
certbot certonly --nginx -d cb.yourdomain.com
nginx -t && systemctl reload nginx
```

---

## Environment Variables

| Variable          | Default          | Description                                    |
|-------------------|------------------|------------------------------------------------|
| `CB_JWT_SECRET`   | auto-generated   | JWT signing secret (auto-generated if empty)    |
| `CB_DB_PATH`      | `cb.db`          | SQLite database path                           |
| `CB_ADDR`         | `:8080`          | HTTP listen address                            |
| `CB_CORS_ORIGIN`  | `*`              | Allowed CORS origin (set to your domain)       |
| `CB_TLS_DIR`      | `~/.cb/certs` or `/data/certs` | Directory for TLS certificates   |
| `CB_TLS_CERT`     |                  | Path to TLS certificate (overrides auto-detect) |
| `CB_TLS_KEY`      |                  | Path to TLS private key (overrides auto-detect) |
| `CB_TLS_AUTO`     | `false`          | Auto-generate self-signed cert on startup      |

---

## Production Checklist

- [ ] Set `CB_CORS_ORIGIN` to your exact domain (not `*`)
- [ ] Enable HTTPS (via web UI or reverse proxy)
- [ ] Use a persistent volume for `CB_DB_PATH`
- [ ] Set `--restart always` (Docker) or `Restart=always` (systemd)
- [ ] Open port 443 in firewall if using built-in HTTPS
