# RepoFleet

RepoFleet is an issue-centered CLI tool for organizing Git workflows. track branches, snapshot uncommitted work, and manage your codebase across one or more repositories.

When a feature or bug spans several services, you normally need to create matching branches, fetch updates, and check the status of each repository separately. RepoFleet ties them together under a single issue context.

## Command Groups

| Command | Description |
|---|---|
| [`rf workspace`](workspace.md) | Create and switch workspaces that group your repos |
| [`rf repo`](repo.md) | Add, remove, and list repositories in a workspace |
| [`rf issue`](issue.md) | Create issue contexts, switch between them, manage repos and branches |
| [`rf snapshot`](snapshot.md) | Save and restore uncommitted changes across repos without stashing |

## Quick Start

```bash
# 1. Set up a workspace
rf workspace switch backend

# 2. Add repos
rf repo add ~/code/service-a
rf repo add ~/code/service-b

# 3. Create an issue context
rf issue create 123 --name auth-fix --type fix

# 4. Check status across all repos
rf issue status

# 5. Navigate to a repo
rf issue goto
```

## Storage

RepoFleet stores all data in `~/.config/repofleet/`:

```
~/.config/repofleet/
  workspaces/          workspace configs and issue contexts
  snapshots/           snapshot patch files, keyed by workspace → issue → hash
  current              pointer to the active workspace
```
