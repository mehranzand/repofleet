package issuecmd

import (
	"fmt"
	"os"
	"path/filepath"
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
			base, _ := os.UserConfigDir()
			dir := filepath.Join(base, "repofleet", "issues")

			entries, err := os.ReadDir(dir)
			if os.IsNotExist(err) {
				fmt.Fprintln(f.IO.Out, iostreams.Dim("No issues found."))
				return nil
			}
			if err != nil {
				return err
			}

			currentID := store.CurrentIssueID()

			t := iostreams.NewTable()
			t.AddField("Id", iostreams.Dim)
			t.AddField("Status", iostreams.Dim)
			t.AddField("Branch", iostreams.Dim)
			t.AddField("Repos", iostreams.Dim)
			t.EndRow()

			found := false
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".yaml") {
					continue
				}
				id := strings.TrimSuffix(e.Name(), ".yaml")
				ctx, err := store.LoadIssue(id)
				if err != nil {
					continue
				}
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
				fmt.Fprintln(f.IO.Out, iostreams.Dim("No issues found."))
				return nil
			}

			t.Render(f.IO.Out)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showArchived, "all", "a", false, "include archived issues")
	return cmd
}
