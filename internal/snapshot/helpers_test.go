package snapshot

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util/git"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func runGitAllowFail(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func gitStatus(t *testing.T, dir string) string {
	t.Helper()
	return runGit(t, dir, "status", "--porcelain")
}

// initRepo creates a git repo with committed base content and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "t@t.com")
	runGit(t, dir, "config", "user.name", "t")
	return dir
}

func newTestIssue(t *testing.T, id, dir string) *store.Issue {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return &store.Issue{
		ID:        id,
		Hash:      "hash-" + id,
		Workspace: "default",
		Repos:     []store.Repo{{Name: "repo", Path: dir}},
	}
}

func newTestRunner() *git.Runner { return git.NewRunner() }

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
