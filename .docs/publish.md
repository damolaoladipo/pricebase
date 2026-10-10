# Publishing and Releasing Pricebase API

This document defines how to structure, version, verify, and release the `pricebase` service. It is adapted from the `common-go` publish guide for an **executable application**, not a consumable Go library.

For local setup and architecture, see **[README.md](../README.md)**. For branching and PR flow, see **[CONTRIBUTING.md](../CONTRIBUTING.md)**.

## Service Identity

Module path:

```go
module github.com/damolaoladipo/pricebase
```

The GitHub repository URL and the `module` directive in `go.mod` must continue to match (confirm against the actual remote once one exists — this checkout has none yet).

**Pricebase API is a deployable HTTP service.** Clients (web, mobile) call it over HTTP at `/v1` (Socket.IO at `/socket/`). They do **not** `go get` this module as a library.

Shared Go helpers live in `common-go` (sibling checkout, see below). That repository has its own publish flow for library releases.

### Do Keep `cmd/app` as the Entry Point

This repository **is** an application. The only `package main` should live under `cmd/`:

```text
cmd/app/main.go
```

Build and run:

```bash
go run ./cmd/app
go build -o bin/pricebase ./cmd/app
```

Do not add a root-level `main.go`. Prefer one binary name (`pricebase`) unless you intentionally add more commands under `cmd/`.

### Do Use `internal/` for Application Code

Unlike a public library, this service **should** keep implementation under `internal/`. Go’s `internal` rule prevents other modules from importing these packages — which is what we want for an API server.

Do **not** move domain packages to the module root to “make them public.” External clients use HTTP, not Go imports.

### Do Not Treat the API Module as a Published Library

Do not instruct consumers to:

```bash
# Wrong for pricebase — this is not a library product.
go get github.com/damolaoladipo/pricebase@v1.0.0
```

Version tags on this repo mark **service releases** (deployable artifacts), not packages for `pkg.go.dev` consumption.

### Local `replace` for `common-go`

Development uses a sibling checkout:

```go
replace github.com/damolaoladipo/common-go => ../common-go
```

Rules:

- Keep `replace` for **local development** when iterating on common-go next to the API.
- **Release / CI builds** must resolve a real tagged version of `common-go` (remove or override the relative `replace`, or use a CI-only `go.mod` strategy your pipeline defines).
- Never commit secrets in `.env`. Commit `.env.example` only.

Before tagging a release, confirm how production will resolve common-go:

```bash
grep -E '^(require|replace).*common-go' go.mod
```



## Pre-Release Verification

Do not tag or deploy if any required check fails.

### 1. Confirm the Module Path

```bash
head -n 1 go.mod
```

Expected:

```text
module github.com/damolaoladipo/pricebase
```



### 2. Normalize Dependencies

```bash
go mod tidy
```

Commit both `go.mod` and `go.sum`. Do not ignore `go.sum`.

### 3. Format and Verify

```bash
gofmt -l .                  # must print nothing (format only files you touched)
go vet ./...
go test ./...
POSTGRES_URL="..." go test ./tests/integration/ -v
go build -o bin/pricebase ./cmd/app
```

Smoke-run locally (Postgres + Redis + `.env`):

```bash
cp -n .env.example .env
# set JWT_SECRET, POSTGRES_URL, REDIS_URL, SUPERADMIN_EMAIL
./bin/pricebase
# or: go run ./cmd/app
```

Confirm:

- process runs migrations 1→N, seeds (roles → permissions → links → users → attach), then listens on `APP_PORT`;
- routes answer under `/v1`, Socket.IO upgrades at `/socket/`;
- `CORS_ORIGINS` includes required client origins;
- no panic on boot when Postgres and Redis are reachable.



### 4. Migrations and Seeds Get Extra Scrutiny

- Never edit an applied migration; add a new `NNNNNN_name.up.sql`. Never add `*.down.sql` (the runner executes every `*.sql` file).
- New tables need PKs matching their `ON CONFLICT` clauses; new `users` columns used by scans must be added to the matching `USER_FIND_*` selects.
- Verify greenfield (empty DB, migrations 1→N + seeds) **and** the upgrade path (previous release state + new migrations) on a scratch Postgres. Never verify against shared/dev databases.



