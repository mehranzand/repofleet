package issuecmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util"
	"github.com/spf13/cobra"
)

type issueItem struct {
	// display columns
	RawID  string
	Hash   string
	Label  string
	Branch string
	Status string

	// row state
	Current  bool
	Archived bool

	// detail panel
	CreatedAt    string
	Description  string
	Kind         string
	ChangeType   string
	BranchRemote string
	Repos        string
}

func newSwitchCmd(f *factory.Factory) *cobra.Command {
	var showArchived bool

	cmd := &cobra.Command{
		Use:   "switch [issue-id]",
		Short: "Switch to an issue, or pick one interactively",
		Long: `Switch all repos to an issue's branch.

Without an argument, shows an interactive list to select from (hash shown beside each #issue).
With an issue ID, name, or hash, switches directly. If the ID matches more than one issue
(two unrelated issues sharing an ID across different repos), retry with the hash shown in
the error, or run with no argument to choose interactively.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name

			var ctx *store.Issue
			if len(args) == 1 {
				id := args[0]
				resolved, err := store.FindIssueByKeyword(ws, id, "")
				if err != nil {
					var ambigErr *store.AmbiguousIssueError
					if errors.As(err, &ambigErr) {
						return err
					}
					return fmt.Errorf("issue %q not found — create it with: rf issue create %s", id, id)
				}
				ctx = resolved
			} else {
				selectedHash, err := promptIssue(ws, showArchived, store.CurrentIssueHash(ws))
				if err != nil {
					return err
				}
				if selectedHash == "" {
					return nil
				}
				resolved, err := store.LoadIssueByHash(ws, selectedHash)
				if err != nil {
					return err
				}
				ctx = resolved
			}

			if err := store.SetCurrentIssueHash(ws, ctx.Hash); err != nil {
				return err
			}

			paths := repoPaths(ctx.Repos)
			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("Switching %d repo(s) to branch %q...", len(paths), ctx.BranchSlug)))

			results := f.GitRunner.Run(paths, "checkout", ctx.BranchSlug)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s: %s\n", iostreams.Red("✗"), r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), r.RepoPath)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showArchived, "archived", "A", false, "include archived issues in the list")
	return cmd
}

func promptIssue(wsName string, showArchived bool, currentHash string) (string, error) {
	status := store.IssueStatusActive
	if showArchived {
		status = ""
	}
	issues, err := store.LoadIssuesForWorkspace(wsName, status)
	if err != nil {
		return "", err
	}

	type row struct {
		issue   *store.Issue
		current bool
	}

	var rows []row
	maxBranch := 0
	for _, issue := range issues {
		rows = append(rows, row{issue, issue.Hash == currentHash})
		if len(issue.BranchSlug) > maxBranch {
			maxBranch = len(issue.BranchSlug)
		}
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("no issues found — create one with: rf issue create <id>")
	}

	for i, r := range rows {
		if r.current {
			rows[0], rows[i] = rows[i], rows[0]
			break
		}
	}

	pad := func(s string, n int) string {
		return s + strings.Repeat(" ", n-len(s))
	}

	repoNames := func(repos []store.Repo) string {
		names := make([]string, len(repos))
		for j, r := range repos {
			names[j] = r.Name
		}
		return strings.Join(names, ", ")
	}

	items := make([]issueItem, len(rows))
	maxLabel := 0
	for _, r := range rows {
		if labelLen := len(iostreams.IssueLabel(r.issue.ID, r.issue.Name)); labelLen > maxLabel {
			maxLabel = labelLen
		}
	}
	for i, r := range rows {
		label := iostreams.IssueLabel(r.issue.ID, r.issue.Name)
		items[i] = issueItem{
			RawID:        r.issue.ID,
			Hash:         r.issue.Hash,
			Label:        pad(label, maxLabel),
			Branch:       pad(r.issue.BranchSlug, maxBranch),
			Status:       string(r.issue.Status),
			Current:      r.current,
			Archived:     r.issue.Status == store.IssueStatusArchived,
			CreatedAt:    iostreams.FormatTime(r.issue.CreatedAt),
			Description:  r.issue.ShortDescription,
			Kind:         string(r.issue.Kind),
			ChangeType:   string(r.issue.ChangeType),
			BranchRemote: r.issue.BranchRemote,
			Repos:        repoNames(r.issue.Repos),
		}
	}

	activeRow :=
		`> ` +
			`{{ if .Archived }}{{ .Label | faint }}{{ else }}{{ .Label | green }}{{ end }}` +
			`  {{ .Hash | faint }}` +
			`  {{ .Branch | faint }}` +
			`  {{ .Status | faint }}` +
			`{{ if .Current }}  {{ "*" | green }}{{ end }}`

	inactiveRow :=
		`  ` +
			`{{ if .Archived }}{{ .Label | faint }}{{ else }}{{ .Label }}{{ end }}` +
			`  {{ .Hash | faint }}` +
			`  {{ .Branch | faint }}` +
			`  {{ .Status | faint }}` +
			`{{ if .Current }}  {{ "*" | green }}{{ end }}`

	details :=
		`{{ if .CreatedAt }}` + "\n" + `  {{ "Created:    " | cyan }}{{ .CreatedAt }}{{ end }}` +
			`{{ if .Description }}` + "\n" + `  {{ "Description: " | cyan }}{{ .Description }}{{ end }}` +
			`{{ if .Kind }}` + "\n" + `  {{ "Kind:        " | cyan }}{{ .Kind }}{{ end }}` +
			`{{ if .ChangeType }}` + "\n" + `  {{ "Type:        " | cyan }}{{ .ChangeType }}{{ end }}` +
			`{{ if .BranchRemote }}` + "\n" + `  {{ "Branch remote: " | cyan }}{{ .BranchRemote }}{{ end }}` +
			`{{ if .Repos }}` + "\n" + `  {{ "Repos:       " | cyan }}{{ .Repos }}{{ end }}`

	templates := &promptui.SelectTemplates{
		Label:    `{{ "Select issue:" | faint }}`,
		Active:   activeRow,
		Inactive: inactiveRow,
		Selected: `{{ "✓" | green }} {{ .Label | cyan }}`,
		Details:  details,
		Help:     `{{ "↑↓ navigate  / search  ↵ select  q/ctrl+c quit" | faint }}`,
	}

	prompt := promptui.Select{
		Label:     "Select",
		Items:     items,
		Templates: templates,
		Size:      10,
		Stdin:     util.QuitOnQ(os.Stdin),
		Stdout:    util.SilenceBell(os.Stdout),
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(items[index].RawID), strings.ToLower(input))
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", nil
	}
	return items[i].Hash, nil
}
