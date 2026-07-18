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

RepoFleet is an issue-centered CLI tool for organizing Git workflows. Track branches, snapshot uncommitted work, and stay in control whether you work in one repo or many.

Working on a feature or bug means juggling branches, tracking which repos are involved, and keeping uncommitted changes safe when you need to switch context. RepoFleet ties it all together under a single issue: create branches, check status, snapshot your work, and navigate without losing track.

---
<p align="center">
  <img src="assets/repofleet-demo.gif" alt="RepoFleet demo">
</p>

---

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
rf snapshot create 123   # preserve uncommitted work, no stash, no worktree
rf snapshot restore 123  # bring it back exactly, or compare AI results side by side
```

RepoFleet creates a single issue context, manages related branches, shows their status in one dashboard, and lets you snapshot uncommitted work at any point, without repeatedly switching directories or losing your place.

**One issue. Across repositories. One workflow.**

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

## Quick Start

```bash
rf workspace switch backend                            # create a workspace
rf repo add ~/code/service-a                           # add repos
rf repo add ~/code/service-b
rf issue create 123 --name auth-fix --type fix         # create an issue context
rf issue status                                        # dashboard across all repos
rf issue goto                                          # cd into a repo interactively
```

## Commands

```
rf
├── workspace
│   ├── switch [name]              Switch to a workspace, or create one if it doesn't exist
│   ├── remove <name>              Remove a workspace
│   └── config                     View or update workspace configuration
│                                  --branch-pattern  set branch naming pattern
├── repo
│   ├── add <path>                 Add a repository to the current workspace
│   ├── remove <name>              Remove a repository from the current workspace
│   └── list                       List repositories in the current workspace
├── issue
│   ├── create <id>                Create an issue context
│   │                              --name --description --kind --type --repo --skip-branch
│   ├── list                       List every issue in the workspace
│   ├── switch [id|name|hash]      Switch to an issue interactively or by ID, name, or hash
│   ├── sync                       Fetch all remotes for every repo in the current issue
│   ├── status                     Show status dashboard for the current issue
│   ├── goto                       Interactively select a repo and cd into it
│   ├── repo
│   │   ├── add <repo-name>        Add a repo to the current issue
│   │   └── remove <repo-name>     Remove a repo from the current issue
│   ├── remove <id|name|hash>      Remove an issue context
│   └── archive <id|name|hash>     Archive a completed issue context
└── snapshot
    ├── create [issue-id]          Save uncommitted changes for all repos in an issue
    ├── restore <issue-id> [hash]  Re-apply a saved snapshot
    ├── list                       List all snapshots in the workspace
    ├── remove <hash>              Remove a specific snapshot by hash
    └── prune                      Remove all snapshots for the current issue
                                   --all  remove across all issues in the workspace
```

## Documentation

Full command reference, flags, and examples: **[Documentation](https://mehranzand.github.io/repofleet)**

<!-- ---

## Support
There are many ways to support RepoFleet:
Use it! Write about it! Star it! If you love RepoFleet, drop me a line and tell me what you love.
Sponsor my work at [](https://www.buymeacoffee.com/mehranzand)https://www.buymeacoffee.com/mehranzand

<a href="https://www.buymeacoffee.com/mehranzand" target="_blank">
  <img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" style="height: 60px !important;width: 217px !important;" ></a> -->
---

See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture details and how to contribute.
