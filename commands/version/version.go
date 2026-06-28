package version

import (
	"fmt"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/spf13/cobra"
)

func NewCmd(f *factory.Factory, appVersion string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the repofleet version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.IO.Out, "repofleet %s\n", appVersion)
			return nil
		},
	}
}
