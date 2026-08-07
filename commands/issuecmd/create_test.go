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

func TestCreateCmd_BranchFlagReusesExistingLocalBranch(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	runGit("branch", "existing-branch")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"111", "--branch", "existing-branch"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got := currentBranch(t, dir); got != "existing-branch" {
		t.Errorf("branch = %q, want %q", got, "existing-branch")
	}
}

func TestCreateCmd_BranchFlagTracksExistingRemoteBranch(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(d string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = d
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v (in %s) failed: %v\n%s", args, d, err, out)
		}
	}

	originDir := t.TempDir()
	runGit(originDir, "init", "-q", "--bare")
	runGit(dir, "remote", "add", "origin", originDir)

	defaultBranch := currentBranch(t, dir)
	runGit(dir, "checkout", "-b", "remote-branch")
	runGit(dir, "push", "-q", "origin", "remote-branch")
	runGit(dir, "checkout", defaultBranch)
	runGit(dir, "branch", "-D", "remote-branch")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"222", "--branch", "remote-branch"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got := currentBranch(t, dir); got != "remote-branch" {
		t.Errorf("branch = %q, want %q", got, "remote-branch")
	}

	trackedRemote := exec.Command("git", "config", "branch.remote-branch.remote")
	trackedRemote.Dir = dir
	out, err := trackedRemote.Output()
	if err != nil {
		t.Fatalf("git config branch.remote-branch.remote: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "origin" {
		t.Errorf("branch.remote-branch.remote = %q, want %q", got, "origin")
	}
}

func TestCreateCmd_BranchCheckoutFailureDoesNotCreateIssue(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	// a leaf branch named "conflict" blocks creating "conflict/sub" (git ref namespace collision)
	runGit("branch", "conflict")

	cmd := newCreateCmd(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"999", "--branch", "conflict/sub"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when branch checkout fails")
	}

	if hash := store.CurrentIssueHash(f.Workspace.Name); hash != "" {
		t.Fatalf("expected no issue to be created, but current issue hash is %q", hash)
	}
	issues, err := store.LoadIssuesForWorkspace(f.Workspace.Name, store.IssueStatusActive)
	if err != nil {
		t.Fatalf("load issues: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues saved, got %+v", issues)
	}
}

func TestCreateCmd_BranchFlagTracksExistingBranchOnNonOriginRemote(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(d string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = d
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v (in %s) failed: %v\n%s", args, d, err, out)
		}
	}

	upstreamDir := t.TempDir()
	runGit(upstreamDir, "init", "-q", "--bare")
	runGit(dir, "remote", "add", "upstream", upstreamDir)

	defaultBranch := currentBranch(t, dir)
	runGit(dir, "checkout", "-b", "contributor-branch")
	runGit(dir, "push", "-q", "upstream", "contributor-branch")
	runGit(dir, "checkout", defaultBranch)
	runGit(dir, "branch", "-D", "contributor-branch")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"333", "--branch", "contributor-branch"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got := currentBranch(t, dir); got != "contributor-branch" {
		t.Errorf("branch = %q, want %q", got, "contributor-branch")
	}

	trackedRemote := exec.Command("git", "config", "branch.contributor-branch.remote")
	trackedRemote.Dir = dir
	out, err := trackedRemote.Output()
	if err != nil {
		t.Fatalf("git config branch.contributor-branch.remote: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "upstream" {
		t.Errorf("branch.contributor-branch.remote = %q, want %q", got, "upstream")
	}
}

