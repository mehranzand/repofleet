package snapshotcmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/snapshot"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newRestoreCmd(f *factory.Factory) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "restore <issue-id> [snapshot-hash]",
		Short: "Re-apply a previously saved snapshot's diff to each repo",
		Long:  "Re-apply a previously saved snapshot's diff to each repo. Defaults to the most recently created snapshot for the issue; pass a specific snapshot hash (see rf snapshot list) to restore an older one.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue, err := store.FindIssueByKeyword(f.Workspace.Name, args[0], "")
			if err != nil {
				return err
			}

			var snap *store.Snapshot
			if len(args) == 2 {
				snap, err = store.LoadSnapshot(issue.Workspace, issue.Hash, args[1])
				if err != nil {
					return fmt.Errorf("no snapshot %q found for issue %q", args[1], issue.ID)
				}
			} else {
				snap, err = store.LatestSnapshot(issue.Workspace, issue.Hash)
				if err != nil {
					return err
				}
				if snap == nil {
					return fmt.Errorf("no snapshot found for issue %q", issue.ID)
				}
			}

			verb := "Restoring"
			if dryRun {
				verb = "Dry run: would restore"
			}
			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("%s snapshot for issue %q...", verb, issue.ID)))

			plan, err := snapshot.Restore(f.GitRunner, snap, dryRun)
			if err != nil {
				return err
			}

			if dryRun {
				for _, step := range plan {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Dim("·"), step)
				}
				fmt.Fprintf(f.IO.Out, "\n%s\n", iostreams.Dim("Dry run — nothing applied"))
				return nil
			}

			for _, rs := range snap.Repos {
				fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), rs.Path)
			}

			fmt.Fprintf(f.IO.Out, "\n%s\n", iostreams.Success(fmt.Sprintf("Snapshot restored for issue %q", issue.ID)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be applied without changing anything")
	return cmd
}
