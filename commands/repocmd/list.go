package repocmd

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List repositories in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace

			fmt.Fprintf(f.IO.Out, "%s %s\n\n", iostreams.Dim("Repos in workspace"), iostreams.Cyan(ws.Name))

			if len(ws.Repos) == 0 {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim("No repositories — add one with: rf repo add <path>"))
				return nil
			}

			iostreams.PrintRepos(f.IO.Out, ws.Repos)
			return nil
		},
	}
}
