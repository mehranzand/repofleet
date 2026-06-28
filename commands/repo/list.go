package repo

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var workspace string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories in a workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := workspace
			if ws == "" {
				ws = f.Config.CurrentWorkspace
			}

			target := f.Config.CurrentWS()
			if workspace != "" {
				for i := range f.Config.Workspaces {
					if f.Config.Workspaces[i].Name == workspace {
						target = &f.Config.Workspaces[i]
						break
					}
				}
			}

			if len(target.Repos) == 0 {
				fmt.Fprintf(f.IO.Out, "No repositories in workspace %q\n", ws)
				return nil
			}

			fmt.Fprintf(f.IO.Out, "%-20s %-10s %s\n", "NAME", "FORGE", "PATH")
			fmt.Fprintf(f.IO.Out, "%-20s %-10s %s\n", "----", "-----", "----")
			for _, r := range target.Repos {
				fmt.Fprintf(f.IO.Out, "%-20s %-10s %s\n", r.Name, r.Forge, r.Path)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "workspace to list (default: current)")
	return cmd
}
