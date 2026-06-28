package issuecmd

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newStatusCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dashboard for all repos in the current issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			id := store.CurrentIssueID()
			if id == "" {
				return fmt.Errorf("no active issue — switch to one with: repofleet issue switch <id>")
			}

			ctx, err := store.LoadIssue(id)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "Issue: %s   Branch: %s\n\n", ctx.ID, ctx.BranchSlug)
			fmt.Fprintf(f.IO.Out, "%-30s %-25s %s\n", "REPO", "BRANCH", "CHANGES")
			fmt.Fprintf(f.IO.Out, "%-30s %-25s %s\n", "----", "------", "-------")

			paths := repoPaths(ctx.Repos)

			// get current branch per repo
			branchResults := f.GitRunner.Run(paths, "rev-parse", "--abbrev-ref", "HEAD")
			// get short status per repo
			statusResults := f.GitRunner.Run(paths, "status", "--short")

			for i, r := range ctx.Repos {
				branch := "?"
				if branchResults[i].Err == nil {
					branch = strings.TrimSpace(branchResults[i].Stdout)
				}
				changes := "clean"
				if statusResults[i].Err == nil {
					lines := strings.TrimSpace(statusResults[i].Stdout)
					if lines != "" {
						changes = fmt.Sprintf("%d change(s)", len(strings.Split(lines, "\n")))
					}
				}
				fmt.Fprintf(f.IO.Out, "%-30s %-25s %s\n", r.Name, branch, changes)
			}
			return nil
		},
	}
}
