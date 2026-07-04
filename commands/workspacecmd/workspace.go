package workspacecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces",
		Long:  "Switch between workspaces, add or remove one.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			fmt.Fprintf(f.IO.Out, "\n%s %s %s\n\n", iostreams.Dim("Current workspace is"), iostreams.BoldGreen(f.Workspace.Name), iostreams.Dim(fmt.Sprintf("(%d repos)", len(f.Workspace.Repos))))
			return nil
		},
	}
	cmd.AddCommand(newSwitchCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	cmd.AddCommand(newConfigCmd(f))
	return cmd
}
