package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newSyncCmd(f *factory.Factory) *cobra.Command {
	var rebase bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Fetch and pull/rebase all repos for the current issue",
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
			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("Fetching %d repo(s)...", len(paths))))

			fetchResults := f.GitRunner.Run(paths, "fetch", "--all")
			for _, r := range fetchResults {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s: %s\n", iostreams.Red("✗"), r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), r.RepoPath)
				}
			}

			pullArgs := []string{"pull"}
			if rebase {
				pullArgs = append(pullArgs, "--rebase")
			}

			fmt.Fprintf(f.IO.Out, "\n%s\n\n", iostreams.Dim(fmt.Sprintf("Pulling %d repo(s)...", len(paths))))
			pullResults := f.GitRunner.Run(paths, pullArgs...)
			for _, r := range pullResults {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s: %s\n", iostreams.Red("✗"), r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), r.RepoPath)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&rebase, "rebase", "r", false, "use rebase instead of merge")
	return cmd
}