### 5. Review the Pending Release

```bash
git status
git diff
git diff --stat
```

Confirm that:

- no credentials or `.env` files are included;
- no binaries (`bin/`), logs, or coverage artifacts are included;
- `README.md` / `.env.example` document new env keys and HTTP contract changes;
- new source files carry no license headers (the repo-wide `LICENSE`, AGPL-3.0, covers them);
- breaking HTTP API changes are identified before selecting the version number;
- `common-go` dependency is intentional (tag vs local `replace`).



### 6. Commit and Push

Follow [CONTRIBUTING.md](../CONTRIBUTING.md) (`staging` → `release/*` → `master`); do not push releases straight to `master`.

## Versioning (Service Releases)

Use annotated Git tags in semver form:

```text
vMAJOR.MINOR.PATCH
```

Tags identify **deployable API releases**. Prefer the same number for GitHub Releases and container image tags (e.g. `pricebase:v1.2.0`) when you introduce Docker/registry publishing.

### Patch Release

Increment `PATCH` for backward-compatible fixes that do not change the HTTP contract in a breaking way.

```text
v1.0.0 -> v1.0.1
```

Examples:

- bug fixes in handlers or middleware;
- logging / reliability improvements;
- dependency patches without breaking responses;
- docs and tests only.



### Minor Release

Increment `MINOR` for backward-compatible API or capability additions.

```text
v1.0.1 -> v1.1.0
```

Examples:

- new endpoints or optional fields;
- new routes that existing clients can ignore;
- new migrations/seeds that apply cleanly on upgrade;
- new env keys with safe defaults.

Reset `PATCH` to zero when incrementing `MINOR`.

### Major Release

Increment `MAJOR` for breaking HTTP or operational contracts.

```text
v1.4.2 -> v2.0.0
```

Examples:

- removing or renaming endpoints;
- incompatible response/request schema changes;
- auth scheme changes that break existing clients;
- required env/config changes without migration path;
- migrations that cannot upgrade existing databases.

Coordinate major bumps with web/mobile clients. Unlike Go module `/v2` import paths, **HTTP majors** are communicated via release notes and client upgrades — you do not change the Go module path to `/v2` unless you intentionally publish this repo as a library (you should not).

## Creating a Version Tag



### First release

This repo has no tags yet. After the first merge to the release line:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

Confirm the tag:

```bash
git ls-remote --tags origin refs/tags/v1.0.0
git show v1.0.0:go.mod | head -n 1
```



### Subsequent releases

```bash
# Patch
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1

# Minor
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# Major
git tag -a v2.0.0 -m "Release v2.0.0"
git push origin v2.0.0
```

**Never** move, delete, or recreate a tag that has already been deployed or referenced by clients/ops. Fix forward with a new patch.

## Deploying a Release

Publishing this service means **shipping a running binary or container**, not registering a module on `proxy.golang.org`.

### Build artifact

```bash
go build -o bin/pricebase ./cmd/app
```



### Runtime configuration

Production must set at least:


| Variable           | Notes                                                          |
| ------------------ | -------------------------------------------------------------- |
| `JWT_SECRET`       | Strong secret; never reuse example values                      |
| `POSTGRES_URL`     | Production database (migrations run at boot)                   |
| `REDIS_URL`        | Production Redis (`rediss://` for TLS, cluster flag if needed) |
| `APP_PORT`         | Listen port behind your proxy (default `5025`)                 |
| `APP_ENV`          | `production` (drives gin mode, logging, key domains)           |
| `CORS_ORIGINS`     | Explicit production web/app origins                            |
| `SUPERADMIN_EMAIL` | Account to bootstrap as superadmin on first boot               |


Do not bake `.env` into images. Inject env via the host, orchestrator, or secret store.

### Client alignment

After a release, verify health (`/v1/health`) and one authenticated path from each critical client (web portal, mobile). Document the client-side API origin env each client reads if it isn't already recorded.

