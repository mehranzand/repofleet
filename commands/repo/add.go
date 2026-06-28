package repo

import (
	"fmt"
	"path/filepath"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/config"
	"github.com/spf13/cobra"
)

func newAddCmd(f *factory.Factory) *cobra.Command {
	var name      string
	var forge     string
	var url       string
	var workspace string

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add a repository to a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			ws := workspace
			if ws == "" {
				ws = f.Config.CurrentWorkspace
			}

			repoName := name
			if repoName == "" {
				repoName = filepath.Base(absPath)
			}

			repo := config.Repo{
				Name:  repoName,
				Path:  absPath,
				Forge: forge,
				URL:   url,
			}

			f.Config.AddRepo(ws, repo)
			if err := f.Config.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "Added %q to workspace %q\n", repoName, ws)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "name for the repo (default: directory basename)")
	cmd.Flags().StringVarP(&forge, "forge", "f", "github", "forge type: github or gitlab")
	cmd.Flags().StringVarP(&url, "url", "u", "", "remote URL of the repository")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "target workspace (default: current)")

	return cmd
}