func TestCreateCmd_RemoteFlagTracksSpecificRemote(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(d string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = d
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v (in %s) failed: %v\n%s", args, d, err, out)
		}
	}

	forkedRepoDir := t.TempDir()
	runGit(forkedRepoDir, "init", "-q", "--bare")
	runGit(dir, "remote", "add", "forked-repo", forkedRepoDir)

	defaultBranch := currentBranch(t, dir)
	runGit(dir, "checkout", "-b", "feature/issue-display-branch-pattern")
	runGit(dir, "push", "-q", "forked-repo", "feature/issue-display-branch-pattern")
	runGit(dir, "checkout", defaultBranch)
	runGit(dir, "branch", "-D", "feature/issue-display-branch-pattern")

	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"19", "--branch", "feature/issue-display-branch-pattern", "--remote", "forked-repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	wantBranch := "feature/issue-display-branch-pattern"
	if got := currentBranch(t, dir); got != wantBranch {
		t.Errorf("branch = %q, want %q", got, wantBranch)
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	issue, err := store.LoadIssueByHash(f.Workspace.Name, hash)
	if err != nil {
		t.Fatalf("load issue: %v", err)
	}
	if issue.BranchSlug != wantBranch {
		t.Errorf("issue.BranchSlug = %q, want %q", issue.BranchSlug, wantBranch)
	}
	if issue.BranchRemote != "forked-repo" {
		t.Errorf("issue.BranchRemote = %q, want %q", issue.BranchRemote, "forked-repo")
	}

	trackedRemote := exec.Command("git", "config", "branch."+wantBranch+".remote")
	trackedRemote.Dir = dir
	out, err := trackedRemote.Output()
	if err != nil {
		t.Fatalf("git config branch.%s.remote: %v", wantBranch, err)
	}
	if got := strings.TrimSpace(string(out)); got != "forked-repo" {
		t.Errorf("tracked remote = %q, want %q", got, "forked-repo")
	}
}

func TestCreateCmd_RemoteFlagErrorsOnUntrackedLocalConflict(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(d string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = d
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v (in %s) failed: %v\n%s", args, d, err, out)
		}
	}

	forkedRepoDir := t.TempDir()
	runGit(forkedRepoDir, "init", "-q", "--bare")
	runGit(dir, "remote", "add", "forked-repo", forkedRepoDir)

	defaultBranch := currentBranch(t, dir)
	runGit(dir, "checkout", "-b", "feature/x")
	runGit(dir, "push", "-q", "forked-repo", "feature/x")
	runGit(dir, "checkout", defaultBranch)

	cmd := newCreateCmd(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"20", "--branch", "feature/x", "--remote", "forked-repo"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error due to untracked local branch conflict")
	}

	if hash := store.CurrentIssueHash(f.Workspace.Name); hash != "" {
		t.Fatalf("expected no issue to be created, but current issue hash is %q", hash)
	}
}

func TestCreateCmd_RemoteFlagErrorsWhenNotOnRemote(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(d string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = d
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v (in %s) failed: %v\n%s", args, d, err, out)
		}
	}

	forkedRepoDir := t.TempDir()
	runGit(forkedRepoDir, "init", "-q", "--bare")
	runGit(dir, "remote", "add", "forked-repo", forkedRepoDir)

	cmd := newCreateCmd(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"21", "--branch", "does-not-exist", "--remote", "forked-repo"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error because branch does not exist on forked-repo remote")
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
	f.Workspace.BranchPattern = "{type}/{issue}-{description}"
	err := f.Workspace.Save()
	if err != nil {
		t.Fatal("Unexpected Error occurred saving the BranchPattern")
	}

	cmd := newCreateCmd(f)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"33"})

	err = cmd.Execute()
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
}

func TestCreateCmd_NewBranchIsCutFromMainOrMasterRegardlessOfCurrentBranch(t *testing.T) {
	f, paths := newTestFactory(t, "repo-a")
	dir := paths["repo-a"]

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	gitOutput := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v failed: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}

	mainBranch := currentBranch(t, dir)
	mainHead := gitOutput("rev-parse", "HEAD")

	runGit("checkout", "-b", "unrelated-work")
	runGit("commit", "--allow-empty", "-q", "-m", "diverging commit")

	out := f.IO.Out.(*bytes.Buffer)
	cmd := newCreateCmd(f)
	cmd.SetArgs([]string{"555"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	hash := store.CurrentIssueHash(f.Workspace.Name)
	issue, err := store.LoadIssueByHash(f.Workspace.Name, hash)
	if err != nil {
		t.Fatalf("load issue: %v", err)
	}

	if got := currentBranch(t, dir); got != issue.BranchSlug {
		t.Fatalf("branch = %q, want %q", got, issue.BranchSlug)
	}
	if got := gitOutput("rev-parse", "HEAD"); got != mainHead {
		t.Errorf("new branch HEAD = %q, want it to match %q (%s), not the unrelated-work branch", got, mainHead, mainBranch)
	}

	if !strings.Contains(out.String(), "created new local branch from "+mainBranch) {
		t.Errorf("expected success output to mention branching from %q, got: %s", mainBranch, out.String())
	}
}
