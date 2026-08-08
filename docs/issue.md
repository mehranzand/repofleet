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
| `--branch <name>`, `-b` | Override branch name (ignores all naming rules) |
| `--remote <name>` | Target this remote specifically for branch reuse (default: search every configured remote, `origin` first) |
| `--name <name>` | Short internal name, max 8 chars, no spaces |
| `--description <text>` | Short description |
| `--kind <kind>` | `bug` \| `feature` \| `task` \| `story` |
| `--type <type>` | `feat` \| `fix` \| `chore` \| `docs` \| `refactor` \| `test` |
| `--repo <name>` | Limit to specific repos (repeatable, or comma-separated) |
| `--skip-branch` | Save context without creating a git branch — captures each repo's current branch instead |

**Branch name resolution order:**

1. `--branch` flag (overrides everything)
2. Workspace branch pattern (`rf workspace config --branch-pattern`)
3. Slugified issue ID (fallback)

**Notes:**

- Errors out before creating any branch if a repo's path no longer exists on disk.
- The same ID can be reused by unrelated issues, including within the same workspace, and creating a duplicate is not blocked. This is intentional: issue trackers are per-repo, so issue #12 in one repo and issue #12 in another are typically unrelated issues that happen to share a number, not the same issue. Each issue context also gets an immutable short hash (like a git commit SHA) shown in `rf issue list` — use the hash to disambiguate between issues sharing an ID.
- Per repo, the resolved branch name is reused rather than erroring if it already exists: a matching local branch is checked out as-is; otherwise, every configured remote is checked (`origin` first, then any others in the order git reports them) and the branch is fetched and checked out with tracking from the first remote that has it; only if it exists on no remote is a new local branch created.
- When a genuinely new local branch is created, it's always cut from the repo's local `main` (or `master` if there's no `main`) — regardless of what branch the repo currently has checked out. This keeps issue branches consistent across a workspace even if a repo was left on some other branch. The success output for each repo says which base was used, e.g. `created new local branch from main`. Errors if a repo has neither a local `main` nor `master` branch.
- `--remote <name>` skips that multi-remote search and only checks/tracks the branch on the named remote — useful when the same branch name exists on more than one remote (e.g. a contributor's fork) and you need to be explicit about which copy to use. Example: `--branch feature/x --remote forked-repo`. If a local branch already named `feature/x` exists but isn't already tracking `forked-repo/feature/x`, this errors instead of silently resetting or renaming it.
- When `--remote` is used, it's persisted on the issue (not just applied once at creation). `rf issue repo add` later reuses that same remote for any repo added to the issue afterward, so the whole issue stays consistent about which remote its branch comes from.
- If branch resolution fails in any repo (e.g. a git ref naming collision), the issue is not created and no repo is registered as active — nothing is saved. Repos where the branch operation already succeeded before the failure are left on that branch; this is not rolled back automatically.

**Examples:**

```bash
rf issue create 123 --name auth-fix --kind bug --type fix
rf issue create 456 --name dashboard --skip-branch
rf issue create 789 --name api-work --repo service-a --repo service-b
rf issue create 789 --name api-work --repo service-a,service-b
rf issue create 987 --branch review/follow-up
```

---

## rf issue list

```
rf issue list
```

List every issue in the current workspace (active and archived) in a detailed table, newest-created first. The currently active issue is marked with `*`.

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

Switch the active issue. Running without an argument opens an interactive selector, listing the current issue first, then the rest newest-created first.

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

Add a repository to the current issue. If the issue branch already exists locally in the repo, it's checked out as-is; if it exists on a remote instead (`origin` checked first, then others, unless the issue was created with `--remote`, in which case only that remote is checked), it's fetched and checked out with tracking; otherwise the branch is created — cut from the repo's local `main` (or `master`), regardless of what's currently checked out.

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
