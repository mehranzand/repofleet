package workspacecmd

import (
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces",
		Long:  "Switch between workspaces, add or remove one.",
	}
	cmd.AddCommand(newSwitchCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	cmd.AddCommand(newConfigCmd(f))
	return cmd
}
