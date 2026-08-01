package issuecmd

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/mehranzand/repofleet/internal/store"
)

func TestRepoAddCmd_UsesIssueBranchRemote(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a", "repo-b")
	dirA, dirB := paths["repo-a"], paths["repo-b"]

	runGit := func(d string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = d
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v (in %s) failed: %v\n%s", args, d, err, out)
		}
	}

	forkedRepoDir := t.TempDir()
	runGit(forkedRepoDir, "init", "-q", "--bare")
	runGit(dirA, "remote", "add", "forked-repo", forkedRepoDir)

	defaultBranch := currentBranch(t, dirA)
	runGit(dirA, "checkout", "-b", "feature/x")
	runGit(dirA, "push", "-q", "forked-repo", "feature/x")
	runGit(dirA, "checkout", defaultBranch)
	runGit(dirA, "branch", "-D", "feature/x")

	createCmd := newCreateCmd(f)
	createCmd.SetArgs([]string{"42", "--repo", "repo-a", "--branch", "feature/x", "--remote", "forked-repo"})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("create: %v", err)
	}

	runGit(dirB, "branch", "feature/x")

	addCmd := newRepoAddCmd(f)
	addCmd.SetOut(&bytes.Buffer{})
	addCmd.SetErr(&bytes.Buffer{})
	addCmd.SetArgs([]string{"repo-b"})
	err := addCmd.Execute()
	if err == nil {
		t.Fatal("expected error: repo-b's untracked local feature/x conflicts with the issue's stored --remote forked-repo")
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	issue, loadErr := store.LoadIssueByHash(f.Workspace.Name, hash)
	if loadErr != nil {
		t.Fatalf("load issue: %v", loadErr)
	}
	for _, r := range issue.Repos {
		if r.Name == "repo-b" {
			t.Fatal("repo-b should not have been added to the issue after a failed checkout")
		}
	}
}
