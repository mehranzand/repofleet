package repocmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var workspace string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories in a workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			target := f.Workspace
			if workspace != "" && workspace != f.Settings.CurrentWorkspace {
				var err error
				target, err = store.LoadWorkspace(workspace)
				if err != nil {
					return err
				}
			}

			if len(target.Repos) == 0 {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No repositories in workspace "+target.Name))
				return nil
			}

			iostreams.PrintRepos(f.IO.Out, target.Repos)
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "workspace to list (default: current)")
	return cmd
}