### Container / platform (when added)

When a `Dockerfile` or deploy pipeline exists under this repo or `ops/`:

1. Build the image from the tagged commit.
2. Tag the image with the same semver (`v1.2.0`) and optionally `latest` for non-prod only.
3. Deploy to staging first; smoke-test; then promote to production.
4. Record the Git tag + image digest in the release notes.

Until Docker/CI is committed here, treat `go build` + process supervisor (or your existing host) as the deploy path, and keep this section updated when automation lands.

## Pre-Release Versions

Use pre-release tags when a build is not ready for production:

```bash
git tag -a v1.1.0-beta.1 -m "Release v1.1.0 beta 1"
git push origin v1.1.0-beta.1
```

Deploy pre-releases only to staging or ephemeral environments. Promote with a new stable tag after validation:

```bash
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```



## Handling a Bad Release

Do not rewrite a published tag or force-push over a deployed commit clients already use.

1. Fix forward on a branch.
2. Merge through your normal release path.
3. Tag a new patch and deploy it.

```bash
# After the fix is on master / release line:
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
```

If you must warn operators, document the bad tag in release notes (`v1.0.1` withdrawn — use `v1.0.2`). The Go `retract` directive applies to **library** modules; it is not the primary tool for an HTTP service. Prefer ops runbooks and client communication.

## Dependency on `common-go`


| Situation                             | Action                                                                        |
| ------------------------------------- | ----------------------------------------------------------------------------- |
| Local feature work                    | `replace => ../common-go` is fine                                             |
| API release needs a new common-go fix | Publish common-go first, then bump the `require` version in this `go.mod`     |
| CI / production build                 | Resolve a tagged common-go version; do not rely on a developer’s sibling path |


```bash
go get github.com/damolaoladipo/common-go@v1.x.x
go mod tidy
go test ./...
go build -o bin/pricebase ./cmd/app
```



## Troubleshooting



### Binary builds but process exits on boot

- Confirm `.env` or process env is loaded from the working directory.
- Check `JWT_SECRET`, `POSTGRES_URL`, `REDIS_URL`, and `SUPERADMIN_EMAIL`.
- Confirm Postgres and Redis are reachable from the host.
- Read the migration error: a failing `NNNNNN_*.sql` aborts boot before serving.



### Clients cannot reach the API

- Align `APP_PORT` (default `5025`) with what clients expect.
- Confirm `CORS_ORIGINS` includes the client origin.
- Confirm Socket.IO clients connect at `/socket/` (WebSocket transport only).



### `replace` breaks CI

CI hosts usually lack `../common-go`. Remove or override the relative replace and use a tagged common-go module version for that pipeline.

### Accidental library-style publish expectations

`go list -m github.com/damolaoladipo/pricebase@v1.0.0` may work if the repo is public, but **that is not how this product is consumed**. Success on the Go proxy does not mean the API is deployed. Always verify a running instance and HTTP smoke tests.

## Release Checklist (copy/paste)

```text
[ ] gofmt / go vet / go test (incl. integration) / go build ./cmd/app
[ ] migrations apply greenfield AND upgrade on scratch Postgres
[ ] seeds + attach run clean twice (idempotent)
[ ] .env.example updated for new keys
[ ] README notes for breaking or notable HTTP changes
[ ] common-go version intentional (no stale local-only replace for prod)
[ ] Merged via staging / release branch per CONTRIBUTING.md
[ ] Annotated tag vX.Y.Z pushed
[ ] Artifact or image built from that tag
[ ] Staging deploy + smoke test
[ ] Production deploy + client smoke test
[ ] Release notes posted (GitHub Release recommended)
```



## Official References

- [Go modules reference](https://go.dev/ref/mod) (dependency management inside this repo)
- [Module release workflow](https://go.dev/doc/modules/release-workflow) (useful for `common-go`, not HTTP deploy)
- [Semantic Versioning](https://semver.org/)
- Service docs: [README.md](../README.md), [CONTRIBUTING.md](../CONTRIBUTING.md)

