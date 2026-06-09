# API Reference

All responses are JSON. Protected endpoints require the header:

```
Authorization: Bearer <jwt>
```

Tokens are obtained from the auth endpoints below.

## Auth

### Register

```
POST /v1/auth/register
```

**Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "token": "string",
  "user_id": 1,
  "username": "string"
}
```

### Login

```
POST /v1/auth/login
```

**Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "token": "string",
  "user_id": 1,
  "username": "string"
}
```

## Snippets

All snippet endpoints are protected.

### Create Snippet

```
POST /v1/snippets
```

**Body:**
```json
{
  "content": "string (required)",
  "alias": "string (optional)",
  "description": "string (optional)",
  "ttl": "duration string, e.g. '24h' (optional)",
  "encrypted": false,
  "category": "string (optional)",
  "language": "string (optional)",
  "tags": ["string"]
}
```

### List Snippets

```
GET /v1/snippets?limit=20&offset=0&category=&tag=
```

| Param      | Type   | Default | Description            |
|------------|--------|---------|------------------------|
| `limit`    | int    | 20      | Max results            |
| `offset`   | int    | 0       | Pagination offset      |
| `category` | string |         | Filter by category     |
| `tag`      | string |         | Filter by tag          |

### Get Snippet by ID

```
GET /v1/snippets/{id}
```

### Update Snippet

```
PUT /v1/snippets/{id}
```

Auto-saves the previous version. Same body fields as create.

### Delete Snippet

```
DELETE /v1/snippets/{id}
```

### Version History

```
GET /v1/snippets/{id}/versions
```

### Rollback

```
POST /v1/snippets/{id}/rollback
```

**Body:**
```json
{
  "version_id": 3
}
```

### Get by Alias

```
GET /v1/snippets/alias/{alias}
```

### Get by ID Prefix

```
GET /v1/snippets/prefix/{prefix}
```

Returns the snippet whose ID starts with the given prefix.

## Webhooks

All webhook endpoints are protected.

### Create Webhook

```
POST /v1/webhooks
```

**Body:**
```json
{
  "name": "string",
  "url": "https://example.com/hook",
  "events": ["snippet_created", "snippet_deleted"],
  "secret": "string (optional)",
  "body_template": "string (optional)"
}
```

### List Webhooks

```
GET /v1/webhooks
```

### Delete Webhook

```
DELETE /v1/webhooks/{id}
```

### Toggle Webhook

```
PUT /v1/webhooks/{id}/toggle
```

Enable or disable the webhook.

### Delivery Logs

```
GET /v1/webhooks/{id}/logs
```

## Devices

### Heartbeat

```
POST /v1/devices/heartbeat
```

**Body:**
```json
{
  "name": "string",
  "type": "string"
}
```

### List Online Devices

```
GET /v1/devices
```

## WebSocket

```
GET /v1/ws?token=<jwt>
```

Upgrade to WebSocket. Events pushed over the connection:

| Event              | Description               |
|--------------------|---------------------------|
| `snippet_created`  | New snippet was created   |
| `snippet_deleted`  | Snippet was deleted       |
| `snippet_updated`  | Snippet was updated       |

## Health

```
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

No authentication required.
