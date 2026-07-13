package snapshotcmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/snapshot"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util"
	"github.com/spf13/cobra"
)

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var clean bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "create [issue-id]",
		Short: "Save the current uncommitted diff of every repo in an issue",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var issue *store.Issue
			var err error
			if len(args) == 1 {
				issue, err = store.FindIssueByKeyword(f.Workspace.Name, args[0], store.IssueStatusActive)
				if err != nil {
					return err
				}
			} else {
				issueHash := store.CurrentIssueHash(f.Workspace.Name)
				if issueHash == "" {
					return fmt.Errorf("no active issue — switch to one with: rf issue switch <id>")
				}
				issue, err = store.LoadIssueByHash(f.Workspace.Name, issueHash)
				if err != nil {
					return err
				}
				if !util.Confirm(fmt.Sprintf("Create snapshot for current issue #%s?", issue.ID)) {
					return nil
				}
			}

			verb := "Capturing"
			if dryRun {
				verb = "Dry run: would capture"
			}
			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("%s snapshot for issue %q (%d repo(s))...", verb, issue.ID, len(issue.Repos))))

			snap, plan, err := snapshot.Create(f.GitRunner, issue, clean, dryRun)
			if err != nil {
				return err
			}

			if dryRun {
				for _, step := range plan {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Dim("·"), step)
				}
				fmt.Fprintf(f.IO.Out, "\n%s\n", iostreams.Dim("Dry run — nothing written"))
				return nil
			}

			for _, rs := range snap.Repos {
				changed := rs.StagedPatch != "" || rs.UnstagedPatch != "" || len(rs.UntrackedFiles) > 0 || len(rs.ConflictedFiles) > 0
				if !changed {
					fmt.Fprintf(f.IO.Out, "  %s %s: no changes\n", iostreams.Green("✓"), rs.Path)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s: saved\n", iostreams.Green("✓"), rs.Path)
				}
			}

			fmt.Fprintf(f.IO.Out, "\n%s\n", iostreams.Success(fmt.Sprintf("Snapshot %s saved for issue %q", snap.Hash, issue.ID)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&clean, "clean", false, "reset each repo's working tree to HEAD after saving its diff")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be captured without writing anything")
	return cmd
}
