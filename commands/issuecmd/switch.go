package issuecmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

type issueItem struct {
	// display columns
	RawID  string
	ID     string
	Branch string
	Status string

	// row state
	Current  bool
	Archived bool

	// detail panel
	Name        string
	Description string
	Kind        string
	ChangeType  string
	Repos       string
}

func newSwitchCmd(f *factory.Factory) *cobra.Command {
	var showArchived bool

	cmd := &cobra.Command{
		Use:   "switch [issue-id]",
		Short: "Switch to an issue, or pick one interactively",
		Long: `Switch all repos to an issue's branch.

Without an argument, shows an interactive list to select from.
With an issue ID or name, switches directly.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Settings.CurrentWorkspace

			var id string
			if len(args) == 1 {
				id = args[0]
			} else {
				selected, err := promptIssue(ws, showArchived, store.CurrentIssueID(ws))
				if err != nil {
					return err
				}
				if selected == "" {
					return nil
				}
				id = selected
			}

			ctx, err := store.LoadIssueByIDOrName(ws, id)
			if err != nil {
				return fmt.Errorf("issue %q not found — create it with: rf issue create %s", id, id)
			}
			if ctx.Workspace != ws {
				return fmt.Errorf("issue %q belongs to workspace %q, not %q", ctx.ID, ctx.Workspace, ws)
			}

			if err := store.SetCurrentIssue(ws, ctx.ID); err != nil {
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

func promptIssue(wsName string, showArchived bool, currentID string) (string, error) {
	issues, err := store.LoadIssuesForWorkspace(wsName)
	if err != nil {
		return "", err
	}

	type row struct {
		issue   *store.Issue
		current bool
	}

	var rows []row
	maxID, maxBranch := 0, 0
	for _, issue := range issues {
		if !showArchived && issue.Status == store.IssueStatusArchived {
			continue
		}
		rows = append(rows, row{issue, issue.ID == currentID})
		if len(issue.ID) > maxID {
			maxID = len(issue.ID)
		}
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
	for i, r := range rows {
		items[i] = issueItem{
			RawID:       r.issue.ID,
			ID:          "#" + pad(r.issue.ID, maxID),
			Branch:      pad(r.issue.BranchSlug, maxBranch),
			Status:      string(r.issue.Status),
			Current:     r.current,
			Archived:    r.issue.Status == store.IssueStatusArchived,
			Name:        r.issue.Name,
			Description: r.issue.ShortDescription,
			Kind:        string(r.issue.Kind),
			ChangeType:  string(r.issue.ChangeType),
			Repos:       repoNames(r.issue.Repos),
		}
	}

	activeRow :=
		`> ` +
		`{{ if .Archived }}{{ .ID | faint }}{{ else }}{{ .ID | green }}{{ end }}` +
		`  {{ .Branch | faint }}` +
		`  {{ .Status | faint }}` +
		`{{ if .Current }}  {{ "*" | green }}{{ end }}`

	inactiveRow :=
		`  ` +
		`{{ if .Archived }}{{ .ID | faint }}{{ else }}{{ .ID }}{{ end }}` +
		`  {{ .Branch | faint }}` +
		`  {{ .Status | faint }}` +
		`{{ if .Current }}  {{ "*" | green }}{{ end }}`

	details :=
		`{{ if .Name }}`        + "\n" + `  {{ "Name:        " | cyan }}{{ .Name }}{{ end }}`        +
		`{{ if .Description }}` + "\n" + `  {{ "Description: " | cyan }}{{ .Description }}{{ end }}` +
		`{{ if .Kind }}`        + "\n" + `  {{ "Kind:        " | cyan }}{{ .Kind }}{{ end }}`        +
		`{{ if .ChangeType }}`  + "\n" + `  {{ "Type:        " | cyan }}{{ .ChangeType }}{{ end }}`  +
		`{{ if .Repos }}`       + "\n" + `  {{ "Repos:       " | cyan }}{{ .Repos }}{{ end }}`

	templates := &promptui.SelectTemplates{
		Label:    `{{ "Select issue:" | faint }}`,
		Active:   activeRow,
		Inactive: inactiveRow,
		Selected: `{{ "✓" | green }} {{ .RawID | cyan }}`,
		Details:  details,
		Help:     `{{ "↑↓ navigate  / search  ↵ select" | faint }}`,
	}

	prompt := promptui.Select{
		Label:     "Select",
		Items:     items,
		Templates: templates,
		Searcher:  func(input string, index int) bool {
			return strings.Contains(strings.ToLower(items[index].RawID), strings.ToLower(input))
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", nil
	}
	return items[i].RawID, nil
}
