# AI Documentation Index

This directory contains documentation written for AI assistants working inside this codebase. Read these files before making changes to understand structure, conventions, and constraints.

## Reading order

| File | Purpose |
|---|---|
| [overview.md](overview.md) | What this template is, tech stack, module rename, startup sequence |
| [architecture.md](architecture.md) | Layer diagram, dependency rules, data flow |
| [packages.md](packages.md) | Every `pkg/` package — purpose, constructor, config |
| [conventions.md](conventions.md) | TID propagation, logger usage, config pattern, error handling |
| [extending.md](extending.md) | Step-by-step: add a new domain resource end-to-end |
| [appendix.md](appendix.md) | All environment variables, full config struct reference |

## Quick orientation

```
cmd/server/main.go          ← wiring only; no logic
config/config.go            ← aggregates all pkg config structs
internal/router/router.go   ← ALL HTTP routes live here
internal/domain/            ← entities and port interfaces (no deps)
internal/usecase/           ← business logic (depends on ports)
internal/repository/        ← DB + cache implementations
internal/handler/           ← HTTP and Kafka entry points
pkg/                        ← self-contained infra packages (no internal imports)
```

## Hard rules (read before writing any code)

- Every string used in logic must be a named constant — no bare string literals for keys, prefixes, header names, or error messages. See `conventions.md` → String constants.
- `pkg/` packages never import `internal/`. Strings shared across `pkg/` packages are exported from the owning `pkg/` package (e.g. `logger.TIDKey`). Strings used only within `internal/` go in `internal/consts/consts.go`.
- All HTTP routes live in `internal/router/router.go` only.
- Always use `logger.FromContext(ctx)` — never `zap.L()` in application code.
- No `init()` functions.

## First thing to do

Replace every occurrence of `github.com/your-org/service-name` with your actual module path, then run `go mod tidy`.
