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
| `--forge <forge>` | Override the auto-detected forge |
| `--remote <url>` | Override the detected remote URL |

**Examples:**

```bash
rf repo add ~/code/service-a
rf repo add ~/code/service-b --forge github
rf repo add .                              # add the current directory
```

---

## rf repo remove

```
rf repo remove <name>
```

Remove a repository from the current workspace. The repository directory on disk is not affected — only the workspace registration is deleted.

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
