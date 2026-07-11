<img src="assets/repofleet-logo.svg" alt="repofleet" width="480" />

<p align="center">
  <a href="https://github.com/mehranzand/repofleet/releases/latest">
    <img src="https://img.shields.io/github/v/release/mehranzand/repofleet?style=flat-square" alt="Latest Release">
  </a>
  <a href="https://github.com/mehranzand/repofleet/actions/workflows/release.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/mehranzand/repofleet/release.yml?style=flat-square" alt="Build Status">
  </a>
  <a href="https://github.com/mehranzand/repofleet/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/mehranzand/repofleet?style=flat-square" alt="License">
  </a>
  <a href="https://github.com/mehranzand/homebrew-tap">
    <img src="https://img.shields.io/badge/homebrew-available-orange?style=flat-square&logo=homebrew" alt="Homebrew">
  </a>
  <a href="https://github.com/mehranzand/scoop-bucket">
    <img src="https://img.shields.io/badge/scoop-available-blue?style=flat-square&logo=windows" alt="Scoop">
  </a>
</p>

RepoFleet is an issue-centered CLI tool for managing Git workflows across multiple repositories.

When a feature spans multiple services, RepoFleet lets you create one issue context, branch all repos together, sync and track status across them — without switching directories.

---
<p align="center">
  <img src="assets/repofleet-demo.gif" alt="RepoFleet demo">
</p>

---
## Commands



```
rf
├── workspace
│   ├── switch [name]              Switch to a workspace, or create one if it doesn't exist
│   │                              Name must be letters, numbers, hyphens, or underscores
│   ├── remove <name>              Remove a workspace
│   └── config                     View or update workspace configuration
│                                  --branch-pattern  set branch naming pattern
├── repo
│   ├── add <path>                 Add a repository to the current workspace
│   │                              Forge is auto-detected from remote URL (--forge to override)
│   ├── remove <name>              Remove a repository from the current workspace
│   └── list                       List repositories in the current workspace
└── issue
    ├── create <id>                Create an issue context (ID must be an integer)
    │                              --name         short internal name (max 8 chars, no spaces)
    │                              --description  short description
    │                              --kind         bug | feature | task | story
    │                              --type         feat | fix | chore | docs | refactor | test
    │                              --repo         limit to specific repos, comma-separated (default: all in workspace)
    │                              --skip-branch  save context without creating a git branch
    ├── switch [id|name]           Switch to an issue interactively, by ID, or by name
    │                              --archived     include archived issues in the list
    ├── repo
    │   ├── add <repo-name>        Add a repo to the current issue (creates or switches branch)
    │   └── remove <repo-name>     Remove a repo from the current issue
    │                              Deletes branch if clean; auto-switches to main/master if checked out
    │                              Protected branches (main, master) are never deleted
    ├── sync                       Fetch all remotes for every repo in the current issue
    ├── status                     Show status dashboard for the current issue
    │                              Columns: Repo, Checkout, Commit, Age, HEAD±
    ├── goto                       Interactively select a repo, cd into it, and switch to the issue branch if needed
    │                              Repos that need a branch switch show a → indicator in the list
    ├── remove <id>                Remove an issue context (does not delete git branches)
    └── archive <id>               Archive a completed issue context
```

---

## Installation

### Homebrew (macOS/Linux)

```bash
brew install mehranzand/tap/repofleet
```

### Scoop (Windows)

```powershell
scoop bucket add mehranzand https://github.com/mehranzand/scoop-bucket
scoop install repofleet
```

### From source

Requires Go 1.22+.

```bash
git clone https://github.com/mehranzand/repofleet
cd repofleet
go build -ldflags="-X main.version=dev" -o rf ./cmd/repofleet
```

---

## Getting Started

### 1. Set up a workspace

A workspace groups your repos together. A `default` workspace is created automatically on first run. Switch to it or create a new one:

```bash
rf workspace switch default        # use the default workspace
rf workspace switch backend        # create and switch to a new workspace
```

### 2. Add repositories

Add each repo that belongs to this workspace. The forge (GitHub/GitLab) is auto-detected from the remote URL:

```bash
rf repo add ~/code/service-a
rf repo add ~/code/service-b
rf repo list                       # verify what's in the workspace
```

### 3. (Optional) Set a branch naming pattern

Define how branches are named when you create an issue. Available tokens:
`{workspace}` `{issue}` `{name}` `{description}` `{kind}` `{type}`

```bash
rf workspace config --branch-pattern "{type}/{issue}-{name}"
rf workspace config                # view current settings
```

### 4. Create an issue context

The issue ID must be an integer. Name is optional (max 8 chars, no spaces):

```bash
rf issue create 123 --name auth-fix --description "fix token refresh" --kind bug --type fix
```

This creates a branch in all workspace repos using the configured pattern. To track existing branches without creating new ones:

```bash
rf issue create 456 --name dashboard --skip-branch
```

With `--skip-branch`, each repo's current checked-out branch is captured as the issue branch.

To limit to specific repos:

```bash
rf issue create 789 --name api-work --repo service-a,service-b
```

### 5. Manage repos in an issue

Add or remove repos from the active issue after creation:

```bash
rf issue repo add service-c        # creates branch if new, switches if exists
rf issue repo remove service-c     # removes repo; auto-switches to main/master then deletes branch if clean
```

### 6. Work across repos

Check status of all repos for the current issue:

```bash
rf issue status
```

To interactively select a repo and `cd` into it:

```bash
rf issue goto
```

> Shell integration is set up automatically on first run. When prompted, run `source ~/.zshrc` (or your shell's rc file) once to activate it. After that, `rf issue goto` changes your working directory on selection. If the selected repo is not on the issue branch, it is automatically switched to it — repos that need a switch show a `→ branch` indicator in the list.

### 7. Switch between issues

Pick interactively from the list (current issue shown first):

```bash
rf issue switch
```

Or switch directly by ID or name:

```bash
rf issue switch 123
rf issue switch auth-fix
```

To include archived issues in the list:

```bash
rf issue switch --archived
```

### 8. Sync and wrap up

Fetch all remotes to keep remote refs up to date:

```bash
rf issue sync
```

When an issue is done, archive it (keeps the context for reference):

```bash
rf issue archive 123
```

To remove it entirely:

```bash
rf issue remove 123
```

---

See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture details and how to contribute.