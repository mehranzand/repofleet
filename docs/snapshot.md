# snapshot

Save and restore uncommitted changes across all repositories in an issue — without using `git stash` or a local commit. Each snapshot captures staged changes, unstaged changes, untracked files, and conflicted files as patch files on disk.

Snapshots are stored at `~/.config/repofleet/snapshots/<workspace>/<issue-hash>/<snapshot-hash>/`.

---

## rf snapshot create

```
rf snapshot create [issue-id] [flags]
```

Save the current uncommitted diff of every repository in an issue. If no issue ID is provided, the active issue is used and a confirmation prompt is shown.

**Flags:**

| Flag | Description |
|---|---|
| `--clean` | Reset each repo's working tree to HEAD after saving its diff |
| `--dry-run` | Show what would be captured without writing anything |
| `-n`, `--name <name>` | A short memo/tag to help distinguish this snapshot later |

**Examples:**

```bash
rf snapshot create              # use current issue (prompts for confirmation)
rf snapshot create 123
rf snapshot create 123 --clean  # save then reset each repo to HEAD
rf snapshot create 123 --dry-run
rf snapshot create 123 -n "before refactor"
rf snapshot create 123 --name "before refactor"
```

---

## rf snapshot restore

```
rf snapshot restore <issue-id> [snapshot-hash] [flags]
```

Re-apply a snapshot's diff to each repository. Defaults to the most recently created snapshot for the issue. Pass a specific snapshot hash (from `rf snapshot list`) to restore an older one.

**Flags:**

| Flag | Description |
|---|---|
| `--dry-run` | Show what would be applied without changing any files |

**Examples:**

```bash
rf snapshot restore 123                    # restore most recent snapshot
rf snapshot restore 123 cf60033            # restore a specific snapshot by hash
rf snapshot restore 123 cf60033 --dry-run  # preview what would be applied
```

---

## rf snapshot list

```
rf snapshot list
```

List all snapshots across all issues in the current workspace, sorted by creation time (newest first).

**Columns:** Hash, Name, Created, Size, Repos

The `Name` column shows the memo/tag passed to `--name` at creation time, or `-` if none was set.

**Examples:**

```bash
rf snapshot list
```

---

## rf snapshot remove

```
rf snapshot remove <snapshot-hash>
```

Remove a specific snapshot by hash. Searches all issues in the workspace — no need to be on the issue that owns it.

A confirmation prompt is shown before deletion, including the issue ID the snapshot belongs to.

**Examples:**

```bash
rf snapshot remove cf60033
```

---

## rf snapshot prune

```
rf snapshot prune [flags]
```

Remove all snapshots for the current issue. With `--all`, removes snapshots across every issue in the workspace.

**Flags:**

| Flag | Description |
|---|---|
| `--all` | Remove snapshots across all issues in the workspace |
| `--force` | Skip the confirmation prompt |

**Examples:**

```bash
rf snapshot prune                  # all snapshots for current issue (prompts)
rf snapshot prune --all            # all snapshots in workspace (prompts)
rf snapshot prune --all --force    # all snapshots in workspace, no prompt
```
