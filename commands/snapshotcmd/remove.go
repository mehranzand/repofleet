package snapshotcmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <snapshot-hash>",
		Short: "Remove a specific snapshot by hash",
		Long:  "Search all issues in the workspace for the given snapshot hash and delete it.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name
			snapHash := args[0]

			allSnaps, err := store.ListSnapshots(ws)
			if err != nil {
				return err
			}
			var snap *store.Snapshot
			for _, s := range allSnaps {
				if s.Hash == snapHash {
					snap = s
					break
				}
			}
			if snap == nil {
				return fmt.Errorf("snapshot %q not found", snapHash)
			}
			if !util.Confirm(fmt.Sprintf("Remove snapshot %s (issue #%s)?", snap.Hash, snap.IssueID)) {
				return nil
			}
			if err := store.DeleteSnapshot(ws, snap.IssueHash, snapHash); err != nil {
				return err
			}
			fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), snap.Hash)
			return nil
		},
	}

	return cmd
}
