package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newPushCmd(f *factory.Factory) *cobra.Command {
	var forceWithLease bool

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push all issue branches to their remotes",
		RunE: func(cmd *cobra.Command, args []string) error {
			id := store.CurrentIssueID(f.Settings.CurrentWorkspace)
			if id == "" {
				return fmt.Errorf("no active issue — switch to one with: repofleet issue switch <id>")
			}

			ctx, err := store.LoadIssue(id)
			if err != nil {
				return err
			}

			paths := repoPaths(ctx.Repos)
			pushArgs := []string{"push", "--set-upstream", "origin", ctx.BranchSlug}
			if forceWithLease {
				pushArgs = append(pushArgs, "--force-with-lease")
			}

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("Pushing branch %q in %d repo(s)...", ctx.BranchSlug, len(paths))))
			results := f.GitRunner.Run(paths, pushArgs...)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s: %s\n", iostreams.Red("✗"), r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), r.RepoPath)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&forceWithLease, "force-with-lease", false, "push with --force-with-lease")
	return cmd
}
