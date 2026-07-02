package workspacecmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/spf13/cobra"
)

var validTokens = map[string]bool{
	"{workspace}":   true,
	"{issue}":       true,
	"{name}":        true,
	"{description}": true,
	"{repo}":        true,
}

var tokenRe = regexp.MustCompile(`\{[^}]+\}`)

func validateBranchPattern(pattern string) error {
	if strings.TrimSpace(pattern) == "" {
		return fmt.Errorf("branch pattern cannot be empty")
	}
	for _, token := range tokenRe.FindAllString(pattern, -1) {
		if !validTokens[token] {
			return fmt.Errorf("unknown token %s — valid tokens: {workspace}, {issue}, {name}, {description}, {repo}", token)
		}
	}
	return nil
}

func newConfigCmd(f *factory.Factory) *cobra.Command {
	var branchPattern string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or update workspace configuration",
		Long: "View or update configuration for the current workspace.\n\n" +
			"Branch pattern tokens:\n" +
			"  {workspace}    Workspace.Name\n" +
			"  {issue}        Issue.ID\n" +
			"  {name}         Issue.Name\n" +
			"  {description}  Issue.ShortDescription\n" +
			"  {repo}         Repo.Name\n\n" +
			"Example: feat/{issue}-{description}",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace

			if cmd.Flags().Changed("branch-pattern") {
				if err := validateBranchPattern(branchPattern); err != nil {
					return err
				}
				ws.BranchPattern = branchPattern
				if err := ws.Save(); err != nil {
					return err
				}
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf("Branch pattern set to %q", branchPattern)))
				return nil
			}

			// show current config
			fmt.Fprintf(f.IO.Out, "%s  %s\n", iostreams.Dim("Workspace:"), iostreams.Cyan(ws.Name))

			if ws.BranchPattern != "" {
				fmt.Fprintf(f.IO.Out, "%s  %s\n", iostreams.Dim("Branch pattern:"), iostreams.Cyan(ws.BranchPattern))
			} else {
				fmt.Fprintf(f.IO.Out, "%s  %s\n", iostreams.Dim("Branch pattern:"), iostreams.Dim("not set (uses slugified issue id)"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&branchPattern, "branch-pattern", "", "branch naming pattern, e.g. feat/{issue}-{description}")
	return cmd
}
