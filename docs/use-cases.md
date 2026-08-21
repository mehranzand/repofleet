# Use Cases

This page walks through RepoFleet by scenario rather than by command. If [workspace](workspace.md), [repo](repo.md), [issue](issue.md), and [snapshot](snapshot.md) are the reference, this is the "why would I run that" companion.

---

## Use case 1: One issue, several repos

A fix or feature that touches `api`, `worker`, and `frontend` normally means creating the same branch three times, by hand, and hoping the name matches everywhere.

```bash
rf workspace switch backend
rf repo add ~/code/api
rf repo add ~/code/worker
rf repo add ~/code/frontend

rf issue create 123 --name auth-fix --kind bug --type fix
```

One command creates a matching branch in every repo in the workspace (or a subset, via `--repo`), each cut from that repo's own `main`/`master`. The output confirms what happened per repo, worth reading rather than skimming:

```
Resolving branch "fix/123-auth-fix" in 3 repo(s)...

  ✓ ~/code/api — created new local branch from main
  ✓ ~/code/worker — created new local branch from main
  ✓ ~/code/frontend — created new local branch from main
```

If a repo already had that branch name in use, locally or on a remote, it's reused instead of recreated, and the line says so (`branch already exists — switched to it`, or `fetched and tracked from origin`). Worth double-checking on a fresh issue: a `fetched and tracked` line here usually means the branch name collided with something unrelated on the remote, not that this issue was somehow already in progress.

**Why the branch pattern matters:** by default RepoFleet slugifies the issue ID into a branch name. Set a pattern once per workspace and every future issue in that workspace names its branches consistently, with no per-issue typing:

```bash
rf workspace config --branch-pattern "{type}/{issue}-{name}"
# rf issue create 123 --name auth-fix --type fix
# → branch: fix/123-auth-fix everywhere
```

Without a pattern, branch naming drifts repo to repo and person to person: `fix/123`, `123-auth-fix`, `bug/123_auth`. With one, every teammate who runs `rf issue create` for the same ID lands on the *same* branch name, which is what makes the next use case (someone else picking up your issue) work without coordination.

