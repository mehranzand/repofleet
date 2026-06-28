# Contributing to RepoFleet

Thanks for your interest in contributing. This document covers how to set up, what patterns to follow, and how to submit changes.

---

## Prerequisites

- **Go 1.22+** — [download](https://go.dev/dl/)
- **Git**

```bash
go version   # should be 1.22 or higher
```

---

## Getting started

```bash
git clone https://github.com/mehranzand/repofleet
cd repofleet
go mod tidy
go build -ldflags="-X main.version=dev" -o repofleet ./cmd/repofleet
./repofleet
```

---

## Project layout

```
cmd/repofleet/          binary entry point (main.go)
commands/               Cobra CLI layer — flags, output, calls internal/
  factory/              dependency injection root (Config, GitRunner, IO)
  root/                 wires factory + registers all subcommands
  repo/                 repo add / remove / list
  issuecmd/             issue create / list / switch / sync / push / status / archive
  gitcmd/               passthrough git
internal/               business logic — not importable outside the module
  config/               Workspace + Repo types; YAML at ~/.config/repofleet/
  issue/                issue context entity; per-issue YAML state
  git/                  concurrent git runner (one goroutine per repo)
  iostreams/            injectable IO + ANSI color helpers
```

**Three-layer rule:** `commands/` may import `internal/`. `internal/` packages must not import `commands/`. `cmd/` imports only `commands/root`.

---

## Architecture patterns

Follow the patterns already in use — don't introduce new ones without discussion.

| Pattern | Where | Rule |
|---|---|---|
| Dependency Injection | `commands/factory` | Commands receive all clients via `*Factory`; never construct them directly |
| Factory Method | `New*()` constructors | Hide initialization from callers |
| Repository | `config.Load/Save`, `issue.Load/Save` | All disk access goes through these; no raw file I/O in commands |
| Facade | `git.Runner.Run()` | Concurrent multi-repo execution is hidden behind one call |
| Value Object | `config.Repo`, `issue.Context` | Plain structs, serialized to YAML — no methods that mutate external state |

---

## Adding a command

1. Create a file under the relevant `commands/` package (e.g. `commands/repo/rename.go`).
2. Write a `NewRenameCmd(f *factory.Factory) *cobra.Command` constructor.
3. Register it in the parent `NewCmd()` function.
4. Keep all output through `f.IO.Out` — never use `fmt.Println` directly.
5. Keep all git operations through `f.GitRunner.Run()`.

---

## Code style

- **No comments by default.** Only add one when the *why* is non-obvious — a hidden constraint, a workaround, a subtle invariant.
- **No dead code.** If a package, function, or field is unused, delete it.
- **No error swallowing.** Return errors up to the command layer; print them once and exit.
- **No global state** outside `cmd/repofleet/main.go`.
- Run `go vet ./...` before committing — CI will reject failures.

---

## Color and output

Use helpers from `internal/iostreams`:

```go
fmt.Fprintln(f.IO.Out, iostreams.Green("ok"))
fmt.Fprintln(f.IO.Out, iostreams.Cyan(repoName))
fmt.Fprintln(f.IO.Out, iostreams.Bold("Summary:"))
```

Never write ANSI codes directly. Helpers check `NO_COLOR` and TTY automatically.

---

## Building

```bash
# development build
go build -ldflags="-X main.version=dev" -o repofleet ./cmd/repofleet

# release build
go build -ldflags="-X main.version=0.x.0" -o repofleet ./cmd/repofleet

# verify no compilation errors across all packages
go build ./...
go vet ./...
```

---

## Submitting a pull request

1. Fork the repo and create a branch: `git checkout -b feat/your-feature`.
2. Make your changes — keep each PR focused on one thing.
3. Run `go build ./...` and `go vet ./...` — both must pass.
4. Open a PR against `main`. Fill in the PR template.
5. Link the related issue with `Closes #N` in the PR description.

---

## Reporting a bug

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include your OS, Go version, and the exact command that failed.

## Requesting a feature

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md). Describe the problem first, then the proposed solution.
