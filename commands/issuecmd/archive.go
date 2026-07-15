package issuecmd

import (
	"errors"
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newArchiveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <query>",
		Short: "Archive a completed issue using its ID, name, or hash",
		Long:  "Archive an active issue in the current workspace. Only searches active issues — errors if it's already archived.",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			ctx, err := store.FindIssueByKeyword(f.Workspace.Name, args[0], store.IssueStatusActive)
			if err != nil {
				var ambigErr *store.AmbiguousIssueError
				if errors.As(err, &ambigErr) {
					return err
				}
				return fmt.Errorf("issue %q not found in workspace %q", args[0], f.Workspace.Name)
			}

			if ctx.Status == store.IssueStatusArchived {
				return fmt.Errorf("issue %q is already archived", ctx.ID)
			}

			ctx.Status = store.IssueStatusArchived
			if err := ctx.Save(); err != nil {
				return err
			}

			if store.CurrentIssueHash(f.Workspace.Name) == ctx.Hash {
				_ = store.SetCurrentIssueHash(f.Workspace.Name, "")
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf("Archived issue %q", ctx.ID)))
			return nil
		},
	}
}
