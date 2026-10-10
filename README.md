# Pricebase

Pricebase is the pricing-platform API: Postgres-backed RBAC (roles, permissions, users), JWT + API-key auth, and realtime updates over Socket.IO. Built with Gin, pgx, and Redis in Go.

## Stack


| Concern     | Choice                                                              |
| ----------- | ------------------------------------------------------------------- |
| HTTP        | Gin (+ CORS, security headers, rate limit, sanitize, Socket.IO)     |
| Database    | Postgres via pgx (`pgxpool`), hand-written SQL, embedded migrations |
| Realtime    | Socket.IO server mounted at `/socket/` (WebSocket transport)        |
| Cache/jobs  | Redis (`go-redis` universal client, standalone or cluster)          |
| Email       | Resend                                                              |
| Logging     | zap (rotating files) + remote shipping; console helper in `utils`   |
| Shared code | `github.com/damolaoladipo/common-go` (local checkout, see below)    |




## Prerequisites

- Go 1.26+
- Postgres 13+ (uses `gen_random_uuid()`)
- Redis (standalone or cluster)
- A local checkout of `common-go` at `../common-go` — `go.mod` contains `replace github.com/damolaoladipo/common-go => ../common-go`, so the sibling directory must exist at build time



## Quickstart

```sh
# 1. Dependencies
go mod tidy

# 2. Configure (see Configuration)
cp .env.example .env   # then fill in real secrets; never commit .env

# 3. Build the binary
npm run build

# 4. Start with PM2 (connects → migrates → seeds → connects Redis → serves)
npm run dev       # development (go run + watch mode)
npm run prod      # production (compiled binary)
```

The API serves on `:$APP_PORT` (default `5025`), versioned routes under `/v1`, Socket.IO at `/socket/`.

### PM2 scripts

| Command           | Action                                              |
| ----------------- | --------------------------------------------------- |
| `npm run dev`     | Start dev process (`go run ./cmd/pricebase`, watch) |
| `npm run build`   | Compile binary to `./bin/pricebase-api`             |
| `npm run prod`    | Start production process (compiled binary)          |
| `npm run restart` | Restart the production process                      |
| `npm run stop`    | Stop the production process                         |
| `npm run status`  | Show PM2 process status                             |
| `npm run logs`    | Tail production logs                                |
| `npm run monitor` | Open PM2 monitor                                   |
| `npm run save`    | Save PM2 process list for resurrect                 |

## Configuration (`.env`)


| Key                                                                  | Default                                               | Used for                                              |
| -------------------------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `APP_ENV`                                                            | `development`                                         | `dev`/`staging`/`production` behavior switches        |
| `APP_PORT`                                                           | `5025`                                                | HTTP listen port                                      |
| `APP_NAME`                                                           | `Pricebase`                                           | Log/service labels                                    |
| `POSTGRES_URL`                                                       | `postgres://localhost:5432/pricebase?sslmode=disable` | App database                                          |
| `REDIS_URL`                                                          | `redis://localhost:6379`                              | Redis (`rediss://` enables TLS)                       |
| `REDIS_CLUSTER`                                                      | `false`                                               | `true` for cluster mode                               |
| `JWT_SECRET`                                                         | `secret`                                              | JWT HMAC key (HS256, hand-rolled, no dependency)      |
| `JWT_EXPIRES`                                                        | `1h`                                                  | Token TTL (Go duration)                               |
| `SUPERADMIN_EMAIL`                                                   | *(empty)*                                             | Account the attach step boosts to `user`+`superadmin` |
| `EMAIL_ENABLED` / `RESEND_API_KEY` / `EMAIL_FROM` / `EMAIL_REPLY_TO` |                                                       | Outbound email                                        |
| `LOGS_PATH`                                                          | `internal/_data/logs`                                | Log sink directory                                    |




## Project structure

