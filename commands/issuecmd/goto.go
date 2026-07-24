package issuecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newGotoCmd(f *factory.Factory) *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:   "goto",
		Short: "Interactively select a repo, cd into it, and switch to the issue branch if needed",
		Long: "Interactively select a repo, cd into it, and switch to the issue branch if needed.\n\n" +
			"Repos whose path no longer exists on disk show a red ! marker and cannot be selected.",
		RunE: func(cmd *cobra.Command, args []string) error {
			hash := store.CurrentIssueHash(f.Workspace.Name)
			if hash == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch <id>")
			}

			ctx, err := store.LoadIssueByHash(f.Workspace.Name, hash)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "\n%s %s %s %s %s\n\n",
				iostreams.Dim("Current issue is"),
				iostreams.Cyan("#"+ctx.ID),
				iostreams.Dim(fmt.Sprintf("(%d repos) in the", len(ctx.Repos))),
				iostreams.BoldGreen(f.Workspace.Name),
				iostreams.Dim("workspace"),
			)

			cwd, _ := filepath.Abs(".")

			paths := repoPaths(ctx.Repos)
			branchResults := f.GitRunner.Run(paths, "rev-parse", "--abbrev-ref", "HEAD")

			items := make([]repoItem, len(ctx.Repos))
			maxName, maxBranch := 0, 0
			for i, r := range ctx.Repos {
				missing := checkRepoPath(r) != nil
				b := "?"
				if missing {
					b = "missing"
				} else if branchResults[i].Err == nil {
					b = strings.TrimSpace(branchResults[i].Stdout)
				}
				switchBranch := ""
				if !missing && b != ctx.BranchSlug {
					switchBranch = ctx.BranchSlug
				}
				absPath, _ := filepath.Abs(r.Path)
				items[i] = repoItem{
					Name:         r.Name,
					Branch:       b,
					SwitchBranch: switchBranch,
					Current:      absPath == cwd,
					Missing:      missing,
					path:         r.Path,
				}
				if len(r.Name) > maxName {
					maxName = len(r.Name)
				}
				if len(b) > maxBranch {
					maxBranch = len(b)
				}
			}
			for i := range items {
				items[i].Name   += strings.Repeat(" ", maxName-len(items[i].Name))
				items[i].Branch += strings.Repeat(" ", maxBranch-len(items[i].Branch))
			}

			// float current repo to top
			for i, item := range items {
				if item.Current && i > 0 {
					items = append([]repoItem{items[i]}, append(items[:i], items[i+1:]...)...)
					break
				}
			}

			templates := &promptui.SelectTemplates{
				Label:    `{{ "Go to repo:" | faint }}`,
				Active:   `> {{ if .Missing }}{{ "!" | red }} {{ else if .Current }}{{ "●" | cyan }} {{ else }}  {{ end }}{{ .Name | cyan }}  {{ .Branch | faint }}{{ if .SwitchBranch }}  {{ "→" | faint }}  {{ .SwitchBranch | cyan }}{{ end }}`,
				Inactive: `  {{ if .Missing }}{{ "!" | red }} {{ else if .Current }}{{ "●" | cyan }} {{ else }}  {{ end }}{{ .Name }}  {{ .Branch | faint }}{{ if .SwitchBranch }}  {{ "→" | faint }}  {{ .SwitchBranch }}{{ end }}`,
				Selected: `{{ "✓" | green }} {{ .Name | cyan }}`,
				Help:     `{{ "↑↓ navigate  ↵ select" | faint }}`,
			}

			tty, ttyErr := openTTY()
			if ttyErr == nil {
				defer tty.Close()
			}

			prompt := promptui.Select{
				Label:     fmt.Sprintf("Issue %s", ctx.ID),
				Items:     items,
				Templates: templates,
				Size:      15,
			}
			if ttyErr == nil {
				prompt.Stdin = tty
				prompt.Stdout = tty
			}

			idx, _, err := prompt.Run()
			if err != nil {
				return nil // cancelled
			}

			selected := items[idx]

			if selected.Missing {
				return fmt.Errorf("repo %q path %q no longer exists — was it moved or deleted?", strings.TrimSpace(selected.Name), selected.path)
			}

			if selected.SwitchBranch != "" {
				switchResults := f.GitRunner.Run([]string{selected.path}, "checkout", selected.SwitchBranch)
				if len(switchResults) > 0 && switchResults[0].Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s could not switch branch %q: %v\n", iostreams.Dim("!"), selected.SwitchBranch, switchResults[0].Err)
				}
			}

			rel, relErr := filepath.Rel(cwd, selected.path)
			if relErr != nil {
				rel = selected.path
			}

			if outFile != "" {
				return os.WriteFile(outFile, []byte(rel), 0o600)
			}
			fmt.Fprintln(f.IO.Out, rel)
			return nil
		},
	}

	cmd.Flags().StringVar(&outFile, "out", "", "write selected path to this file (used by shell wrapper)")
	return cmd
}
