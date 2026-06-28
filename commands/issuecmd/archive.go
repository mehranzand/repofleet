package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	issueCtx "github.com/mehranzand/repofleet/internal/issue"
	"github.com/spf13/cobra"
)

func newArchiveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <issue-id>",
		Short: "Archive a completed issue workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := issueCtx.Load(args[0])
			if err != nil {
				return fmt.Errorf("issue %q not found", args[0])
			}

			ctx.Status = issueCtx.StatusArchived
			if err := ctx.Save(); err != nil {
				return err
			}

			// clear current if it was active
			if issueCtx.CurrentID() == args[0] {
				_ = issueCtx.SetCurrent("")
			}

			fmt.Fprintf(f.IO.Out, "Archived issue %q\n", args[0])
			return nil
		},
	}
}