```text
pricebase/
├── .docs/
│   └── publish.md          # Release and deployment flow
├── cmd/
│   └── pricebase/
│       └── main.go         # Process entry (package main)
├── internal/
│   ├── common/             # Logger, redis, email, httpclient (not for external import)
│   ├── config/             # App/server/DB wiring, migrations runner, seeds/
│   ├── dtos/               # Request/response shapes
│   ├── factory/            # Composition root (repositories → services)
│   ├── handlers/           # Plain-function Gin handlers
│   ├── middlewares/        # Recovery, errors, CORS, security, search, …
│   ├── models/             # Domain structs + pure methods
│   ├── repository/         # All Postgres access (queries live in utils)
│   ├── routes/             # /v1 route groups
│   ├── services/           # Orchestration over repositories
│   ├── utils/              # Leaf package: SQL strings, enums, ENV, IDs
│   └── _data/              # Seed JSON + runtime logs under logs/ (see LOGS_PATH)
├── migrations/             # Embedded *.sql, applied in version order at boot
├── tests/
│   ├── unit/
│   └── integration/
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── README.md
└── CONTRIBUTING.md
```



## Architecture & conventions

- **Leaf-down imports:** `utils` imports nothing internal. `models` may import `utils`. `dtos` may import `utils`. Nothing internal imports `dtos` from below (a past import cycle taught us this — keep it that way).
- **SQL lives in** `internal/utils/queries.util.go`**; DB access lives in** `internal/repository/`**.** Services orchestrate repositories; seeds orchestrate repositories; handlers call services via the `factory` globals. Raw `db.Exec`/`QueryRow` outside `repository/` is a bug.
- **Handlers are plain functions** (`func Register(ctx *gin.Context)`), wired in `routes/` with no structs.
- **Factory is the composition root** (`factory.InitServices(db)`), called once in `main.go` *after* seeds. Seeds build their own repositories from the pool — they must not depend on factory globals (bootstrap runs before the app graph exists).
- **IDs** are app-generated 24-char hex (`utils.GenerateID`), stored as `TEXT`. Repositories default empty IDs on create; migrations convert legacy UUIDs (see `000006`).
- **RBAC snapshot:** each user row carries `role_ids` + `permissions` JSONB, recomputed by the attach seeder every boot. The `user_roles` / `role_permissions` joins remain the relational record. If you ever build a runtime role editor, it must recompute the snapshot.
- **JSON** goes through `common-go/utils` (`ToJSON`, `FromJSON`, `ToJSONArray`) — no `encoding/json` in app code except the models JWT builder.
- **Enums** in `utils/enums.util.go` are real enums: `String()`, `IsValid()`, `<Type>Values()`. Validate at system boundaries.
- **Explicit over abstract:** shared private helpers are removed on sight when only 2–3 call sites exist — duplicate the five lines instead.



## Boot order (`cmd/pricebase/main.go`)

connect Postgres → run migrations → run seeds (roles → permissions → role-permissions → users → attach superadmin) → connect Redis → `factory.InitFactories(db)` → serve (Gin + Socket.IO, graceful shutdown on SIGINT/SIGTERM).

## Migrations

Hand-rolled runner (`internal/config/migration.config.go`): embeds `migrations/*.sql`, tracks applied versions in a `migrations` table, runs each file once inside a transaction. To add one, create `migrations/NNNNNN_name.up.sql` — never edit an applied file, and never add `*.down.sql` files (the runner executes every `*.sql` file it finds).

## Seeds

JSON fixtures live in `internal/_data/` (`roles.json`, `permissions.json`, `role-permissions.json`, `users.json`). Seeders are idempotent — re-running changes nothing (`ON CONFLICT DO NOTHING` + existence checks). Set `SUPERADMIN_EMAIL` to enable the attach step (roles + encrypted password + API key for that account).

## Testing

```sh
go test ./...                        # unit always runs; integration skips without Postgres
go test ./tests/unit/ -v
POSTGRES_URL="postgres://pro@/pricebase_test?host=/tmp/pg&port=55433&sslmode=disable" \
  go test ./tests/integration/ -v
```

Gates before pushing: `gofmt -l` clean on touched files, `go build ./...`, `go vet ./...`, `go test ./...`.

## Releasing

Service releases are semver Git tags (`vMAJOR.MINOR.PATCH`) marking deployable artifacts — never library publishes. The full flow (pre-release verification, migrations/seed scrutiny, tagging, deploy config, checklist) lives in `[.docs/publish.md](.docs/publish.md)`. Read it before cutting any release.

## License

AGPL-3.0 — see [LICENSE](LICENSE). Made by Damola.