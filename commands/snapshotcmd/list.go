package snapshotcmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved snapshots in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			snaps, err := store.ListSnapshots(f.Workspace.Name)
			if err != nil {
				return err
			}
			if len(snaps) == 0 {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No snapshots. create one with: rf snapshot create <issue-id>"))
				return nil
			}
			iostreams.PrintSnapshots(f.IO.Out, snaps)
			return nil
		},
	}
}
