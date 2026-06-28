package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newSwitchCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <issue-id>",
		Short: "Switch all repos to the issue branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := store.LoadIssue(args[0])
			if err != nil {
				return fmt.Errorf("issue %q not found — create it first with: repofleet issue create %s", args[0], args[0])
			}

			if err := store.SetCurrentIssue(ctx.ID); err != nil {
				return err
			}

			paths := repoPaths(ctx.Repos)
			fmt.Fprintf(f.IO.Out, "Switching %d repo(s) to branch %q...\n\n", len(paths), ctx.BranchSlug)

			results := f.GitRunner.Run(paths, "checkout", ctx.BranchSlug)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Err, "  x %s: %s\n", r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  ok %s\n", r.RepoPath)
				}
			}
			return nil
		},
	}
}
