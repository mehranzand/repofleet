# workspace

Manage workspaces. A workspace groups a set of repositories together and holds all issue contexts for that group.

A `default` workspace is created automatically on first run.

---

## rf workspace switch

```
rf workspace switch [name]
```

Switch to a workspace by name. If no workspace with that name exists, it is created automatically.

**Name rules:** letters, numbers, hyphens, and underscores only.

**Examples:**

```bash
rf workspace switch default
rf workspace switch backend
rf workspace switch mobile-team
```

---

## rf workspace remove

```
rf workspace remove <name>
```

Remove a workspace and all its issue contexts. Repositories on disk are not affected — only the RepoFleet metadata is deleted.

**Examples:**

```bash
rf workspace remove old-project
```

---

## rf workspace config

```
rf workspace config [flags]
```

View or update configuration for the current workspace. Running without flags prints the current settings.

**Flags:**

| Flag | Description |
|---|---|
| `--branch-pattern <pattern>` | Set the branch naming pattern for new issues |

**Branch pattern tokens:**

| Token | Value |
|---|---|
| `{workspace}` | Workspace name |
| `{issue}` | Issue ID |
| `{name}` | Issue name |
| `{description}` | Issue description |
| `{kind}` | Issue kind (`bug`, `feature`, `task`, `story`) |
| `{type}` | Issue type (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`) |

**Examples:**

```bash
rf workspace config                                    # view current settings
rf workspace config --branch-pattern "{type}/{issue}-{name}"
rf workspace config --branch-pattern "feature/{issue}"
```
