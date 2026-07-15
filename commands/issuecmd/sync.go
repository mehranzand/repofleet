package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newSyncCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Fetch all remotes for every repo in the current issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			hash := store.CurrentIssueHash(f.Workspace.Name)
			if hash == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch <id>")
			}

			ctx, err := store.LoadIssueByHash(f.Workspace.Name, hash)
			if err != nil {
				return err
			}

			paths := repoPaths(ctx.Repos)
			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("Fetching %d repo(s)...", len(paths))))

			for _, r := range f.GitRunner.Run(paths, "fetch", "--all") {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s: %s\n", iostreams.Red("✗"), r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), r.RepoPath)
				}
			}
			return nil
		},
	}
}
