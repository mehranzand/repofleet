package snapshotcmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util"
	"github.com/spf13/cobra"
)

func newPruneCmd(f *factory.Factory) *cobra.Command {
	var all, force bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove all snapshots for the current issue (or entire workspace with --all)",
		Long:  "Delete every snapshot saved under the active issue. Use --all to remove snapshots across all issues in the workspace. Use --force to skip the confirmation prompt.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name

			if all {
				snaps, err := store.ListSnapshots(ws)
				if err != nil {
					return err
				}
				if len(snaps) == 0 {
					fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No snapshots to remove."))
					return nil
				}
				if !force && !util.Confirm(fmt.Sprintf("Remove all %d snapshot(s) in workspace %q?", len(snaps), ws)) {
					return nil
				}
				if err := store.DeleteSnapshotsForWorkspace(ws); err != nil {
					return err
				}
				for _, s := range snaps {
					fmt.Fprintf(f.IO.Out, "  %s %s %s\n", iostreams.Green("✓"), s.Hash, iostreams.Dim("#"+s.IssueID))
				}
				return nil
			}

			issueHash := store.CurrentIssueHash(ws)
			if issueHash == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch <id>")
			}
			issue, err := store.LoadIssueByHash(ws, issueHash)
			if err != nil {
				return err
			}
			isCurrent := issue.Hash == store.CurrentIssueHash(ws)
			issueLabel := "#" + issue.ID
			if isCurrent {
				issueLabel = "current issue #" + issue.ID
			}
			snaps, err := store.ListSnapshotsForIssue(ws, issueHash)
			if err != nil {
				return err
			}
			if len(snaps) == 0 {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No snapshots to remove for "+issueLabel+"."))
				return nil
			}
			if !force && !util.Confirm(fmt.Sprintf("Remove all %d snapshot(s) for %s?", len(snaps), issueLabel)) {
				return nil
			}
			if err := store.DeleteSnapshotsForIssue(ws, issueHash); err != nil {
				return err
			}
			for _, s := range snaps {
				fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), s.Hash)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "remove snapshots across all issues in the workspace")
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}
