package issuecmd

import (
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issue contexts across repos",
		Long:  "Create, switch, sync, push, and archive issue contexts across multiple repositories.",
	}
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newSwitchCmd(f))
	cmd.AddCommand(newSyncCmd(f))
	cmd.AddCommand(newPushCmd(f))
	cmd.AddCommand(newStatusCmd(f))
	cmd.AddCommand(newArchiveCmd(f))
	return cmd
}
