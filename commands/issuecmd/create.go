package issuecmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

var (
	slugRe      = regexp.MustCompile(`[^a-z0-9/]+`)
	multiDashRe = regexp.MustCompile(`-{2,}`)
	nameRe      = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func toBranchSlug(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = multiDashRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func resolveBranchSlug(overrideBranch string, ws *store.Workspace, issue *store.Issue) (string, error) {
	if overrideBranch != "" {
		return overrideBranch, nil
	}
	if ws.BranchPattern != "" {
		slug, err := applyBranchPattern(ws.BranchPattern, ws, issue)
		if err != nil {
			return "", err
		}
		return slug, nil
	}
	return toBranchSlug(issue.ID), nil
}

func applyBranchPattern(pattern string, ws *store.Workspace, issue *store.Issue) (string, error) {
	tokenValues := map[string]string{
		"{workspace}":   toBranchSlug(ws.Name),
		"{issue}":       toBranchSlug(issue.ID),
		"{name}":        toBranchSlug(issue.Name),
		"{description}": toBranchSlug(issue.ShortDescription),
		"{kind}":        string(issue.Kind),
		"{type}":        string(issue.ChangeType),
	}

	flagHints := map[string]string{
		"{name}":        "--name",
		"{description}": "--description",
		"{kind}":        "--kind",
		"{type}":        "--type",
	}

	var missing []string
	for _, token := range regexp.MustCompile(`\{[^}]+\}`).FindAllString(pattern, -1) {
		if val, ok := tokenValues[token]; ok && val == "" {
			missing = append(missing, fmt.Sprintf("%s (provide %s)", token, flagHints[token]))
		}
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("branch pattern requires missing values: %s", strings.Join(missing, ", "))
	}

	pairs := make([]string, 0, len(tokenValues)*2)
	for token, val := range tokenValues {
		pairs = append(pairs, token, val)
	}
	slug := strings.NewReplacer(pairs...).Replace(pattern)
	slug = multiDashRe.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-/"), nil
}

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var branch string
	var name string
	var description string
	var kind string
	var changeType string
	var repoNames []string
	var skipBranch bool

	cmd := &cobra.Command{
		Use:   "create <issue-id>",
		Short: "Create an issue and branch across repos",
		Long: "Create an issue context and optionally create its branch across selected repos.\n\n" +
			"Branch name resolution order:\n" +
			"  1. --branch flag (overrides everything)\n" +
			"  2. Workspace branch pattern (rf workspace config --branch-pattern)\n" +
			"  3. Slugified issue ID (fallback)\n\n" +
			"Errors out before creating any branch if a selected repo's path no longer exists on disk.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace

			issueID := args[0]
			if _, err := strconv.Atoi(issueID); err != nil {
				return fmt.Errorf("issue ID must be an integer (e.g. 123), got %q", issueID)
			}
			if name != "" {
				if len(name) > 8 {
					return fmt.Errorf("issue name must be 8 characters or fewer, got %d", len(name))
				}
				if !nameRe.MatchString(name) {
					return fmt.Errorf("issue name must contain only letters, digits, hyphens, or underscores — no spaces or special characters")
				}
			}

			if kind != "" {
				validKinds := map[string]bool{"bug": true, "feature": true, "task": true, "story": true}
				if !validKinds[kind] {
					return fmt.Errorf("invalid --kind %q: must be bug, feature, task, or story", kind)
				}
			}
			if changeType != "" {
				validTypes := map[string]bool{"feat": true, "fix": true, "chore": true, "docs": true, "refactor": true, "test": true}
				if !validTypes[changeType] {
					return fmt.Errorf("invalid --type %q: must be feat, fix, chore, docs, refactor, or test", changeType)
				}
			}

			issue := &store.Issue{
				ID:               issueID,
				Name:             name,
				ShortDescription: description,
				Kind:             store.IssueKind(kind),
				ChangeType:       store.IssueChangeType(changeType),
				Workspace:        f.Workspace.Name,
				Status:           store.IssueStatusActive,
			}

			existing, err := store.LoadIssuesForWorkspace(f.Workspace.Name)
			if err != nil {
				return err
			}
			for _, ex := range existing {
				if strings.EqualFold(ex.ID, issue.ID) {
					return fmt.Errorf("issue ID %q already exists in this workspace", issue.ID)
				}
				if issue.Name != "" && strings.EqualFold(ex.Name, issue.Name) {
					return fmt.Errorf("issue name %q already exists in this workspace (ID: %s)", issue.Name, ex.ID)
				}
			}

			issue.Repos = ws.Repos
			if len(repoNames) > 0 {
				filtered, err := filterRepos(ws.Repos, repoNames)
				if err != nil {
					return err
				}
				issue.Repos = filtered
			}
			if len(issue.Repos) == 0 {
				return fmt.Errorf("no repos in current workspace — add one with: rf repo add <path>")
			}

			for _, r := range issue.Repos {
				if err := checkRepoPath(r); err != nil {
					return err
				}
			}

			if skipBranch {
				results := f.GitRunner.Run([]string{issue.Repos[0].Path}, "rev-parse", "--abbrev-ref", "HEAD")
				if len(results) > 0 && results[0].Err == nil {
					issue.BranchSlug = strings.TrimSpace(results[0].Stdout)
				}
			} else {
				slug, err := resolveBranchSlug(branch, ws, issue)
				if err != nil {
					return err
				}
				issue.BranchSlug = slug
			}

			if err := issue.Save(); err != nil {
				return err
			}
			if err := store.SetCurrentIssue(f.Workspace.Name, issue.ID); err != nil {
				return err
			}

			if skipBranch {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf("Issue %q created, tracking existing branches", issue.ID)))
				return nil
			}

			fmt.Fprintf(f.IO.Out, "%s\n\n", iostreams.Dim(fmt.Sprintf("Creating branch %q in %d repo(s)...", issue.BranchSlug, len(issue.Repos))))

			results := f.GitRunner.Run(repoPaths(issue.Repos), "checkout", "-b", issue.BranchSlug)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s: %s\n", iostreams.Red("✗"), r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"), r.RepoPath)
				}
			}

			fmt.Fprintf(f.IO.Out, "\n%s\n", iostreams.Success(fmt.Sprintf("Issue %q is now active on branch %q", issue.ID, issue.BranchSlug)))
			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "override branch name (ignores all naming rules)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "issue name (used in {name} token)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "short description (used in {description} token)")
	cmd.Flags().StringVar(&kind, "kind", "", "issue kind: bug, feature, task, story (used in {kind} token)")
	cmd.Flags().StringVar(&changeType, "type", "", "change type: feat, fix, chore, docs, refactor, test (used in {type} token)")
	cmd.Flags().StringSliceVarP(&repoNames, "repo", "r", nil, "repos to include (default: all in workspace)")
	cmd.Flags().BoolVar(&skipBranch, "skip-branch", false, "save issue context without creating a git branch")
	return cmd
}
