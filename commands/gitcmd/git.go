package gitcmd

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "git [git args...]",
		Short:              "Run a git command across all repos in the current workspace",
		Example:            "  repofleet git status\n  repofleet git fetch --all\n  repofleet git checkout main",
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace
			if len(ws.Repos) == 0 {
				return fmt.Errorf("no repos in workspace %q — add one with: repofleet repo add <path>", ws.Name)
			}

			paths := make([]string, len(ws.Repos))
			for i, r := range ws.Repos {
				paths[i] = r.Path
			}

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("Running: git %s  [%d repos]", strings.Join(args, " "), len(paths))))

			results := f.GitRunner.Run(paths, args...)
			for _, r := range results {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Cyan("── "+r.RepoPath))
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "   %s %s\n", iostreams.Red("✗"), r.Err)
				} else if strings.TrimSpace(r.Stdout) != "" {
					for _, line := range strings.Split(strings.TrimRight(r.Stdout, "\n"), "\n") {
						fmt.Fprintf(f.IO.Out, "   %s\n", line)
					}
				} else {
					fmt.Fprintf(f.IO.Out, "   %s\n", iostreams.Dim("(no output)"))
				}
				fmt.Fprintln(f.IO.Out)
			}
			return nil
		},
	}
	return cmd
}
