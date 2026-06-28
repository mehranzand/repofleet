package repo

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/store"
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument: <path>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch forge {
			case "github", "gitlab":
			default:
				return fmt.Errorf("invalid forge %q: must be github or gitlab", forge)
			}

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

			remoteURL := url
			if remoteURL == "" {
				out, err := exec.Command("git", "-C", absPath, "remote", "get-url", "origin").Output()
				if err == nil {
					remoteURL = strings.TrimSpace(string(out))
				}
			}

			repo := store.Repo{
				Name:  repoName,
				Path:  absPath,
				Forge: forge,
				URL:   remoteURL,
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
	cmd.Flags().StringVarP(&forge, "forge", "f", "github", `forge type: "github" or "gitlab"`)
	cmd.Flags().StringVarP(&url, "url", "u", "", "remote URL of the repository")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "target workspace (default: current)")

	return cmd
}
