package snapshotcmd

import (
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Capture and restore an issue's uncommitted changes",
		Long:  "Save each repo's uncommitted diff to a patch file and restore it later, without using git stash or a local commit.",
	}
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newRestoreCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	cmd.AddCommand(newPruneCmd(f))
	return cmd
}
