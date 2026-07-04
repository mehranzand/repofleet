package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove an issue context by ID",
		Long:  "Remove an issue context from the current workspace. Does not delete any git branches.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Settings.CurrentWorkspace
			id := args[0]

			issue, err := store.LoadIssueByIDOrName(ws, id)
			if err != nil {
				return fmt.Errorf("issue %q not found in workspace %q", id, ws)
			}
			if issue.Workspace != ws {
				return fmt.Errorf("issue %q belongs to workspace %q, not %q", issue.ID, issue.Workspace, ws)
			}

			if err := store.DeleteIssue(ws, issue.ID); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf("Issue %q removed", issue.ID)))
			return nil
		},
	}
}
