package workspacecmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if name == f.Settings.CurrentWorkspace {
				return fmt.Errorf("cannot remove the active workspace %q — switch to another workspace first", name)
			}

			if err := store.DeleteWorkspace(name); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Removed workspace %q", name)))
			return nil
		},
	}
}
