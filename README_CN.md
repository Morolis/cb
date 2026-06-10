<h1 align="center">cb</h1>

<p align="center">跨设备剪贴板 & 代码片段同步</p>

<p align="center">一款轻量级 CLI + Web 工具，帮助开发者在多台设备之间同步短文本、代码片段和命令 —— 端到端加密、本地优先存储、实时同步。</p>

<p>
  <strong><a href="README.md">English</a> | 中文</strong>
</p>

## 为什么用 cb？

**SSH 到线上服务器，需要本地笔记本上的文件。**
不用切微信找聊天记录、不用开笔记 App、不用手打。`cb get deploy-key` — 搞定。

**公司写了个复杂的 SQL，想回家继续。**
不用发邮件给自己、不用粘贴到备忘录、不用 AirDrop。公司：`cb stash my-query "SELECT ..."`，家里：`cb get my-query`。

**同事急着要你的 nginx 配置。**（TODO: 定向分享 & 团队空间开发中）
不用"等我找一下"然后挂电话去翻文件。`cb stash nginx-conf < nginx.conf` — 他运行 `cb get nginx-conf`。

**要发个 API 密钥，但不想它永远留在聊天记录里。**
`cb send --encrypt --ttl 30m "sk-xxxx"` — 加密传输，30 分钟后自动销毁，服务端看不到明文。

**管 5 台服务器，每台的部署命令都不一样。**
`cb save deploy-prod --category ops "kubectl apply -f prod.yaml"` — 按分类整理，加描述，打标签。`cb list --category ops` 一目了然。

## 快速开始

```bash
# 安装
go install github.com/Morolis/cb@latest

# 登录（没有账号会自动注册）
cb login --user me --api-url http://your-server:8080/v1

# 发送文本到云端（跨设备同步）
cb send "hello from my laptop"

# 快速保存到云端并命名
cb stash deploy "kubectl apply -f prod.yaml"

# 在另一台机器上获取
cb get deploy

# 直接执行
cb exec deploy
```

## 安装

### Go install

```bash
go install github.com/Morolis/cb@latest
```

### 下载二进制

