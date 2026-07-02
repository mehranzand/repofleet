package root

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/commands/gitcmd"
	"github.com/mehranzand/repofleet/commands/issuecmd"
	"github.com/mehranzand/repofleet/commands/repocmd"
	"github.com/mehranzand/repofleet/commands/workspacecmd"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRootCmd(appVersion string) *cobra.Command {
	f, err := factory.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %s\n", err)
		os.Exit(1)
	}

	cobra.AddTemplateFunc("logo", func() string { return iostreams.Logo(appVersion) })
	cobra.AddTemplateFunc("green", iostreams.Green)
	cobra.AddTemplateFunc("cyan", iostreams.Cyan)
	cobra.AddTemplateFunc("bold", iostreams.Bold)
	cobra.AddTemplateFunc("dim", iostreams.Dim)
	cobra.AddTemplateFunc("rpadColor", func(s string, padding int) string {
		padded := s + strings.Repeat(" ", padding-len(s))
		return iostreams.Cyan(padded)
	})
	cobra.AddTemplateFunc("colorFlags", func(flags *pflag.FlagSet) string {
		return iostreams.ColorizeFlags(flags.FlagUsages())
	})
	cobra.AddTemplateFunc("availableCmds", func(cmds []*cobra.Command) []*cobra.Command {
		available := make([]*cobra.Command, 0, len(cmds))
		for _, c := range cmds {
			if c.IsAvailableCommand() {
				available = append(available, c)
			}
		}
		return available
	})

	binaryName := filepath.Base(os.Args[0])

	cmd := &cobra.Command{
		Use:          binaryName,
		Short:        "Multi-repo Git workflow manager",
		Long:         "",
		Version:      appVersion,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cobra.EnableCommandSorting = false

	cmd.SetVersionTemplate(iostreams.Logo(appVersion) + "\n\n")

	cmd.SetHelpTemplate(
		"\n" +
		`{{logo}}` + "\n\n" +

			`{{with .Long}}` +
			`{{. | trimRightSpace | dim}}` + "\n\n" +
			`{{end}}` +

			`{{green "Usage:"}}` + "\n" +
			`  {{.UseLine}}{{if .HasAvailableSubCommands}} [command]{{end}}` + "\n\n" +

			`{{if .HasAvailableSubCommands}}` +
			`{{green "Commands:"}}` + "\n" +
			`{{range availableCmds .Commands}}` +
			`  {{rpadColor .Name $.NamePadding}} {{.Short}}` + "\n" +
			`{{end}}` + "\n" +
			`{{end}}` +

			`{{if .HasAvailableLocalFlags}}` +
			`{{green "Options:"}}` + "\n" +
			`{{colorFlags .LocalFlags | trimRightSpace}}` + "\n" +
			`{{end}}` +

			"\n",
	)

	cmd.AddCommand(workspacecmd.NewCmd(f))
	cmd.AddCommand(repocmd.NewCmd(f))
	cmd.AddCommand(issuecmd.NewCmd(f))
	cmd.AddCommand(gitcmd.NewCmd(f))

	return cmd
}
