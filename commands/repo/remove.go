package repo

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/spf13/cobra"
)

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	var workspace string

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a repository from a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := workspace
			if ws == "" {
				ws = f.Config.CurrentWorkspace
			}

			if !f.Config.RemoveRepo(ws, args[0]) {
				return fmt.Errorf("repo %q not found in workspace %q", args[0], ws)
			}

			if err := f.Config.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Removed %q from workspace %q", args[0], ws)))

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
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No repositories in workspace "+ws))
			} else {
				iostreams.PrintRepos(f.IO.Out, target.Repos)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "target workspace (default: current)")
	return cmd
}
