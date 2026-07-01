package issuecmd

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var showArchived bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all issue contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Settings.CurrentWorkspace
			issues, err := store.LoadIssuesForWorkspace(ws)
			if err != nil {
				return err
			}

			currentID := store.CurrentIssueID(ws)

			t := iostreams.NewTable()
			t.AddField("Id", iostreams.Dim)
			t.AddField("Status", iostreams.Dim)
			t.AddField("Branch", iostreams.Dim)
			t.AddField("Repos", iostreams.Dim)
			t.EndRow()

			found := false
			for _, ctx := range issues {
				if !showArchived && ctx.Status == store.IssueStatusArchived {
					continue
				}

				names := make([]string, len(ctx.Repos))
				for i, r := range ctx.Repos {
					names[i] = r.Name
				}

				idColor := iostreams.Cyan
				if ctx.ID == currentID {
					idColor = iostreams.BoldGreen
				}
				statusColor := iostreams.Dim
				if ctx.Status == store.IssueStatusActive {
					statusColor = iostreams.Green
				}

				t.AddField(ctx.ID, idColor)
				t.AddField(string(ctx.Status), statusColor)
				t.AddField(ctx.BranchSlug, iostreams.Dim)
				t.AddField(strings.Join(names, ", "), iostreams.Dim)
				t.EndRow()
				found = true
			}

			if !found {
				fmt.Fprintln(f.IO.Out, iostreams.Dim("No issues in workspace "+ws))
				return nil
			}

			t.Render(f.IO.Out)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showArchived, "all", "a", false, "include archived issues")
	return cmd
}
