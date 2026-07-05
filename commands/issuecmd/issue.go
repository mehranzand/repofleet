package issuecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issue contexts across repos",
		Long:  "Create, switch, sync, and archive issue contexts across multiple repositories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			id := store.CurrentIssueID(f.Workspace.Name)
			if id != "" {
				issue, err := store.LoadIssue(id)
				repoCount := 0
				if err == nil {
					repoCount = len(issue.Repos)
				}
				fmt.Fprintf(f.IO.Out, "\n%s %s %s %s %s\n\n",
					iostreams.Dim("Current issue is"),
					iostreams.Cyan("#"+id),
					iostreams.Dim(fmt.Sprintf("(%d repos) in the", repoCount)),
					iostreams.BoldGreen(f.Workspace.Name),
					iostreams.Dim("workspace"),
				)
			}
			return nil
		},
	}
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newSwitchCmd(f))
	cmd.AddCommand(newSyncCmd(f))
	cmd.AddCommand(newStatusCmd(f))
	cmd.AddCommand(newRepoCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	cmd.AddCommand(newArchiveCmd(f))
	return cmd
}
