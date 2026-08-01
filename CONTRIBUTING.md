# Contributing to RepoFleet
We follow a simple principle in development: **iteration over perfection**.

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
  factory/              dependency injection root (Workspace, GitRunner, IO)
  root/                 wires factory + registers all subcommands
  repocmd/              repo add / remove / list
  workspacecmd/         workspace switch / remove / config
  issuecmd/             issue create / switch / sync / status / goto / repo / remove / archive
  snapshotcmd/          snapshot create / restore / list / remove / prune
internal/               business logic — not importable outside the module
  store/                Workspace, Issue, and Snapshot types; YAML + pointer files at ~/.config/repofleet/
  snapshot/             snapshot capture and restore logic (patch files per repo)
  git/                  concurrent git runner (one goroutine per repo)
  iostreams/            injectable IO + ANSI color helpers
  util/                 shared helpers (short hash generation, confirmation prompt)
docs/                   MkDocs source — one file per command group
mkdocs.yml              MkDocs configuration (Material theme)
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

## Branch naming convention

Branches follow this pattern:

```
<type>/issue-<number>-<short-description>
```

For example, this documentation change lives on `docs/issue-22-improve-docs` — branch type `docs`, issue number `22`, and a short kebab-case description of the work.

- **`<type>`** — matches the kind of change, same vocabulary as commit types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`.
- **`issue-<number>`** — the GitHub issue this branch addresses. If there's no tracking issue yet, open one first — it keeps the PR linkable via `Closes #N`.
- **`<short-description>`** — a few kebab-case words summarizing the change, not the full issue title.

Other examples from this repo: `feat/issue-20-switch-into-first-repo`, `fix/issue-command`.

## Submitting a pull request

1. **Fork the repo** on GitHub, then clone your fork and add the upstream remote:
   ```bash
   git clone https://github.com/<your-username>/repofleet
   cd repofleet
   git remote add upstream https://github.com/mehranzand/repofleet
   ```
2. Create a branch following the [naming convention](#branch-naming-convention) above, e.g. `git checkout -b feat/issue-42-add-rename-command`.
3. Make your changes — keep each PR focused on one thing.
4. Run `go build ./...` and `go vet ./...` — both must pass.
5. Push to your fork and open a PR against `main`. Fill in the PR template.
6. Link the related issue with `Closes #N` in the PR description.

**One commit per PR.** Squash your work into a single commit before opening the PR. If review feedback requires changes, amend that same commit (`git commit --amend`) and force-push your branch — don't add follow-up commits. The PR should stay one commit from open to merge.

## The Merge Window Process

RepoFleet uses a two-week merge window for community contributions.

Pull requests can be submitted at any time. During each two-week cycle, maintainers review, test, and prepare approved pull requests for the next release.

To be included in a release, a pull request should:

- Pass all automated tests and checks.
- Be approved by a maintainer.
- Address all requested review comments.
- Be up to date with the target branch.
- Include documentation or tests when applicable.

At the end of each merge window, approved pull requests are merged and included in a new RepoFleet release.

Pull requests that are not ready before the window closes will remain open and may be included in the following release.

---

## Reporting a bug

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include your OS, Go version, and the exact command that failed.

## Requesting a feature

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md). Describe the problem first, then the proposed solution.
