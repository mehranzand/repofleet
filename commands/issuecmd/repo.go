package issuecmd

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func newRepoCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Add or remove repos from the current issue",
	}
	cmd.AddCommand(newRepoAddCmd(f))
	cmd.AddCommand(newRepoRemoveCmd(f))
	return cmd
}

func newRepoAddCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "add <repo-name>",
		Short: "Add a repo to the current issue and create its branch",
		Long:  "Add a repo to the current issue and create its branch. Errors if the repo's path no longer exists on disk.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name
			repoName := args[0]

			hash := store.CurrentIssueHash(ws)
			if hash == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch")
			}

			issue, err := store.LoadIssueByHash(ws, hash)
			if err != nil {
				return err
			}

			// verify repo exists in workspace
			var target *store.Repo
			for _, r := range f.Workspace.Repos {
				if r.Name == repoName {
					r := r
					target = &r
					break
				}
			}
			if target == nil {
				available := make([]string, len(f.Workspace.Repos))
				for i, r := range f.Workspace.Repos {
					available[i] = r.Name
				}
				return fmt.Errorf("repo %q not found in workspace — available: %s", repoName, strings.Join(available, ", "))
			}

			// check not already in issue
			for _, r := range issue.Repos {
				if r.Name == repoName {
					return fmt.Errorf("repo %q is already part of issue %q", repoName, issue.ID)
				}
			}

			if err := checkRepoPath(*target); err != nil {
				return err
			}

			// check if branch already exists; if so, switch otherwise create
			exists := f.GitRunner.Run([]string{target.Path}, "rev-parse", "--verify", "--quiet", "refs/heads/"+issue.BranchSlug)
			var checkoutArgs []string
			if len(exists) > 0 && exists[0].Err == nil {
				checkoutArgs = []string{"checkout", issue.BranchSlug}
			} else {
				checkoutArgs = []string{"checkout", "-b", issue.BranchSlug}
			}
			results := f.GitRunner.Run([]string{target.Path}, checkoutArgs...)
			for _, r := range results {
				if r.Err != nil {
					return fmt.Errorf("failed to switch to branch %q in %q: %s", issue.BranchSlug, repoName, r.Err)
				}
			}

			issue.Repos = append(issue.Repos, *target)
			if err := issue.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf(
				"Added %q to issue %q on branch %q", repoName, issue.ID, issue.BranchSlug,
			)))
			return nil
		},
	}
}

func newRepoRemoveCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <repo-name>",
		Short: "Remove a repo from the current issue",
		Long: "Remove a repo from the current issue. Deletes its branch if there are no uncommitted changes. If the branch is currently checked out, automatically switches to main or master first.\n\n" +
			"If the repo's path no longer exists on disk, it is still unlinked from the issue, but branch cleanup is skipped with a warning.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := f.Workspace.Name
			repoName := args[0]

			hash := store.CurrentIssueHash(ws)
			if hash == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch")
			}

			issue, err := store.LoadIssueByHash(ws, hash)
			if err != nil {
				return err
			}

			var target *store.Repo
			remaining := issue.Repos[:0]
			for _, r := range issue.Repos {
				if r.Name == repoName {
					r := r
					target = &r
				} else {
					remaining = append(remaining, r)
				}
			}
			if target == nil {
				return fmt.Errorf("repo %q is not part of issue %q", repoName, issue.ID)
			}

			if pathErr := checkRepoPath(*target); pathErr != nil {
				issue.Repos = remaining
				if err := issue.Save(); err != nil {
					return err
				}
				fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"),
					fmt.Sprintf("Removed %q from issue %q", repoName, issue.ID))
				fmt.Fprintf(f.IO.Out, "  %s %s — skipped branch cleanup\n", iostreams.Dim("!"), pathErr)
				return nil
			}

			// check for uncommitted changes
			statusResults := f.GitRunner.Run([]string{target.Path}, "status", "--short")
			hasChanges := len(statusResults) > 0 && statusResults[0].Err == nil &&
				strings.TrimSpace(statusResults[0].Stdout) != ""

			issue.Repos = remaining
			if err := issue.Save(); err != nil {
				return err
			}

			protected := issue.BranchSlug == "main" || issue.BranchSlug == "master"

			if hasChanges || protected {
				fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"),
					fmt.Sprintf("Removed %q from issue %q", repoName, issue.ID))
				if hasChanges {
					fmt.Fprintf(f.IO.Out, "  %s branch %q in %q has uncommitted changes — branch not deleted\n",
						iostreams.Dim("!"), issue.BranchSlug, repoName)
				} else {
					fmt.Fprintf(f.IO.Out, "  %s branch %q is protected — not deleted\n",
						iostreams.Dim("!"), issue.BranchSlug)
				}
				return nil
			}

			currentResults := f.GitRunner.Run([]string{target.Path}, "rev-parse", "--abbrev-ref", "HEAD")
			if len(currentResults) > 0 && currentResults[0].Err == nil &&
				strings.TrimSpace(currentResults[0].Stdout) == issue.BranchSlug {
				switched := false
				for _, fallback := range []string{"main", "master"} {
					r := f.GitRunner.Run([]string{target.Path}, "checkout", fallback)
					if len(r) > 0 && r[0].Err == nil {
						switched = true
						break
					}
				}
				if !switched {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"),
						fmt.Sprintf("Removed %q from issue %q", repoName, issue.ID))
					fmt.Fprintf(f.IO.Out, "  %s branch %q is checked out — switch to main/master first to allow deletion\n",
						iostreams.Dim("!"), issue.BranchSlug)
					return nil
				}
			}

			deleteResults := f.GitRunner.Run([]string{target.Path}, "branch", "-d", issue.BranchSlug)
			for _, r := range deleteResults {
				if r.Err != nil {
					fmt.Fprintf(f.IO.Out, "  %s %s\n", iostreams.Green("✓"),
						fmt.Sprintf("Removed %q from issue %q", repoName, issue.ID))
					fmt.Fprintf(f.IO.Out, "  %s could not delete branch %q: %s\n",
						iostreams.Dim("!"), issue.BranchSlug, r.Err)
					return nil
				}
			}

			fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Success(fmt.Sprintf(
				"Removed %q from issue %q and deleted branch %q", repoName, issue.ID, issue.BranchSlug,
			)))
			return nil
		},
	}
}
