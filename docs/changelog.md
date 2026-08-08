# Changelog

All notable changes to RepoFleet are listed here, newest first.

---

## v0.7.3 — 2026-08-07

### Added
- CI workflow that builds, vets, and tests every pull request targeting `main`
- `rf issue switch`'s interactive detail panel now shows the issue's created time and branch remote (when `--remote` was used); replaces the old `Name` field, since the name is now shown inline next to the ID in the list
- Interactive selectors (`rf issue switch`, `rf issue goto`, `rf workspace switch`) now support `q` to quit, in addition to `Ctrl+C`

### Changed
- When `rf issue create` or `rf issue repo add` needs to create a genuinely new branch (no existing local or remote branch found), it's now cut from the repo's local `main` (or `master`) instead of whatever branch happens to be checked out — keeps issue branches consistent across a workspace. The success output for each repo says which base was used, e.g. `created new local branch from main`
- `rf issue list` and `rf issue switch` now list issues newest-created first, instead of the order they were read from disk
- Terminal bell (triggered by promptui on invalid input) is now silenced in all interactive selectors and prompts

---

## v0.7.2 — 2026-08-01

### Added
- `rf issue create --branch` now reuses an existing branch instead of erroring: checks out a matching local branch as-is, or fetches and tracks it from the first configured remote that has it (origin first), only creating a new branch if it exists nowhere; `rf issue repo add` reuses the same logic
- `--remote <name>` flag to target a specific remote explicitly, for cases where the same branch name exists on more than one remote (e.g. a contributor's fork); persisted as the issue's branch remote so repos added later stay consistent
- `rf workspace remove` now prompts for confirmation, listing the repo/issue/snapshot counts that will be deleted (omitting any that are zero), and fully deletes the workspace's issues and snapshots along with it

### Changed
- `rf issue create` now fails atomically — if branch resolution fails for any repo, the issue is not saved and no repo is marked active

### Fixed
- Branch pattern is now shown when `rf issue create` errors on missing branch values, so users can see what tokens are expected
- Fixed a bug where a missing `default.yaml` caused an empty "default" workspace to be recreated on every run — including right after a user deliberately removed it

---

## v0.7.1 — 2026-07-24

### Added
- `rf snapshot create` — new `-n`/`--name` flag to tag a snapshot with a short memo, shown in `rf snapshot list`'s new `Name` column and in the create success message

### Changed
- `rf issue goto` now prints the "Current issue is #ID ..." header before the interactive repo picker, matching bare `rf issue`

---

## v0.7.0 — 2026-07-12

### Added
- `rf snapshot` — new top-level command to save and restore uncommitted changes across all repos in an issue as patch files, without using `git stash` or a local commit
  - `rf snapshot create [issue-id]` — capture staged, unstaged, untracked, and conflicted files; uses current issue if no ID given
  - `rf snapshot restore <issue-id> [hash]` — re-apply a snapshot; defaults to most recent
  - `rf snapshot list` — list all snapshots across the workspace
  - `rf snapshot remove <hash>` — delete a specific snapshot by hash (searches all issues)
  - `rf snapshot prune` — remove all snapshots for the current issue; `--all` for the entire workspace

### Fixed
- Issues now have a stable immutable short hash (like a git commit SHA) so the same issue ID can be safely reused across different contexts without ambiguity
- Workspace-scoped storage for issues — issue data is isolated per workspace

---

## v0.6.4 — 2026-07-15

### Fixed
- Backport: stable issue hash identity and workspace-scoped storage to the v0.6.x line

---

## v0.6.3 — 2026-07-13

### Fixed
- Normalize repo remote URLs consistently across platforms
- Handle missing repo paths gracefully — commands no longer crash when a repo's directory no longer exists on disk

---

## v0.6.1 — 2026-07-10

### Added
- `rf issue goto` promoted from a flag (`rf issue status --go-to`) to a dedicated subcommand
- Current repo shown first in the goto selector with a `●` indicator
- Repos that need a branch switch show a `→ branch` indicator in the list

---

## v0.6.0 — 2026-07-06

### Improved
- `rf issue sync` simplified to fetch-only — no rebasing, safe to run with uncommitted changes
- `rf issue goto` automatically switches to the issue branch on selection
- `rf issue repo remove` auto-switches to `main`/`master` before deleting a checked-out branch

---

## v0.5.4 — 2026-07-05

### Fixed
- Shell wrapper only intercepts `rf issue status --go-to`, not plain `rf issue status`

---

## v0.5.3 — 2026-07-05

### Fixed
- PowerShell profile detection queries `$PROFILE` correctly instead of using a hardcoded path

---

## v0.5.2 — 2026-07-05

### Fixed
- Shell integration installs to both PS5 and PS7 profiles on Windows

---

## v0.5.1 — 2026-07-05

### Fixed
- Use default stdio for `promptui` on Windows instead of `CONOUT$` to fix interactive selector on Windows terminals

---

## v0.5.0 — 2026-07-05

### Added
- `rf issue status` expanded with new columns: Commit, Age, HEAD± (includes untracked files)
- Interactive go-to mode (`--go-to`) to cd into a repo and switch branches from the status view
- Auto-install shell integration on first run for directory-changing support (bash, zsh, fish, PowerShell)

---

## v0.4.0 — 2026-07-05

### Added
- `rf issue repo add <name>` — add a repo to an existing issue context
- `rf issue repo remove <name>` — remove a repo from an issue; deletes branch if clean
- `--repo` flag in `rf issue create` now accepts comma-separated values

### Changed
- Replaced `settings.yaml` with lightweight plaintext `.current` pointer files for active workspace and issue tracking

### Removed
- Git passthrough command

---

## v0.3.0 — 2026-07-04

### Fixed
- Tightened command validation across repo and issue commands
- Improved error messages and UX for missing or invalid inputs

---

## v0.2.0 — 2026-07-02

### Added
- Interactive issue picker for `rf issue switch` — fuzzy-searchable list with ID, name, branch, and repo count
- Branch naming patterns with token support: `{workspace}`, `{issue}`, `{name}`, `{type}`, `{kind}`, `{description}`
- `rf workspace config` to view and set the branch pattern for a workspace

---

## v0.1.0 — 2026-06-30

### Added
- Initial release
- Core commands: `rf workspace switch/remove/config`, `rf repo add/remove/list`, `rf issue create/switch/list/sync/status/archive/remove`
- Homebrew (macOS/Linux) and Scoop (Windows) distribution via GoReleaser
