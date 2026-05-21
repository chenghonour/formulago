# FormulaGo CLAUDE.md

## Build & Test

```bash
go build ./...
go test ./...
go vet ./...
go mod tidy
```

## Project Architecture (4-Layer DDD)

```
biz/handler/  →  biz/logic/  →  biz/domain/  →  data/
(HTTP layer)    (business)     (interfaces)    (data access)
```

- **`biz/handler/`** — HTTP handlers: bind request, call logic, format response
- **`biz/logic/`** — Business logic: implements domain interfaces, orchestrates data
- **`biz/domain/`** — Pure Go interfaces + model structs (no external dependencies)
- **`data/`** — Data layer: ent client, Redis, Casbin, S3, caching
- **`pkg/`** — Shared utility packages

### Adding / Editing an Entity (Admin Module as example)
1. Create or modify schema in `data/ent/schema/`
2. Run ent code generation
3. Define or update domain interface in `biz/domain/admin/`
4. Implement or update logic in `biz/logic/admin/`
5. Define the API in `api/admin/admin.proto` using Protobuf
6. Run hz to generate handlers and routes
7. Implement handler logic in `biz/handler/admin/` — initialize req/resp together, ensure all returns use `resp`

## Code Generation

### ent (schema → CRUD)
Run from project root:
```bash
go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/versioned-migration ./data/ent/schema
```

### hz (protobuf → handlers/routes)
```bash
hz update -I api -idl api/admin/admin.proto -model_dir api/model --unset_omitempty
```

## Code Conventions

### Error Handling
- Use `fmt.Errorf("msg: %w", err)` — **not** `github.com/pkg/errors`
- Never swallow errors — every `err` must be at least logged if not returned
- Database `Count()` errors must be returned, never discarded with `_`
- Always propagate `context.Context` — no `context.Background()` in biz logic

### Security
- Secrets go in **environment variables**, not config files
- Use `crypto/rand` for cryptographic randomness (never `math/rand`)
- No `WithInsecureSkipVerify(true)` in production code
- Validate authorized user matches resource owner on mutating operations

### Configuration
- YAML configs are embedded via `//go:embed` — only non-sensitive defaults
- All sensitive fields (passwords, keys, tokens) have env-var override in `overrideSecretFromEnv()`
- Config env vars: `DB_PASSWORD`, `ACCESS_SECRET`, `OAUTH_KEY`, `OSS_SECRET_ID`, `OSS_SECRET_KEY`, `WECOM_CORP_ID`, `WECOM_SECRET_ID`, `WECOM_TOKEN`, `WECOM_AES_KEY`, `REDIS_PASSWORD`

### Testing
- Unit tests in `pkg/*_test.go` (no external dependencies)
- Handler tests in `biz/handler/*_test.go` (may need DB/Redis)
- All exported functions must have tests for both success and error paths

### General
- Package naming: lowercase, single word, no underscores
- Time format: use `pkg/times/format.go  `TimeFormat` constant instead of `"2006-01-02 15:04:05"`
- Comments only for non-obvious WHY — don't explain what the code does

