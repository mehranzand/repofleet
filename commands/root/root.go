package root

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/commands/gitcmd"
	"github.com/mehranzand/repofleet/commands/issuecmd"
	"github.com/mehranzand/repofleet/commands/repo"
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

	cobra.AddTemplateFunc("green", iostreams.Green)
	cobra.AddTemplateFunc("cyan", iostreams.Cyan)
	cobra.AddTemplateFunc("bold", iostreams.Bold)
	cobra.AddTemplateFunc("rpadColor", func(s string, padding int) string {
		padded := s + strings.Repeat(" ", padding-len(s))
		return iostreams.Cyan(padded)
	})
	cobra.AddTemplateFunc("colorFlags", func(flags *pflag.FlagSet) string {
		return iostreams.ColorizeFlags(flags.FlagUsages())
	})
	cmdOrder := []string{"repo", "issue", "git"}
	cobra.AddTemplateFunc("orderedCmds", func(cmds []*cobra.Command) []*cobra.Command {
		index := make(map[string]*cobra.Command, len(cmds))
		for _, c := range cmds {
			index[c.Name()] = c
		}
		ordered := make([]*cobra.Command, 0, len(cmds))
		for _, name := range cmdOrder {
			if c, ok := index[name]; ok && c.IsAvailableCommand() {
				ordered = append(ordered, c)
			}
		}
		return ordered
	})

	binaryName := filepath.Base(os.Args[0])

	cmd := &cobra.Command{
		Use:          binaryName,
		Short:        "Multi-repo Git workflow manager",
		Long:         "\n" + iostreams.Logo(appVersion),
		Version:      appVersion,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.SetVersionTemplate(iostreams.Logo(appVersion) + "\n\n")

	cmd.SetHelpTemplate(
		`{{with .Long}}` +
			`{{. | trimRightSpace}}` + "\n\n" +
			`{{end}}` +

			`{{bold "Usage:"}}` + "\n" +
			`  {{.UseLine}}{{if .HasAvailableSubCommands}} [command]{{end}}` + "\n\n" +

			`{{if .HasAvailableSubCommands}}` +
			`{{bold "Commands:"}}` + "\n" +
			`{{range orderedCmds .Commands}}` +
			`  {{rpadColor .Name $.NamePadding}} {{.Short}}` + "\n" +
			`{{end}}` + "\n" +
			`{{end}}` +

			`{{if .HasAvailableLocalFlags}}` +
			`{{bold "Options:"}}` + "\n" +
			`{{colorFlags .LocalFlags | trimRightSpace}}` + "\n" +
			`{{end}}` +

			"\n",
	)

	cmd.AddCommand(repo.NewCmd(f))
	cmd.AddCommand(issuecmd.NewCmd(f))
	cmd.AddCommand(gitcmd.NewCmd(f))

	return cmd
}
