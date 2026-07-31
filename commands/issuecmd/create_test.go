package issuecmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util/git"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	run("commit", "--allow-empty", "-q", "-m", "initial commit")
}

func newTestFactory(t *testing.T, repoNames ...string) (*factory.Factory, map[string]string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := store.Initialize(); err != nil {
		t.Fatalf("store.Initialize: %v", err)
	}

	paths := make(map[string]string, len(repoNames))
	ws := &store.Workspace{Name: store.CurrentWorkspaceName()}
	for _, name := range repoNames {
		dir := filepath.Join(t.TempDir(), name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		initGitRepo(t, dir)
		paths[name] = dir
		ws.AddRepo(store.Repo{Name: name, Path: dir})
	}
	if err := ws.Save(); err != nil {
		t.Fatalf("save workspace: %v", err)
	}

	loaded, err := store.LoadWorkspace(ws.Name)
	if err != nil {
		t.Fatalf("load workspace: %v", err)
	}

	return &factory.Factory{
		Workspace: loaded,
		GitRunner: git.NewRunner(),
		IO: &iostreams.IOStreams{
			In:  bytes.NewReader(nil),
			Out: &bytes.Buffer{},
			Err: &bytes.Buffer{},
		},
	}, paths
}

func currentBranch(t *testing.T, repoDir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rev-parse in %s failed: %v\n%s", repoDir, err, out)
	}
	return string(bytes.TrimSpace(out))
}

func TestCreateCmd_RepoFlagFiltersToSelectedRepos(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a", "repo-b", "repo-c")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"123", "--repo", "repo-a", "--repo", "repo-c"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	if hash == "" {
		t.Fatal("expected current issue hash to be set")
	}
	issue, err := store.LoadIssueByHash(f.Workspace.Name, hash)
	if err != nil {
		t.Fatalf("load issue: %v", err)
	}

	if len(issue.Repos) != 2 {
		t.Fatalf("expected 2 repos on issue, got %d: %+v", len(issue.Repos), issue.Repos)
	}
	names := map[string]bool{}
	for _, r := range issue.Repos {
		names[r.Name] = true
	}
	if !names["repo-a"] || !names["repo-c"] {
		t.Fatalf("expected repo-a and repo-c, got %+v", issue.Repos)
	}
	if names["repo-b"] {
		t.Fatalf("repo-b should have been excluded, got %+v", issue.Repos)
	}

	if got := currentBranch(t, paths["repo-a"]); got != issue.BranchSlug {
		t.Errorf("repo-a branch = %q, want %q", got, issue.BranchSlug)
	}
	if got := currentBranch(t, paths["repo-c"]); got != issue.BranchSlug {
		t.Errorf("repo-c branch = %q, want %q", got, issue.BranchSlug)
	}
	if got := currentBranch(t, paths["repo-b"]); got == issue.BranchSlug {
		t.Errorf("repo-b should not have had the branch created, got %q", got)
	}
}

func TestCreateCmd_RepoFlagCommaSeparatedList(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a", "repo-b", "repo-c")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"456", "--repo", "repo-a,repo-b"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	issue, err := store.LoadIssueByHash(f.Workspace.Name, hash)
	if err != nil {
		t.Fatalf("load issue: %v", err)
	}
	if len(issue.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %+v", len(issue.Repos), issue.Repos)
	}

	if got := currentBranch(t, paths["repo-a"]); got != issue.BranchSlug {
		t.Errorf("repo-a branch = %q, want %q", got, issue.BranchSlug)
	}
	if got := currentBranch(t, paths["repo-b"]); got != issue.BranchSlug {
		t.Errorf("repo-b branch = %q, want %q", got, issue.BranchSlug)
	}
}

func TestCreateCmd_RepoFlagUnknownRepoErrors(t *testing.T) {
	f, _ := newTestFactory(t, "repo-a", "repo-b")

	cmd := newCreateCmd(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"789", "--repo", "repo-a", "--repo", "does-not-exist"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown repo")
	}

	if hash := store.CurrentIssueHash(f.Workspace.Name); hash != "" {
		t.Fatalf("expected no issue to be created, but current issue hash is %q", hash)
	}
}

func TestCreateCmd_NoRepoFlagIncludesAllRepos(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a", "repo-b")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"321"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	issue, err := store.LoadIssueByHash(f.Workspace.Name, hash)
	if err != nil {
		t.Fatalf("load issue: %v", err)
	}
	if len(issue.Repos) != 2 {
		t.Fatalf("expected all 2 repos included by default, got %d", len(issue.Repos))
	}
	for _, name := range []string{"repo-a", "repo-b"} {
		if got := currentBranch(t, paths[name]); got != issue.BranchSlug {
			t.Errorf("%s branch = %q, want %q", name, got, issue.BranchSlug)
		}
	}
}

func TestCreateCmd_MissingBranchValuesShowPattern(t *testing.T) {
	f, _ := newTestFactory(t, "repo-a")
	cmd := newCreateCmd(f)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"33"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing branch pattern values")
	}
	m := "branch pattern requires missing values"

	if !strings.Contains(err.Error(), m) {
		t.Errorf("actual: %s, it must contain: %s", err.Error(), m)
	}

	m = "Branch pattern: {type}/{issue}-{description}"
	if !strings.Contains(err.Error(), m) {
		t.Errorf("actual: %s, it must contain: %s", err.Error(), m)
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	if hash != "" {
		t.Fatalf("expected no issue to be created, but current issue hash is %q", hash)
	}
}
