# repo

Manage repositories in the current workspace. Repositories must be added to a workspace before they can be included in issue contexts.

---

## rf repo add

```
rf repo add <path>
```

Add a repository to the current workspace. The forge (GitHub, GitLab, etc.) is auto-detected from the remote URL. The remote is normalized to SSH form (`git@host:owner/repo.git`) regardless of how it is configured locally.

**Flags:**

| Flag | Description |
|---|---|
| `--name <name>`, `-n` | Name for the repo (default: directory basename) |
| `--forge <forge>`, `-f` | Override the auto-detected forge (`github` or `gitlab`) |
| `--remote <url>`, `-u` | Override the detected remote URL (default: `git remote get-url origin`) |

**Notes:**

- `<path>` must exist and be a git repository, or the command errors.
- Errors if a repo with the same name, or the same path, already exists in the workspace.
- If `--forge` isn't given and the forge can't be auto-detected from the remote URL, errors asking for `--forge github` or `--forge gitlab` explicitly.

**Examples:**

```bash
rf repo add ~/code/service-a
rf repo add ~/code/service-b --forge github
rf repo add .                              # add the current directory
rf repo add ~/code/service-c --name svc-c
```

---

## rf repo remove

```
rf repo remove <name>
```

Remove a repository from the current workspace. The repository directory on disk is not affected; only the workspace registration is deleted.

**Examples:**

```bash
rf repo remove service-a
```

---

## rf repo list

```
rf repo list
```

List all repositories registered in the current workspace. Repositories whose path no longer exists on disk are shown with a red `!` marker.

**Examples:**

```bash
rf repo list
```
