package snapshot

import (
	"path/filepath"

	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util/git"
)

func Restore(runner *git.Runner, snap *store.Snapshot, dryRun bool) ([]string, error) {
	chain := NewChain()

	for _, rs := range snap.Repos {
		if rs.StagedPatch != "" {
			chain.Add(&GitCommand{
				Runner: runner, RepoPath: rs.Path,
				Args:     []string{"apply", "--binary", "--cached", rs.StagedPatch},
				UndoArgs: []string{"apply", "--binary", "--cached", "-R", rs.StagedPatch},
			})
			chain.Add(&GitCommand{
				Runner: runner, RepoPath: rs.Path,
				Args:     []string{"apply", "--binary", rs.StagedPatch},
				UndoArgs: []string{"apply", "--binary", "-R", rs.StagedPatch},
			})
		}

		if rs.UnstagedPatch != "" {
			chain.Add(&GitCommand{
				Runner: runner, RepoPath: rs.Path,
				Args:     []string{"apply", "--binary", rs.UnstagedPatch},
				UndoArgs: []string{"apply", "--binary", "-R", rs.UnstagedPatch},
			})
		}

		for _, rel := range rs.UntrackedFiles {
			chain.Add(&CopyFileCmd{
				Src: filepath.Join(rs.UntrackedDir, rel),
				Dst: filepath.Join(rs.Path, rel),
			})
		}

		for _, rel := range rs.ConflictedFiles {
			chain.Add(&CopyFileCmd{
				Src: filepath.Join(rs.ConflictedDir, rel),
				Dst: filepath.Join(rs.Path, rel),
			})
		}
	}

	if dryRun {
		return chain.Plan(), nil
	}
	return nil, chain.Run()
}
