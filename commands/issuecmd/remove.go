package issuecmd

import (
	"errors"
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	var showArchived bool

	cmd := &cobra.Command{
		Use:   "remove <query>",
		Short: "Remove an issue context using its ID, name, or hash.",
		Long:  "Remove an issue context from the current workspace. Does not delete any git branches.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name
			id := args[0]

			status := store.IssueStatusActive
			if showArchived {
				status = ""
			}

			issue, err := store.FindIssueByKeyword(ws, id, status)
			if err != nil {
				var ambigErr *store.AmbiguousIssueError
				if errors.As(err, &ambigErr) {
					return err
				}
				return fmt.Errorf("issue %q not found in workspace %q", id, ws)
			}

			if err := store.DeleteIssue(ws, issue.Hash); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf("Issue %q removed", issue.ID)))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showArchived, "archived", "A", false, "also search archived issues")
	return cmd
}
