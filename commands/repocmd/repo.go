package repocmd

import (
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories in a workspace",
		Long:  "Add, remove, and list repositories grouped into workspaces.",
	}
	cmd.AddCommand(newAddCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	return cmd
}
