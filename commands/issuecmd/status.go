package issuecmd

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
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

			fmt.Fprintf(f.IO.Out, "%s %s   %s %s\n\n",
				iostreams.Dim("Issue:"), iostreams.Cyan(ctx.ID),
				iostreams.Dim("Branch:"), iostreams.Cyan(ctx.BranchSlug),
			)

			paths := repoPaths(ctx.Repos)
			branchResults := f.GitRunner.Run(paths, "rev-parse", "--abbrev-ref", "HEAD")
			statusResults := f.GitRunner.Run(paths, "status", "--short")

			t := iostreams.NewTable()
			t.AddField("Repo", iostreams.Dim)
			t.AddField("Branch", iostreams.Dim)
			t.AddField("Changes", iostreams.Dim)
			t.EndRow()

			for i, r := range ctx.Repos {
				branch := "?"
				if branchResults[i].Err == nil {
					branch = strings.TrimSpace(branchResults[i].Stdout)
				}
				changes, changesColor := "clean", iostreams.Green
				if statusResults[i].Err == nil {
					if lines := strings.TrimSpace(statusResults[i].Stdout); lines != "" {
						changes = fmt.Sprintf("%d change(s)", len(strings.Split(lines, "\n")))
						changesColor = iostreams.Cyan
					}
				}
				t.AddField(r.Name, iostreams.Cyan)
				t.AddField(branch, iostreams.Dim)
				t.AddField(changes, changesColor)
				t.EndRow()
			}

			t.Render(f.IO.Out)
			return nil
		},
	}
}
