# FormulaGo

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)
[![Hertz](https://img.shields.io/badge/Hertz-0.10.4-purple)](https://github.com/cloudwego/hertz)

English | [中文](README.md)

## Introduction

**FormulaGo** is a high-performance backend admin framework built with [Hertz](https://github.com/cloudwego/hertz) and [Ent](https://entgo.io/). It follows Clean Architecture / DDD principles and provides a modern RBAC management system out of the box.

> Frontend repo: [formulago-front](https://github.com/chenghonour/formulago-front)

- **AI-Native MCP Server** — Built-in Model Context Protocol (MCP) server lets Claude and other AI agents operate your backend data through natural conversation — the "AI-managed admin" era.
- **High Performance** — Powered by Hertz, the most performant Go HTTP framework used by ByteDance across tens of thousands of services.
- **Modular Architecture** — Clean 4-layer design: Handler → Logic → Domain → Data.
- **Type-Safe ORM** — Ent generates type-safe Go code for all database operations.
- **RBAC Authorization** — Casbin enforces fine-grained API-level access control.
- **Multi-Database** — Supports MySQL and PostgreSQL with auto-migration.
- **OAuth 2.0** — Built-in support for Google, GitHub, and WeCom (WeChat Work) login.
- **Plug & Play** — User, role, menu, dictionary, token, and log management included.

## Architecture

```
biz/handler/  →  biz/logic/  →  biz/domain/  →  data/
(HTTP layer)    (business)     (interfaces)    (data access)
```

## Tech Stack

| Component | Library |
|-----------|---------|
| HTTP Framework | [Hertz](https://github.com/cloudwego/hertz) |
| ORM | [Ent](https://entgo.io) |
| Auth | [hertz-contrib/jwt](https://github.com/hertz-contrib/jwt) |
| RBAC | [Casbin](https://github.com/casbin/casbin) |
| MCP | [mcp-go](https://github.com/mark3labs/mcp-go) |
| Database | MySQL / PostgreSQL |
| Cache | [go-cache](https://github.com/patrickmn/go-cache) + Redis |
| Config | [koanf](https://github.com/knadh/koanf) |
| Object Storage | Aliyun OSS |

## Quick Start

### Prerequisites

- Go 1.26+
- MySQL (or PostgreSQL)
- Docker (optional, for local MySQL)

### 1. Configure

Edit `configs/config.yaml` with your database credentials. Use environment variables for secrets to avoid committing them to git:

```bash
# Copy the env template
cp secret.sh.example secret.sh

# Edit secret.sh with your real credentials
# Then load into your shell
source secret.sh
```
`secret.sh` is in `.gitignore` and won't be committed. 

### 2. Start MySQL

```bash
docker-compose up -d
```

### 3. Run

```bash
go build -o formulago && ./formulago
```

### 4. Initialize Database

```bash
curl http://localhost:8191/api/initDatabase
```

Default admin account: `admin` / `formulago`

## Built-in Features

| Module | Description |
|--------|-------------|
| User Management | Create, update, query, delete users with role assignment |
| Role Management | RBAC role CRUD with menu and API permission assignment |
| Menu Management | Hierarchical menu tree with button-level permissions |
| Dictionary | Key-value dictionary management with multi-level details |
| API Management | Register and manage API resources for authorization |
| Operation Logs | Automatic request/response logging with query |
| Token Management | Active user token monitoring and forced logout |
| File Management | File upload with Aliyun OSS adapter and image compression |
| OAuth 2.0 Login | Google, GitHub, and WeCom authentication |
| Captcha | Digit captcha with configurable length and size |
| MCP Server | Exposes admin capabilities to AI agents with the same auth as HTTP |

## MCP Server (AI Agent Access)

FormulaGo ships a built-in MCP (Model Context Protocol) server — it turns the whole admin backend into a toolset AI agents can call, opening a controlled data entry point for AI-native applications:

- **Talk to your data** — e.g. "list the first 10 users" queries the backend directly; no need to teach the agent raw HTTP endpoints.
- **Browser OAuth authorization** — users paste one URL into Claude and the browser opens automatically for login & consent (PKCE protected); no second set of credentials.
- **Same security boundary as HTTP** — reuses the very same JWT sessions and Casbin policies: what a role cannot do on the web, the AI cannot do either. Tokens are only issued through the PKCE authorization flow — no "raw token bypass".
- **Zero extra deployment** — mounted under the existing Hertz service at `/api/mcp`, no separate port. 

```bash
# One-line Claude Code integration
claude mcp add --transport http formulago https://your-domain/api/mcp
```

## Project Structure

```
├── api/                  # Protobuf definitions & generated models
│   ├── model/            # Generated Go protobuf structs
│   └── admin/            # Admin API protobuf IDL
├── biz/                  # Business logic
│   ├── domain/           # Pure Go interfaces (no external deps)
│   ├── handler/          # HTTP handlers (request binding, response)
│   │   └── middleware/   # JWT auth, Casbin RBAC, logging
│   ├── logic/            # Business logic implementation
│   ├── mcp/              # MCP server (tools, OAuth/PKCE, auth)
│   └── router/           # Route registration (incl. mcp/ Hertz mount)
├── configs/              # YAML configuration (embedded)
├── data/                 # Data access layer
│   ├── ent/              # Generated Ent ORM code
│   │   └── schema/       # Ent schema definitions
│   ├── s3/               # Aliyun OSS adapter
│   └── ...               # MySQL, PostgreSQL, Redis, Casbin
└── pkg/                  # Shared utilities
    ├── captcha/          # Captcha cache store
    ├── encrypt/          # bcrypt password hashing
    ├── img/              # Image compression
    ├── types/            # Common type helpers
    └── wecom/            # WeCom (WeChat Work) integration
```

## Development

### Code Generation

**Ent** (schema → CRUD):
```bash
go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/versioned-migration ./data/ent/schema
```

**Hertz** (protobuf → handlers/routes):
```bash
hz update -I api -idl api/admin/admin.proto -model_dir api/model --unset_omitempty
```

### Adding / Editing an Entity
The following steps are for the admin module; other modules follow the same pattern.

1. Create or modify schema in `data/ent/schema/`
2. Run ent code generation
3. Define or update domain interface in `biz/domain/admin/`
4. Implement or update logic in `biz/logic/admin/`
5. Define the API in `api/admin/admin.proto` using Protobuf
6. Run hz to generate handlers and routes
7. Implement handler logic in `biz/handler/admin/` — initialize req/resp together and ensure all returns use `resp`

## License

Apache License 2.0