`--branch` overrides the pattern entirely for a one-off case (e.g. reusing an existing branch name that doesn't match the convention).

---

## Use case 2: Picking up someone else's issue (reviewing across repos)

RepoFleet has no separate "reviewer" role; a reviewer just runs the same commands as the issue owner, pointed at the same branch. The branch pattern's job is generating a *new* branch name when an issue is first created; it's not relevant here, because for a reviewer the branch already exists: it was created once, by the owner, and every repo already has it. So the reviewer doesn't need the pattern to derive anything; they need the exact name that already exists, passed straight through with `--branch`. What a reviewer needs from the owner is three concrete pieces of information:

- the **issue ID**, already known, since it's the underlying tracker ticket number
- the **names of the affected repos** (`rf repo list` on the owner's side)
- the **shared branch name**, the one branch name that's the same across every affected repo for this issue (`rf issue status` shows it, in the Checkout column, for each repo)

With those three inputs in hand, the reviewer checks which workspace they're in (`rf workspace`, no subcommand), adds the same repos, and creates their own issue context with `--branch`, passing the exact shared name rather than letting it be derived:

```bash
rf workspace          # confirm which workspace this lands in
rf repo add ~/code/api
rf repo add ~/code/worker

rf issue create 123 --branch fix/123-auth-fix
rf issue              # confirm this is now the active issue
```

`--branch` skips branch-pattern resolution entirely, so it works regardless of what pattern (if any) is configured in the reviewer's workspace. Because an existing branch is checked out as-is (locally, or fetched from the first remote that has it) rather than recreated, this lands the reviewer on the owner's exact branch in every repo, not a new one.

The per-repo line printed by `rf issue create` says exactly which of those happened, so it's worth reading rather than skimming:

```
Resolving branch "fix/123-auth-fix" in 2 repo(s)...

  ✓ ~/code/api — fetched and tracked from origin
  ✓ ~/code/worker — branch already exists — switched to it
```

`fetched and tracked from origin` (or `branch already exists — switched to it`, if it was already local) confirms the reviewer landed on the owner's actual branch. `created new local branch from main` on any repo is the sign something's wrong: the branch name didn't match what's on the remote, and the reviewer is now reviewing an empty branch instead of the owner's changes.

From there, reviewing is just:

```bash
rf issue status   # checkout, last commit, and uncommitted diff size across every repo, one table
rf issue goto      # jump into whichever repo needs a closer look
```

No manual `cd`, no remembering which of the three repos still has uncommitted worker changes.

---

## Use case 3: `switch` and `goto`: not losing your place

**`rf issue switch`** changes which issue is "active", useful when you're interrupted mid-fix by a more urgent issue:

```bash
rf issue switch          # interactive picker, current issue listed first
rf issue switch 456      # jump straight to issue 456
rf issue switch          # later, come back: issue 123 is right there in the list
```

Nothing about issue 123's repos or branches is touched by switching away from it; the context is just parked, not lost.

**`rf issue goto`** operates *within* the active issue: it's the multi-repo version of `cd` + `git switch`. Given an issue with three repos, instead of:

```bash
cd ../api && git switch fix/123-auth-fix
cd ../worker && git switch fix/123-auth-fix
```

you run:

```bash
rf issue goto
```

pick a repo from the list (the one you're already in, if any, is marked `●`), and it `cd`s you there and puts you on the issue branch, switching automatically if the repo was left on something else (shown as `→ branch` in the picker). Repos with missing paths are flagged and skipped. This is what makes bouncing between the repos of a many-repo issue fast enough to actually do, rather than something you avoid.

---

## Use case 4: Snapshot: pausing uncommitted work across repos

`git stash` works, but it's scoped to a single repo and a single stack. Once an issue spans several repos, that stops matching how the work actually needs to be paused and resumed:

| | `git stash` | `rf snapshot` |
|---|---|---|
| Scope | One repo | Every repo in the issue, together |
| Identity | Position in a per-repo stack (`stash@{0}`); meaning drifts as you stash more things in that repo for unrelated reasons | One hash per snapshot, tagged with `--name`, listed workspace-wide |
| Restore | Must `stash pop` in each repo, in the right order, and not forget one | One `rf snapshot restore <issue-id>` reapplies to every repo |
| Collisions | Shares the stash stack with any other unrelated work you (or a script) stash in that repo | Stored separately under the issue, never mixed with anything else |
| Portability | Lives in the repo's local `.git`, tied to that clone | Plain patch files under `~/.config/repofleet/snapshots/...` |

**Why this matters with several issues in flight:** if you're carrying uncommitted changes on issue 123 across three repos and need to jump to a production hotfix, `git stash` in each of the three repos leaves you with three anonymous stash entries to track and unwind correctly later. `rf snapshot create` captures all three repos' diffs (staged, unstaged, untracked, even conflicted) as one named unit tied to issue 123:

```bash
rf snapshot create 123 -n "before refactor"
rf snapshot create 123 --clean   # snapshot, then reset every repo's working tree to HEAD
```

Later, possibly after other snapshots have been taken for other issues, or by other people using the same repos, bring it back exactly:

```bash
rf snapshot list                 # every snapshot in the workspace, newest first
rf snapshot restore 123          # most recent snapshot for issue 123
rf snapshot restore 123 cf60033  # a specific older one, by hash
```

Because a snapshot is a plain patch set rather than a stash-stack entry, it can also be restored somewhere else entirely, e.g. to compare two different attempts at a fix (including AI-generated ones) side by side without re-running the change that produced them.

---

## Data checklist by role

| You are... | You need |
|---|---|
| **Issue owner, starting fresh** | A workspace with the relevant repos added ([repo add](repo.md)); an issue ID; a branch pattern configured once (optional but recommended) |
| **Reviewer / picking up an existing issue** | The affected repo names, added to your own workspace; the issue ID; the exact shared branch name (from `rf issue status`), passed via `--branch` |
| **Anyone pausing uncommitted work** | Nothing extra: `rf snapshot create` works against whatever the active issue already has |
