package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newArchiveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <issue-id>",
		Short: "Archive a completed issue workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := store.LoadIssue(args[0])
			if err != nil {
				return fmt.Errorf("issue %q not found", args[0])
			}

			if ctx.Workspace != f.Settings.CurrentWorkspace {
				return fmt.Errorf("issue %q belongs to workspace %q, not %q", ctx.ID, ctx.Workspace, f.Settings.CurrentWorkspace)
			}

			ctx.Status = store.IssueStatusArchived
			if err := ctx.Save(); err != nil {
				return err
			}

			if store.CurrentIssueID(f.Settings.CurrentWorkspace) == args[0] {
				_ = store.SetCurrentIssue(f.Settings.CurrentWorkspace, "")
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf("Archived issue %q", args[0])))
			return nil
		},
	}
}
