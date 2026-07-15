package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all issues in the current workspace",
		Long:  "Show a detailed table of every issue in the current workspace, including archived ones. The active issue is marked with *.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name

			issues, err := store.LoadIssuesForWorkspace(ws, "")
			if err != nil {
				return err
			}
			if len(issues) == 0 {
				return fmt.Errorf("no issues found — create one with: rf issue create <id>")
			}

			iostreams.PrintIssues(f.IO.Out, issues, store.CurrentIssueHash(ws))
			return nil
		},
	}
}