从 [GitHub Releases](https://github.com/Morolis/cb/releases) 下载：
- Linux: `cb-linux-amd64.tar.gz`、`.deb`
- macOS: `cb-darwin-amd64.tar.gz`、`cb-darwin-arm64.tar.gz`
- Windows: `CBSetup-amd64.exe`（安装程序）或 `cb-windows-amd64.zip`

### Homebrew (macOS)

```bash
brew tap Morolis/cb
brew install cb
```

### Docker（服务端）

```bash
# 预构建镜像（推荐，支持 amd64 和 arm64）
docker pull ghcr.io/morolis/cb:latest
docker run -d -p 8080:8080 -v cb-data:/data --name cb --restart always ghcr.io/morolis/cb:latest

# 或从源码构建
git clone https://github.com/Morolis/cb.git && cd cb
docker build -t cb-server .
docker run -d -p 8080:8080 -v cb-data:/data --name cb cb-server
```

## 命令

### `send` — 云端剪贴板

发送文本到云端，跨设备同步。

```bash
cb send "kubectl get pods -A"
cb send --alias mylink "https://example.com"
cb send --ttl 1h --encrypt "secret data"
cb send --id 597ebc3e "updated content"       # 更新已有片段（产生版本历史）
cb send --alias myconfig "v2"                  # 有则更新，无则创建
cat config.yaml | cb send --alias myconfig
```

| 参数 | 说明 |
|------|------|
| `--alias` | 设置别名（有则更新，无则创建） |
| `--id` | 通过 ID 或前缀更新已有片段 |
| `--ttl` | 存活时间：`30s`、`5m`、`1h`、`1d` |
| `--encrypt` | AES-256-GCM 加密 |
| `--desc` | 描述 |
| `--var KEY=VALUE` | 变量替换：替换内容中的 `{{.KEY}}` |

### `save` — 本地优先存储

保存到本地，离线可用。加 `--remote` 同步到云端。

```bash
cb save mycmd "kubectl get pods -A"
cb save --desc "查看所有 Pod" --category k8s --lang bash mycmd "kubectl get pods -A"
cb save --remote mycmd "ls -la"                # 同时同步到云端
cb save --ttl 7d --encrypt mydb "postgresql://user:pass@host/db"
```

| 参数 | 说明 |
|------|------|
| `--remote` | 同时推送到远程服务器 |
| `--ttl` | 存活时间 |
| `--encrypt` | AES-256-GCM 加密 |
| `--desc` | 描述 |
| `--category` | 分类 |
| `--lang` | 语言提示（如 `python`、`bash`） |
| `--tags` | 逗号分隔的标签 |

### `stash` — 快速云端保存

快捷方式：带别名保存到云端。

```bash
cb stash deploy "kubectl apply -f deploy.yaml"
cb stash --desc "数据库备份" db-backup "pg_dump mydb > backup.sql"
```

| 参数 | 说明 |
|------|------|
| `--ttl` | 存活时间 |
| `--encrypt` | AES-256-GCM 加密 |
| `--desc` | 描述 |

### `get` — 获取片段

通过 ID 或别名获取片段。先查本地，再查远程。无参数返回最近的片段。

```bash
cb get mycmd              # 通过别名
cb get 597ebc3e           # 通过 ID（或前缀）
cb get                    # 最近的片段
```

### `list` — 列出所有片段

本地 + 云端合并视图。

```bash
cb list
cb list --source local     # 仅本地
cb list --source remote    # 仅云端
cb list --limit 50
```

输出：
```
SOURCE  ALIAS  DESC  ID        PREVIEW                     CREATED           EXPIRES
------  -----  ----  --        -------                     -------           -------
local   mycmd  -     loc_c8fd  kubectl get pods -A         2024-01-01 10:30  -
remote  -      -     597ebc3e  hello world                 2024-01-01 10:25  -
```

### `exec` — 执行片段

将保存的片段作为 shell 命令执行。

```bash
cb exec mycmd             # 通过别名执行
cb exec 597ebc3e          # 通过 ID 执行
```

### `rm` — 删除片段

通过 ID 或别名删除。自动判断本地/远程。

```bash
cb rm mycmd               # 自动判断
cb rm loc_c8fdf403a07d    # 本地 ID
cb rm --source local mycmd
cb rm --source remote mycmd
```

### `history` — 版本历史

查看片段的所有历史版本。

```bash
cb history mycmd
cb history 597ebc3e
```

### `rollback` — 回滚版本

回滚到指定版本。当前内容会自动保存为新版本。

```bash
cb rollback mycmd 3           # 回滚到版本 ID 3
cb rollback 597ebc3e 7
```

### `webhook` — 管理 Webhook

片段变更时接收 HTTP POST 通知。

```bash
# 添加 webhook（默认 JSON 格式）
cb webhook add myhook https://example.com/hook created,updated,deleted

# 添加带自定义模板的 webhook
cb webhook add slack https://hooks.slack.com/xxx created \
  --body '{"text":"[{{.Event}}] {{.Snippet.Content}}"}'

# 列出所有 webhook
cb webhook list

# 查看投递日志
cb webhook logs <webhook-id>

# 删除
cb webhook rm <webhook-id>
```

模板变量：`{{.Event}}`、`{{.DateTime}}`、`{{.Snippet.ID}}`、`{{.Snippet.Alias}}`、`{{.Snippet.Content}}`、`{{.Snippet.Description}}`、`{{.Snippet.Category}}`、`{{.Snippet.Language}}`、`{{.Snippet.Encrypted}}`、`{{.Snippet.ExpiresAt}}`、`{{.Snippet.CreatedAt}}`、`{{.Snippet.UpdatedAt}}`。

使用 `{{json .Snippet.Content}}` 安全转义内容（自动处理换行、引号等）。

### `login` / `logout` — 认证

```bash
cb login --user myname --api-url http://server:8080/v1
cb logout
```

### `config` — 查看/修改配置

```bash
cb config show
cb config set api_url http://your-server:8080/v1
```

配置文件：`~/.cb/config.yaml`

### 全局参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--config` | `~/.cb/config.yaml` | 配置文件路径 |
| `--api-url` | `http://localhost:8080/v1` | 覆盖服务端 API 地址 |
| `--verbose` | `false` | 输出调试信息 |
| `-v, --version` | — | 显示版本 |

## 功能特性

| 功能 | 说明 |
|------|------|
| **三种模式** | `send`（云端）、`save`（本地优先）、`stash`（快速云端别名） |
| **版本历史** | 每次更新自动保存旧版本 |
| **端到端加密** | AES-256-GCM，服务端看不到明文 |
| **实时同步** | WebSocket 推送，所有已连接设备实时更新 |
| **Web UI** | 控制台：语法高亮、编辑器、版本历史、Webhook 管理 |
| **命令执行** | `cb exec mycmd` 直接运行保存的片段 |
| **自动过期** | `--ttl 30m` / `1h` / `1d` — 片段自动销毁 |
| **Webhook** | 片段变更时通知飞书、钉钉、企业微信、Slack 或任意 HTTP 端点 |
| **组织管理** | 分类、标签、描述、别名 |
| **离线优先** | 本地 SQLite 缓存，无网络也能用 |
| **管道支持** | `cat file \| cb send`，`cb get deploy \| sh` |

## 服务端

```bash
# 启动服务端
CB_JWT_SECRET="your-secret" cb-server

# 启用 TLS
CB_TLS_CERT=cert.pem CB_TLS_KEY=key.pem cb-server

# 自动生成自签证书
CB_TLS_AUTO=true cb-server
```

打开 `http://localhost:8080` 访问 Web UI。

## 文档

- [API 接口](docs/api.md)
- [加密设计](docs/encryption.md)
- [部署指南](docs/deploy.md)

## 参与贡献

```bash
git clone https://github.com/Morolis/cb.git
cd cb

# CLI
go build -o cb .

# 服务端（内嵌前端）
cd web && npm install && npm run build && cd ..
go build -o cb-server ./server/main.go

# 前端开发
cd web && npm run dev
```

## 开源协议

[Apache License 2.0](LICENSE)
