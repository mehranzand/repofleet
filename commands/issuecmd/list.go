package issuecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
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
				fmt.Fprintln(f.IO.Out, "No issues found.")
				return nil
			}
			if err != nil {
				return err
			}

			currentID := store.CurrentIssueID()

			fmt.Fprintf(f.IO.Out, "%-20s %-12s %-30s %s\n", "ID", "STATUS", "BRANCH", "REPOS")
			fmt.Fprintf(f.IO.Out, "%-20s %-12s %-30s %s\n", "--", "------", "------", "-----")

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

				active := ""
				if ctx.ID == currentID {
					active = " *"
				}

				names := make([]string, len(ctx.Repos))
				for i, r := range ctx.Repos {
					names[i] = r.Name
				}

				fmt.Fprintf(f.IO.Out, "%-20s %-12s %-30s %s%s\n",
					ctx.ID,
					string(ctx.Status),
					ctx.BranchSlug,
					strings.Join(names, ", "),
					active,
				)
				found = true
			}

			if !found {
				fmt.Fprintln(f.IO.Out, "No issues found.")
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showArchived, "all", "a", false, "include archived issues")
	return cmd
}
