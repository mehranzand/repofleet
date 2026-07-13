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

RepoFleet is an issue-centered CLI for managing Git workflows across multiple repositories.

When a feature or bug spans several services, you normally need to create matching branches, fetch updates, and check the status of each repository separately.

## Without RepoFleet

```bash
cd api
git switch -c fix/123-auth

cd ../frontend
git switch -c fix/123-auth

cd ../worker
git switch -c fix/123-auth
```

## With RepoFleet

```bash
rf issue create 123 --name auth --type fix
rf issue status
```

RepoFleet creates a single issue context, manages related branches across your repositories, synchronizes them, and shows their status in one dashboard—without repeatedly switching directories.

**One issue. Multiple repositories. One workflow.**

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
│   │                              --remote overrides the detected remote (default: git remote get-url origin)
│   │                              Remote is normalized to SSH form (git@host:owner/repo.git) regardless of input
│   ├── remove <name>              Remove a repository from the current workspace
│   └── list                       List repositories in the current workspace
│                                  Repos whose path no longer exists show a red ! marker
├── issue
│   ├── create <id>                Create an issue context (ID must be an integer)
│   │                              --name         short internal name (max 8 chars, no spaces)
│   │                              --description  short description
│   │                              --kind         bug | feature | task | story
│   │                              --type         feat | fix | chore | docs | refactor | test
│   │                              --repo         limit to specific repos, comma-separated (default: all in workspace)
│   │                              --skip-branch  save context without creating a git branch
│   │                              Errors out before creating any branch if a repo's path no longer exists
│   ├── list                       List every issue in the workspace, including archived, in a detailed table
│   │                              Columns: ID, Hash, Name, Branch, Status, Repos — current issue marked with *
│   ├── switch [id|name|hash]      Switch to an issue interactively, or by ID, name, or hash
│   │                              --archived     include archived issues in the list
│   │                              Each issue has an immutable short hash (like a git commit SHA),
│   │                              shown beside #issue in the interactive list — the same ID can be
│   │                              reused by unrelated issues across different repos; if an ID match
│   │                              is ambiguous, retry with the hash shown in the error
│   ├── repo
│   │   ├── add <repo-name>        Add a repo to the current issue (creates or switches branch)
│   │   │                          Errors if the repo's path no longer exists on disk
│   │   └── remove <repo-name>     Remove a repo from the current issue
│   │                              Deletes branch if clean; auto-switches to main/master if checked out
│   │                              Protected branches (main, master) are never deleted
│   │                              If the repo's path no longer exists, still unlinks it but skips branch cleanup
│   ├── sync                       Fetch all remotes for every repo in the current issue
│   ├── status                     Show status dashboard for the current issue
│   │                              Columns: Repo, Checkout, Commit, Age, HEAD±
│   │                              HEAD± includes untracked (new) files, not just tracked changes
│   │                              Repos with a missing path show a red ! marker instead of status
│   ├── goto                       Interactively select a repo, cd into it, and switch to the issue branch if needed
│   │                              Repos that need a branch switch show a → indicator in the list
│   │                              Repos with a missing path show a red ! marker and cannot be selected
│   ├── remove <id|name|hash>      Remove an issue context (does not delete git branches)
│   │                              --archived     also search archived issues
│   └── archive <id|name|hash>     Archive a completed issue context (only searches active issues)
└── snapshot
    ├── create <issue-id>          Save the uncommitted diff of every repo in an issue as a snapshot
    │                              --clean    reset each repo to HEAD after saving
    │                              --dry-run  preview what would be captured without writing anything
    ├── restore <issue-id> [hash]  Re-apply a snapshot's diff to each repo
    │                              Defaults to the most recent snapshot; pass a hash to restore an older one
    │                              --dry-run  preview what would be applied without changing any files
    ├── list                       List all snapshots in the current workspace
    ├── remove <hash>              Remove a specific snapshot by hash (searches all issues in the workspace)
    └── prune                      Remove all snapshots for the current issue
                                   --all    remove snapshots across all issues in the workspace
                                   --force  skip confirmation prompt
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

### 7. List and switch between issues

See every issue in the workspace, including archived, with full detail (ID, hash, name, branch, status, repos):

```bash
rf issue list
```

To switch, pick interactively from the list (current issue shown first):

```bash
rf issue switch
```

Or switch directly by ID, name, or hash:

```bash
rf issue switch 123
rf issue switch auth-fix
rf issue switch a1b2c3d
```

Since the same ID can be reused by unrelated issues in different repos, an ID lookup that matches more
than one issue returns an error listing each candidate's hash — retry with one of those, or run
`rf issue switch` with no argument to pick from the interactive list instead.

To include archived issues in the list:

```bash
rf issue switch --archived
```

### 8. Snapshots

A snapshot captures the uncommitted diff of every repo in an issue and saves it to disk — without using git stash or a local commit. Useful before a rebase, a branch switch, or handing work off to another machine.

Save the current state of all repos in an issue:

```bash
rf snapshot create 123
rf snapshot create 123 --clean    # also resets each repo to HEAD after saving
rf snapshot create 123 --dry-run  # preview what would be captured
```

List all snapshots across the workspace:

```bash
rf snapshot list
```

Restore the most recent snapshot (or a specific one by hash):

```bash
rf snapshot restore 123
rf snapshot restore 123 cf60033   # restore a specific snapshot
rf snapshot restore 123 --dry-run # preview what would be applied
```

Remove a specific snapshot by hash (searches all issues automatically):

```bash
rf snapshot remove cf60033
```

Remove all snapshots for the current issue, or the entire workspace:

```bash
rf snapshot prune               # current issue only (prompts for confirmation)
rf snapshot prune --all         # all issues in the workspace
rf snapshot prune --all --force # skip confirmation
```

### 9. Sync and wrap up

Fetch all remotes to keep remote refs up to date:

```bash
rf issue sync
```

When an issue is done, archive it (keeps the context for reference):

```bash
rf issue archive 123
```

`archive`, `remove`, and `switch` all accept an ID, name, or hash. To remove an issue entirely:

```bash
rf issue remove 123
```

By default `remove` only searches active issues; add `--archived` to also match archived ones.

<!-- ---

## Support
There are many ways to support RepoFleet:
Use it! Write about it! Star it! If you love RepoFleet, drop me a line and tell me what you love.
Sponsor my work at [](https://www.buymeacoffee.com/mehranzand)https://www.buymeacoffee.com/mehranzand

<a href="https://www.buymeacoffee.com/mehranzand" target="_blank">
  <img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" style="height: 60px !important;width: 217px !important;" ></a> -->
---

See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture details and how to contribute.
