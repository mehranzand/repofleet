package repocmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	var workspace string

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a repository from a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := f.Workspace
			if workspace != "" && workspace != f.Settings.CurrentWorkspace {
				var err error
				target, err = store.LoadWorkspace(workspace)
				if err != nil {
					return err
				}
			}

			if !target.RemoveRepo(args[0]) {
				return fmt.Errorf("repo %q not found in workspace %q", args[0], target.Name)
			}

			if err := target.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Removed %q from workspace %q", args[0], target.Name)))

			if len(target.Repos) == 0 {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No repositories in workspace "+target.Name))
			} else {
				iostreams.PrintRepos(f.IO.Out, target.Repos)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "target workspace (default: current)")
	return cmd
}
