<img src="assets/repofleet-logo.svg" alt="repofleet" width="480" />

RepoFleet is an issue-centered CLI tool for managing Git workflows across multiple repositories.

When a feature spans multiple services, RepoFleet  lets you create one issue context, branch all repos together, run git commands across them in parallel, and track every open MR/PR — without switching directories.

---

## Commands

```
repofleet
├── repo
│   ├── add <path>                 Add a repository to a workspace
│   ├── remove <name>              Remove a repository from a workspace
│   └── list                       List repositories in the current workspace
├── git [git args...]              Run any git command across all workspace repos
└── issue
    ├── create <id>                Create an issue context across selected repos
    ├── list                       List all issue contexts
    ├── switch <id>                Switch all repos to the issue branch
    ├── sync                       Fetch and rebase all repos for the current issue
    ├── push                       Push all issue branches to their remotes
    ├── status                     Show status dashboard for the current issue
    └── archive <id>               Archive a completed issue context
```

---

## Architecture

Three layers — each with one job, depending only on the layer below.

```
cmd/repofleet/main.go           binary entry point; version via ldflags
        │
commands/                       Cobra CLI layer — parse flags, call internal, print output
  ├── root/                     wires factory + registers all subcommands
  ├── factory/                  dependency injection: Config, GitRunner, IO
  ├── repo/
  ├── gitcmd/
  └── issuecmd/
        │
internal/                       business logic — compiler-enforced, not importable outside
  ├── config/                   Workspace + Repo types; YAML persistence (~/.config/repofleet/)
  ├── issue/                    issue context entity; per-issue YAML state
  ├── git/                      concurrent git runner (one goroutine per repo)
  └── iostreams/                injectable IO + ANSI color helpers
```

## Getting Started

**Install Go** (if not already installed):
```bash
wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
```

**Build:**
```bash
go mod tidy
go build -ldflags="-X main.version=0.1.0" -o repofleet ./cmd/repofleet
```

**Run:**
```bash
./repofleet
```