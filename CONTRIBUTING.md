# Contributing to Pricebase

For architecture, env vars, and run commands, see README.md. Do not commit `.env` or secrets. Keep layer boundaries clear (see below).

## Branches

| Branch              | Purpose                                                    |
|---------------------|------------------------------------------------------------|
| `master`            | Production-ready code. Always stable. Protected.           |
| `staging`           | QA/testing branch integrating features before release.     |
| `release/vX.Y.Z`    | Pre-production branch for final testing before going live. |
| `@username/feature-*` | Feature branches under a personal namespace.             |
| `@username/fix-*`     | Bug-fix branches under a personal namespace.             |

Naming: lowercase, hyphens, short descriptions — e.g. `@damolaoladipo/feature-identity-login`, `@topeokuselu/fix-jwt-expire`, `release/v1.0.2`.

## Workflow

1. Confirm the slice of work and which layer it touches (`queries → repository → services → handlers/routes`, plus `models`/`dtos` as needed). Open a GitHub Issue for bugs/proposals, or escalate to your team lead for triage.
2. Create a branch from `staging`:
   ```sh
   git checkout staging && git pull origin staging
   git checkout -b @username/feature-your-task-name
   ```
   Ensure the `../common-go` sibling checkout exists, copy `.env.example` → `.env`, then `go mod tidy`.
3. Check for concurrent edits first — several contributors work in these files at once. Read the exact current text before editing; exact-match edits fail loudly on drift, which is your signal to re-read.
4. Keep diffs minimal and in-pattern (see Conventions). Don't reformat untouched files. If you add a required env key, update `.env.example` in the same diff.
5. Sync before pushing: `git fetch origin && git rebase origin/staging`, then push your branch.
6. Gates before pushing:
   ```sh
   gofmt -l <touched files>   # must print nothing
   go build ./...
   go vet ./...
   go test ./...
   ```
   Useful commands:
    ```sh
    npm run dev                   # start the API in dev mode (PM2 + watch)
    npm run build                 # compile binary to ./bin/pricebase-api
    npm run prod                  # start production process
    POSTGRES_URL="..." go test ./tests/integration/ -v
    ```
7. Open a PR into `staging` (never `master`): reference issues (`Closes #502`), request reviewers, and call out migration / seed / env notes plus CORS, port, or contract changes affecting clients. Document breaking API changes.
8. Release: cut `release/vX.Y.Z` from `staging` for final testing, then merge it into both `master` and `staging`. Tagging, versioning, and deploy checks live in `.docs/publish.md` — follow it for anything reaching production.
9. For anything touching Postgres, verify live on a scratch database (initdb + `go run` a throwaway harness or the new `tests/integration` cases), then delete the scratch artifacts. Never verify against shared/dev databases.

## Layer conventions

- **Queries** live only in `internal/utils/queries.util.go` as `VAR_NAME` constants. Parameterized `$n` always — never interpolate values (the file header documents why).
- **Repositories** (`internal/repository/`) own every `db.Exec`/`QueryRow`/`Query`. Wrap errors with `fmt.Errorf("...: %w", err)` naming the operation; return `pgx.ErrNoRows` unchanged on miss so callers can branch; check `RowsAffected()` on writes that must hit a row.
- **Services** orchestrate repositories — no raw SQL, no direct pool use. Constructors take the repositories they need (`NewAuthService(roles, permissions)`); `factory.InitServices` is the only caller.
- **Seeds** (`internal/config/seeds/`) orchestrate repositories from `(ctx, db)`. They must be idempotent (second run changes nothing: `ON CONFLICT DO NOTHING` + existence checks) and must construct their own repositories — never consume `factory` globals (seeds run before the factory exists; see `main.go` order). New seed JSON goes in `internal/_data/`.
- **Handlers** are plain functions taking `*gin.Context`, wired in `routes/`. No handler structs.
- **Models** hold data + pure methods (hashing, tokens, validation). No DB access from models — `repository` imports `models`, never the reverse.
- **DTOs** may import `utils`; `utils` must import nothing internal. If you create a cycle, the build tells you with `import cycle not allowed` — the fix is always to move the shared shape down into `utils`, never to work around it.
- **IDs**: `utils.GenerateID()` (24-char hex strings, `TEXT` columns). Default empty IDs in repository `Create*` methods. Never `google/uuid` in app code.
- **JSON**: `common-go/utils` (`ToJSON`, `FromJSON`, `ToJSONArray`) — not `encoding/json` in app code (the models JWT builder is the deliberate exception, documented there).
- **Enums**: string types with `String()`, `IsValid()`, `<Type>Values()`. Validate at boundaries (handlers, seeds); keep const values stable — seed JSON and stored rows depend on the literal strings.

## Code style

- **Explicit over abstract.** Don't introduce shared private helpers for 2–3 call sites — inline the five lines in each caller with its own error text (production logs must say *which* operation failed). Delete a helper the moment it has one caller left.
- Comments explain *why* and *what it mirrors* (e.g. "mirrors TS `extractPermissions`"), not *what the code says*. Every repository method names the query var it wraps.
- `gofmt`-clean on touched files. No new linter is configured; match surrounding style.

## Migrations

- New file per change: `migrations/NNNNNN_name.up.sql`. Next number is max+1.
- Never edit an applied migration (applied DBs won't re-run it) and never add `*.down.sql` (the runner executes every `*.sql` file).
- New tables need PKs that match their `ON CONFLICT` clauses, or the upsert queries fail at runtime. New `users` columns used by scans must also be added to the relevant `USER_FIND_*` selects, or reads silently miss them.
- Verify by applying 1→N on a scratch Postgres, plus the upgrade path if you altered existing tables/columns.

## Tests

- Pure logic → `tests/unit/` (no DB, runs everywhere).
- Anything touching Postgres → `tests/integration/` using `testDB(t)` (skips when unreachable) and `truncateData(t, db)` between cases. Prefer structural assertions (links resolve, re-seed is a no-op, missing rows yield `ErrNoRows`) over exact row counts.
- A bug fix should arrive with the test that would have caught it (the enum count tests have already done this once).

## The `common-go` relationship

`go.mod` pins `common-go` with `replace => ../common-go`, so the sibling checkout is part of your build. Shared helpers (`ToJSON`, `ArrayIncludes`, crypto, …) live there when 2+ features need them — but a `replace` is a local-dev affordance: publishing means commit + tag `common-go`, bump the `require`, drop the `replace`. Coordinate that cutover explicitly; don't let the replace ride into a release unnoticed.
