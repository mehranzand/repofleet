package repocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util/giturl"
	"github.com/spf13/cobra"
)

func detectForge(remoteURL string) (store.RepoForge, bool) {
	u := strings.ToLower(remoteURL)
	if strings.Contains(u, "gitlab") {
		return store.RepoForgeGitLab, true
	}
	if strings.Contains(u, "github") {
		return store.RepoForgeGitHub, true
	}
	return "", false
}

func newAddCmd(f *factory.Factory) *cobra.Command {
	var name string
	var forge string
	var remote string

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add a repository to the current workspace",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument: <path>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			} else if err != nil {
				return fmt.Errorf("cannot access path: %s", absPath)
			}

			if err := exec.Command("git", "-C", absPath, "rev-parse", "--git-dir").Run(); err != nil {
				return fmt.Errorf("not a git repository: %s", absPath)
			}

			target := f.Workspace

			repoName := name
			if repoName == "" {
				repoName = filepath.Base(absPath)
			}

			for _, r := range target.Repos {
				if r.Name == repoName {
					return fmt.Errorf("a repo named %q already exists in workspace %q", repoName, target.Name)
				}
				if r.Path == absPath {
					return fmt.Errorf("path %s is already added to workspace %q as %q", absPath, target.Name, r.Name)
				}
			}

			remoteURL := remote
			if remoteURL == "" {
				out, err := exec.Command("git", "-C", absPath, "remote", "get-url", "origin").Output()
				if err == nil {
					remoteURL = strings.TrimSpace(string(out))
				}
			}
			remoteURL = giturl.Normalize(remoteURL)

			var resolvedForge store.RepoForge
			if forge != "" {
				switch store.RepoForge(forge) {
				case store.RepoForgeGitHub, store.RepoForgeGitLab:
					resolvedForge = store.RepoForge(forge)
				default:
					return fmt.Errorf("invalid --forge %q: must be github or gitlab", forge)
				}
			} else {
				detected, ok := detectForge(remoteURL)
				if !ok {
					return fmt.Errorf("could not detect forge from remote URL — use --forge github or --forge gitlab")
				}
				resolvedForge = detected
			}

			target.AddRepo(store.Repo{
				Name:   repoName,
				Path:   absPath,
				Forge:  resolvedForge,
				Remote: remoteURL,
			})
			if err := target.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Added %q to workspace %q", repoName, target.Name)))
			iostreams.PrintRepos(f.IO.Out, target.Repos)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "name for the repo (default: directory basename)")
	cmd.Flags().StringVarP(&forge, "forge", "f", "", "override forge type: github or gitlab (default: auto-detected from remote URL)")
	cmd.Flags().StringVarP(&remote, "remote", "u", "", "remote URL (default: git remote get-url origin)")

	return cmd
}
