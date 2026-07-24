package snapshot

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util/git"
	"gopkg.in/yaml.v3"
)

func Create(runner *git.Runner, issue *store.Issue, name string, clean, dryRun bool) (*store.Snapshot, []string, error) {
	hash, err := store.NewSnapshotHash()
	if err != nil {
		return nil, nil, err
	}

	snap := &store.Snapshot{
		Hash:      hash,
		IssueID:   issue.ID,
		IssueHash: issue.Hash,
		Workspace: issue.Workspace,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
	}
	chain := NewChain()

	type cleanupStep struct {
		repoPath     string
		hadConflicts bool
		hadUntracked bool
	}
	var cleanups []cleanupStep

	for _, repo := range issue.Repos {
		// branch currently checked out, recorded for informational purposes only
		branch, err := runOne(runner, repo.Path, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", repo.Path, err)
		}
		// HEAD commit the patches below are diffed against
		sha, err := runOne(runner, repo.Path, "rev-parse", "HEAD")
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", repo.Path, err)
		}

		// paths with an unresolved merge conflict (UU) — excluded from the
		// diffs below and captured verbatim instead
		conflictedOut, err := runOne(runner, repo.Path, "diff", "--name-only", "--diff-filter=U")
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", repo.Path, err)
		}
		conflicted := splitLines(conflictedOut)

		rs := store.RepoSnapshot{
			Path:    repo.Path,
			Branch:  strings.TrimSpace(branch),
			BaseSHA: strings.TrimSpace(sha),
		}

		// staged changes (index vs HEAD) — reproduces A/M/D staged codes
		staged, err := runOne(runner, repo.Path, diffArgs("--cached", conflicted)...)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", repo.Path, err)
		}
		if staged != "" {
			rs.StagedPatch = store.SnapshotStagedPatchPath(issue.Workspace, issue.Hash, hash, repo.Name)
			chain.Add(&WriteFileCmd{Path: rs.StagedPatch, Data: []byte(staged)})
		}

		// unstaged changes (worktree vs index) — reproduces M/D unstaged codes
		unstaged, err := runOne(runner, repo.Path, diffArgs("", conflicted)...)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", repo.Path, err)
		}
		if unstaged != "" {
			rs.UnstagedPatch = store.SnapshotUnstagedPatchPath(issue.Workspace, issue.Hash, hash, repo.Name)
			chain.Add(&WriteFileCmd{Path: rs.UnstagedPatch, Data: []byte(unstaged)})
		}

		// untracked, non-ignored files (??) — can't be diffed, copied verbatim
		untrackedOut, err := runOne(runner, repo.Path, "ls-files", "--others", "--exclude-standard")
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", repo.Path, err)
		}
		for _, rel := range splitLines(untrackedOut) {
			rs.UntrackedDir = store.SnapshotUntrackedDir(issue.Workspace, issue.Hash, hash, repo.Name)
			chain.Add(&CopyFileCmd{
				Src: filepath.Join(repo.Path, rel),
				Dst: filepath.Join(rs.UntrackedDir, rel),
			})
			rs.UntrackedFiles = append(rs.UntrackedFiles, rel)
		}

		for _, rel := range conflicted {
			rs.ConflictedDir = store.SnapshotConflictedDir(issue.Workspace, issue.Hash, hash, repo.Name)
			chain.Add(&CopyFileCmd{
				Src: filepath.Join(repo.Path, rel),
				Dst: filepath.Join(rs.ConflictedDir, rel),
			})
			rs.ConflictedFiles = append(rs.ConflictedFiles, rel)
		}

		cleanups = append(cleanups, cleanupStep{
			repoPath:     repo.Path,
			hadConflicts: len(rs.ConflictedFiles) > 0,
			hadUntracked: len(rs.UntrackedFiles) > 0,
		})
		snap.Repos = append(snap.Repos, rs)
	}

	metaData, err := yaml.Marshal(snap)
	if err != nil {
		return nil, nil, err
	}
	chain.Add(&WriteFileCmd{Path: store.SnapshotMetaPath(issue.Workspace, issue.Hash, hash), Data: metaData})

	if dryRun {
		plan := chain.Plan()
		for _, c := range cleanups {
			if !clean {
				continue
			}
			if c.hadConflicts {
				plan = append(plan, fmt.Sprintf("git merge --abort (%s)", c.repoPath))
			} else {
				plan = append(plan, fmt.Sprintf("git checkout -- . (%s)", c.repoPath))
			}
			if c.hadUntracked {
				plan = append(plan, fmt.Sprintf("git clean -fd (%s)", c.repoPath))
			}
		}
		return snap, plan, nil
	}

	if err := chain.Run(); err != nil {
		return nil, nil, err
	}

	if clean {
		for _, c := range cleanups {
			if c.hadConflicts {
				// Aborting is the only way to leave no unmerged paths behind;
				// this discards the in-progress merge entirely (its content
				// is already captured above).
				if _, err := runOne(runner, c.repoPath, "merge", "--abort"); err != nil {
					return nil, nil, fmt.Errorf("%s: --clean failed: %w", c.repoPath, err)
				}
				// discard tracked changes, back to HEAD
			} else if _, err := runOne(runner, c.repoPath, "checkout", "--", "."); err != nil {
				return nil, nil, fmt.Errorf("%s: --clean failed: %w", c.repoPath, err)
			}
			if c.hadUntracked {
				// remove untracked files/dirs (already captured above)
				if _, err := runOne(runner, c.repoPath, "clean", "-fd"); err != nil {
					return nil, nil, fmt.Errorf("%s: --clean failed: %w", c.repoPath, err)
				}
			}
		}
	}

	return snap, nil, nil
}

func diffArgs(staged string, excludePaths []string) []string {
	args := []string{"diff", "--binary"}
	if staged != "" {
		args = append(args, staged)
	}
	if len(excludePaths) > 0 {
		args = append(args, "--", ".")
		for _, p := range excludePaths {
			args = append(args, ":(exclude)"+p)
		}
	}
	return args
}

func splitLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// runOne runs a git command in a single repo and returns its stdout.
func runOne(runner *git.Runner, repoPath string, args ...string) (string, error) {
	results := runner.Run([]string{repoPath}, args...)
	r := results[0]
	if r.Err != nil {
		return "", r.Err
	}
	return r.Stdout, nil
}
