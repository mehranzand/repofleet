# issue

Manage issue contexts across repositories. An issue context ties a set of repositories and their branches together under a single identifier, so you can create, switch, sync, and inspect them all at once.

---

## rf issue create

```
rf issue create <id> [flags]
```

Create an issue context. The ID must be an integer (e.g. a Jira or GitHub issue number). A branch is created in every repo in the workspace using the configured branch pattern.

**Flags:**

| Flag | Description |
|---|---|
| `--name <name>` | Short internal name, max 8 chars, no spaces |
| `--description <text>` | Short description |
| `--kind <kind>` | `bug` \| `feature` \| `task` \| `story` |
| `--type <type>` | `feat` \| `fix` \| `chore` \| `docs` \| `refactor` \| `test` |
| `--repo <name>` | Limit to specific repos (repeatable, or comma-separated) |
| `--skip-branch` | Save context without creating a git branch — captures each repo's current branch instead |

**Notes:**

- Errors out before creating any branch if a repo's path no longer exists on disk.
- The same ID can be reused by unrelated issues in different workspaces. Each issue also gets an immutable short hash (like a git commit SHA) shown in `rf issue list` — use the hash to disambiguate.

**Examples:**

```bash
rf issue create 123 --name auth-fix --kind bug --type fix
rf issue create 456 --name dashboard --skip-branch
rf issue create 789 --name api-work --repo service-a --repo service-b
rf issue create 789 --name api-work --repo service-a,service-b
```

---

## rf issue list

```
rf issue list
```

List every issue in the current workspace (active and archived) in a detailed table. The currently active issue is marked with `*`.

**Columns:** ID, Hash, Name, Branch, Status, Repos

**Examples:**

```bash
rf issue list
```

---

## rf issue switch

```
rf issue switch [id|name|hash]
```

Switch the active issue. Running without an argument opens an interactive selector. The current issue is shown first in the list.

**Flags:**

| Flag | Description |
|---|---|
| `--archived` | Include archived issues in the list |

**Notes:**

- Accepts an ID, short name, or immutable hash.
- If an ID matches more than one issue, the command errors and lists each candidate's hash — retry with one of those, or run `rf issue switch` with no argument.

**Examples:**

```bash
rf issue switch                    # interactive
rf issue switch 123
rf issue switch auth-fix
rf issue switch a1b2c3d            # by hash
rf issue switch --archived         # include archived in selector
```

---

## rf issue sync

```
rf issue sync
```

Fetch all remotes for every repository in the current issue. Keeps remote refs up to date without modifying local branches.

**Examples:**

```bash
rf issue sync
```

---

## rf issue status

```
rf issue status
```

Show a status dashboard for all repositories in the current issue.

**Columns:** Repo, Checkout, Commit, Age, HEAD±

- `HEAD±` includes both tracked changes and untracked (new) files.
- Repos with a missing path show a red `!` marker instead of status.

**Examples:**

```bash
rf issue status
```

---

## rf issue goto

```
rf issue goto
```

Interactively select a repository, `cd` into it, and switch to the issue branch if the repo is not already on it.

- Prints a "Current issue is #ID (N repos) in the WORKSPACE workspace" header before the picker, same as running bare `rf issue`.
- The current repo (if your shell is inside one of the issue's repos) is shown first with a `●` indicator.
- Repos that need a branch switch show a `→ branch` indicator.
- Repos with a missing path show a red `!` marker and cannot be selected.

**Shell integration:** set up automatically on first run. When prompted, run `source ~/.zshrc` (or your shell's rc file) once to activate directory-changing. After that, `rf issue goto` changes your working directory on selection.

**Examples:**

```bash
rf issue goto
```

---

## rf issue repo add

```
rf issue repo add <repo-name>
```

Add a repository to the current issue. If the issue branch already exists in the repo, the repo is switched to it; otherwise the branch is created.

Errors if the repo's path no longer exists on disk.

**Examples:**

```bash
rf issue repo add service-c
```

---

## rf issue repo remove

```
rf issue repo remove <repo-name>
```

Remove a repository from the current issue.

- If the repo is currently checked out on the issue branch, it automatically switches to `main` or `master` first.
- Deletes the issue branch if it has no uncommitted changes.
- Protected branches (`main`, `master`) are never deleted.
- If the repo's path no longer exists on disk, the repo is still unlinked from the issue but branch cleanup is skipped.

**Examples:**

```bash
rf issue repo remove service-c
```

---

## rf issue remove

```
rf issue remove <id|name|hash>
```

Remove an issue context entirely. Does not delete git branches.

**Flags:**

| Flag | Description |
|---|---|
| `--archived` | Also search archived issues |

**Examples:**

```bash
rf issue remove 123
rf issue remove auth-fix
rf issue remove a1b2c3d --archived
```

---

## rf issue archive

```
rf issue archive <id|name|hash>
```

Archive a completed issue context. Archived issues are hidden from the default list and switch selector but kept for reference. Only searches active issues.

**Examples:**

```bash
rf issue archive 123
rf issue archive auth-fix
```
