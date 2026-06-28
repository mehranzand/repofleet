package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newAddCmd(f *factory.Factory) *cobra.Command {
	var name string
	var forge string
	var url string
	var workspace string

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add a repository to a workspace",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("%s", iostreams.Dim("missing required argument: <path>"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch forge {
			case "github", "gitlab":
			default:
				return fmt.Errorf("%s", iostreams.Dim(fmt.Sprintf("invalid forge %q: must be github or gitlab", forge)))
			}

			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("%s", iostreams.Dim("path does not exist: "+absPath))
			} else if err != nil {
				return fmt.Errorf("%s", iostreams.Dim("cannot access path: "+absPath))
			}

			if err := exec.Command("git", "-C", absPath, "rev-parse", "--git-dir").Run(); err != nil {
				return fmt.Errorf("%s", iostreams.Dim("not a git repository: "+absPath))
			}

			ws := workspace
			if ws == "" {
				ws = f.Config.CurrentWorkspace
			}

			repoName := name
			if repoName == "" {
				repoName = filepath.Base(absPath)
			}

			for _, r := range f.Config.WorkspaceRepos(ws) {
				if r.Name == repoName {
					return fmt.Errorf("%s", iostreams.Dim(fmt.Sprintf("a repo named %q already exists in workspace %q", repoName, ws)))
				}
				if r.Path == absPath {
					return fmt.Errorf("%s", iostreams.Dim(fmt.Sprintf("path %s is already added to workspace %q as %q", absPath, ws, r.Name)))
				}
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

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Added %q to workspace %q", repoName, ws)))

			target := f.Config.CurrentWS()
			if workspace != "" {
				for i := range f.Config.Workspaces {
					if f.Config.Workspaces[i].Name == workspace {
						target = &f.Config.Workspaces[i]
						break
					}
				}
			}
			iostreams.PrintRepos(f.IO.Out, target.Repos)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "name for the repo (default: directory basename)")
	cmd.Flags().StringVarP(&forge, "forge", "f", "github", `forge type: "github" or "gitlab"`)
	cmd.Flags().StringVarP(&url, "url", "u", "", "remote URL of the repository")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "target workspace (default: current)")

	return cmd
}
