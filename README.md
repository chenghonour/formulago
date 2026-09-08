# FormulaGo

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)
[![Hertz](https://img.shields.io/badge/Hertz-0.10.4-purple)](https://github.com/cloudwego/hertz)

[English](README_EN.md) | 中文

## 简介

**FormulaGo** 是一个高性能的企业后台管理框架，基于 [Hertz](https://github.com/cloudwego/hertz) 和 [Ent](https://entgo.io/) 构建。遵循 Clean Architecture / DDD 设计理念，开箱即用的 RBAC 权限管理系统。

> 前端项目地址：[formulago-front](https://github.com/chenghonour/formulago-front)

- **AI 原生 MCP Server** — 内置 MCP(Model Context Protocol)服务，Claude 等 AI 助手可直接对话操作后台数据，进入"AI 管理后台"时代。
- **高性能** — 使用字节跳动核心框架 Hertz，已在数万个服务上验证。
- **模块化架构** — 4 层清晰分层：Handler → Logic → Domain → Data。
- **类型安全 ORM** — Ent 生成类型安全的 Go 数据库操作代码。
- **RBAC 权限控制** — Casbin 实现 API 级细粒度权限管理。
- **多数据库** — 支持 MySQL 和 PostgreSQL，自动迁移。
- **OAuth 2.0 登录** — 内置 Google、GitHub、企业微信登录。
- **开箱即用** — 用户、角色、菜单、字典、Token、日志管理完整实现。

## 架构

```
biz/handler/  →  biz/logic/  →  biz/domain/  →  data/
(HTTP 层)        (业务逻辑)       (接口定义)     (数据访问)
```

## 技术栈

| 组件 | 库 |
|------|----|
| HTTP 框架 | [Hertz](https://github.com/cloudwego/hertz) |
| ORM | [Ent](https://entgo.io) |
| 认证 | [hertz-contrib/jwt](https://github.com/hertz-contrib/jwt) |
| 权限 | [Casbin](https://github.com/casbin/casbin) |
| MCP | [mcp-go](https://github.com/mark3labs/mcp-go) |
| 数据库 | MySQL / PostgreSQL |
| 缓存 | [go-cache](https://github.com/patrickmn/go-cache) + Redis |
| 配置 | [koanf](https://github.com/knadh/koanf) |
| 对象存储 | 阿里云 OSS |

## 快速开始

### 前置条件

- Go 1.26+
- MySQL（或 PostgreSQL）
- Docker（可选，用于本地 MySQL）

### 1. 配置

编辑 `configs/config.yaml` 配置文件。敏感信息建议通过环境变量设置，避免提交到 git：

```bash
# 复制环境变量模板
cp secret.sh.example secret.sh

# 编辑 secret.sh 填入真实密钥
# 然后加载到当前 shell
source secret.sh
```
`secret.sh` 已加入 `.gitignore`，不会误提交。


### 2. 启动 MySQL

```bash
docker-compose up -d
```

### 3. 运行

```bash
go build -o formulago && ./formulago
```

### 4. 初始化数据库

```bash
curl http://localhost:8191/api/initDatabase
```

默认管理员账号：`admin` / `formulago`

## 内置功能

| 模块 | 功能描述 |
|------|----------|
| 用户管理 | 系统用户增删改查，角色分配 |
| 角色管理 | RBAC 角色 CRUD，菜单和 API 权限分配 |
| 菜单管理 | 分层级菜单树，按钮级权限标识 |
| 字典管理 | 键值对字典维护，支持多级字典明细 |
| API 管理 | API 资源注册和管理，用于权限配置 |
| 操作日志 | 自动记录请求/响应日志，支持查询 |
| Token 管理 | 在线用户 Token 监控，强制下线 |
| 文件管理 | 文件上传，阿里云 OSS 适配，图片压缩 |
| OAuth 2.0 登录 | 支持 Google、GitHub、企业微信认证 |
| 验证码 | 数字验证码，支持配置长度和尺寸 |
| MCP Server | 向 Claude 等 AI 助手暴露管理能力，与 HTTP 同一套鉴权 |

## MCP Server（AI 助手接入）

FormulaGo 内置 MCP（Model Context Protocol）服务端 —— 把整个后台"变成"AI 可以调用的工具集，向 Claude 等 AI 原生应用打开一个受控的数据入口：

- **AI 直接对话管理数据** — 例如"帮我列出前 10 个用户"即可查询后台数据，无需教它 HTTP 接口。
- **浏览器 OAuth 授权** — 用户在 Claude 里粘贴一个地址，自动打开浏览器登录授权（PKCE 保护），无二次密码管理。
- **与 HTTP 同等的安全边界** — 复用同一套 JWT 会话与 Casbin 权限：网页上没权限的，AI 同样调不了；token 必须经 PKCE 授权流下发，杜绝"裸 token 直连"旁路。
- **部署零成本** — 挂在主 Hertz 服务下（`/api/mcp`），不另起端口。

```bash
# Claude Code 一行接入
claude mcp add --transport http formulago https://your-domain/api/mcp
```

## 项目结构

```
├── api/                  # Protobuf 定义和生成模型
│   ├── model/            # 生成的 Go 结构体
│   └── admin/            # 管理员 API Protobuf IDL
├── biz/                  # 业务逻辑
│   ├── domain/           # 纯 Go 接口定义（无外部依赖）
│   ├── handler/          # HTTP 处理层（参数绑定、响应）
│   │   └── middleware/   # JWT 认证、Casbin 权限、日志
│   ├── logic/            # 业务逻辑实现
│   ├── mcp/              # MCP Server（工具、OAuth/PKCE、鉴权）
│   └── router/           # 路由注册（含 mcp/ 的 Hertz 挂载）
├── configs/              # YAML 配置文件（嵌入二进制）
├── data/                 # 数据访问层
│   ├── ent/              # Ent ORM 生成代码
│   │   └── schema/       # Ent 数据模型定义
│   ├── s3/               # 阿里云 OSS 适配器
│   └── ...               # MySQL、PostgreSQL、Redis、Casbin
└── pkg/                  # 通用工具包
    ├── captcha/          # 验证码缓存
    ├── encrypt/          # bcrypt 密码加密
    ├── img/              # 图片压缩
    ├── types/            # 常用类型工具
    └── wecom/            # 企业微信集成
```

## 开发指南

### 代码生成

**Ent**（数据模型 → CRUD）：
```bash
go generate ./data/ent
```

**Hertz**（Protobuf → 处理器和路由）：
```bash
hz update -I api -idl api/admin/admin.proto -model_dir api/model --unset_omitempty
```
hz update -I api -idl api/admin/admin.proto -model_dir api/model --unset_omitempty
### 新增/编辑实体步骤
下面是admin管理模块相关的步骤，其他模块的步骤类似
1. 在 `data/ent/schema/` 创建/修改 schema
2. 运行 ent 代码生成
3. 在 `biz/domain/admin/` 定义领域接口
4. 在 `biz/logic/admin/` 实现业务逻辑
5. 在 `api/admin/admin.proto` Protobuf定义API接口
6. 运行 hz 生成处理器和路由
7. 在 `biz/handler/admin/` 实现处理器逻辑，注意把req\resp放一起初始化，让所有返回都使用上resp

## License

Apache License 2.0
