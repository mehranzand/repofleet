package workspacecmd

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a workspace",
		Long:  "Remove a workspace along with all its issues and snapshots. Prompts for confirmation, showing how many repos, issues, and snapshots will be deleted. Repositories on disk are not affected — only RepoFleet metadata is deleted.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if name == f.Workspace.Name {
				return fmt.Errorf("cannot remove the active workspace %q — switch to another workspace first", name)
			}

			ws, err := store.LoadWorkspace(name)
			if err != nil {
				return err
			}
			if ws == nil {
				return fmt.Errorf("workspace %q not found", name)
			}

			issues, err := store.LoadIssuesForWorkspace(name, "")
			if err != nil {
				return err
			}
			snaps, err := store.ListSnapshots(name)
			if err != nil {
				return err
			}

			var scope []string
			if n := len(ws.Repos); n > 0 {
				scope = append(scope, fmt.Sprintf("%d repo(s)", n))
			}
			if n := len(issues); n > 0 {
				scope = append(scope, fmt.Sprintf("%d issue(s)", n))
			}
			if n := len(snaps); n > 0 {
				scope = append(scope, fmt.Sprintf("%d snapshot(s)", n))
			}

			msg := fmt.Sprintf("Remove workspace %q?", name)
			if len(scope) > 0 {
				msg += " This will delete " + strings.Join(scope, ", ") + "."
			}
			if !util.Confirm(msg) {
				return nil
			}

			if err := store.DeleteWorkspace(name); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Removed workspace %q", name)))
			return nil
		},
	}
}
