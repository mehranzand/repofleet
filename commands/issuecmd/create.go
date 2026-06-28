package issuecmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func toBranchSlug(id string) string {
	s := strings.ToLower(id)
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var branch string
	var repoNames []string

	cmd := &cobra.Command{
		Use:   "create <issue-id>",
		Short: "Create an issue context and branch across selected repos",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			slug := branch
			if slug == "" {
				slug = toBranchSlug(id)
			}

			ws := f.Config.CurrentWS()
			repos := ws.Repos
			if len(repoNames) > 0 {
				repos = filterRepos(ws.Repos, repoNames)
			}
			if len(repos) == 0 {
				return fmt.Errorf("no repos selected — add repos with: repofleet repo add <path>")
			}

			ctx := &store.Issue{
				ID:         id,
				BranchSlug: slug,
				Repos:      repos,
				Status:     store.IssueStatusActive,
			}
			if err := ctx.Save(); err != nil {
				return err
			}
			if err := store.SetCurrentIssue(id); err != nil {
				return err
			}

			paths := repoPaths(repos)
			fmt.Fprintf(f.IO.Out, "Creating branch %q in %d repo(s)...\n\n", slug, len(paths))
			results := f.GitRunner.Run(paths, "checkout", "-b", slug)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Err, "  x %s: %s\n", r.RepoPath, r.Err)
				} else {
					fmt.Fprintf(f.IO.Out, "  ok %s\n", r.RepoPath)
				}
			}

			fmt.Fprintf(f.IO.Out, "\nIssue %q is now active on branch %q\n", id, slug)
			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "branch name (default: slugified issue id)")
	cmd.Flags().StringArrayVarP(&repoNames, "repo", "r", nil, "repos to include (default: all in workspace)")
	return cmd
}
