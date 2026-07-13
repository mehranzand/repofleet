package snapshot

import (
	"os"
	"testing"

	"github.com/mehranzand/repofleet/internal/store"
)

func TestCreateAndRestore_RoundTrip(t *testing.T) {
	dir := initRepo(t)

	writeFile(t, dir, "tracked.txt", "base\n")
	writeFile(t, dir, "deleted.txt", "to-delete\n")
	writeFile(t, dir, "deleted_staged.txt", "to-delete-staged\n")
	writeFile(t, dir, "oldname.txt", "old\n")
	writeFile(t, dir, "mm.txt", "mm-base\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "base")

	writeFile(t, dir, "tracked.txt", "base\nmodified\n")
	writeFile(t, dir, "staged_new.txt", "new\n")
	runGit(t, dir, "add", "staged_new.txt")
	writeFile(t, dir, "mm.txt", "mm-base\nstaged-part\n")
	runGit(t, dir, "add", "mm.txt")
	writeFile(t, dir, "mm.txt", "mm-base\nstaged-part\nunstaged-part\n")
	if err := os.Remove(dir + "/deleted.txt"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(dir + "/deleted_staged.txt"); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "deleted_staged.txt")
	runGit(t, dir, "mv", "oldname.txt", "newname.txt")
	writeFile(t, dir, "untracked.txt", "brand new\n")

	want := gitStatus(t, dir)

	issue := newTestIssue(t, "1", dir)
	runner := newTestRunner()

	snap, _, err := Create(runner, issue, false, false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := gitStatus(t, dir); got != want {
		t.Fatalf("Create without --clean mutated the working tree.\nbefore:\n%s\nafter:\n%s", want, got)
	}

	runGit(t, dir, "reset", "--hard", "-q", "HEAD")
	runGit(t, dir, "clean", "-fdq")
	if got := gitStatus(t, dir); got != "" {
		t.Fatalf("expected clean working tree after wipe, got:\n%s", got)
	}

	if _, err := Restore(runner, snap, false); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if got := gitStatus(t, dir); got != want {
		t.Errorf("status mismatch after restore.\nwant:\n%s\ngot:\n%s", want, got)
	}

	mm, err := os.ReadFile(dir + "/mm.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(mm) != "mm-base\nstaged-part\nunstaged-part\n" {
		t.Errorf("mm.txt content not restored correctly, got: %q", mm)
	}
}

func TestCreateAndRestore_Conflict(t *testing.T) {
	dir := initRepo(t)

	writeFile(t, dir, "f.txt", "line1\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "base")
	runGit(t, dir, "checkout", "-qb", "branchA")
	writeFile(t, dir, "f.txt", "line1\nlineA\n")
	runGit(t, dir, "commit", "-qam", "a")
	runGit(t, dir, "checkout", "-qb", "branchB", "master")
	writeFile(t, dir, "f.txt", "line1\nlineB\n")
	runGit(t, dir, "commit", "-qam", "b")

	// merge is expected to fail with a conflict — that's the point of this test
	runGitAllowFail(t, dir, "merge", "-q", "branchA")

	writeFile(t, dir, "extra.txt", "extra\n")
	runGit(t, dir, "add", "extra.txt")

	status := gitStatus(t, dir)
	if status == "" {
		t.Fatal("expected a conflicted status, got clean tree — merge may not have conflicted as expected")
	}

	issue := newTestIssue(t, "2", dir)
	runner := newTestRunner()

	snap, _, err := Create(runner, issue, true, false) // --clean: should merge --abort
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got := gitStatus(t, dir); got != "" {
		t.Fatalf("expected --clean to leave a clean tree (merge aborted), got:\n%s", got)
	}

	if _, err := Restore(runner, snap, false); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	content, err := os.ReadFile(dir + "/f.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := "line1\n<<<<<<< HEAD\nlineB\n=======\nlineA\n>>>>>>> branchA\n"
	if string(content) != want {
		t.Errorf("conflicted file content not restored verbatim.\nwant: %q\ngot:  %q", want, content)
	}

	got := gitStatus(t, dir)
	if got == "" {
		t.Error("expected extra.txt to still show as staged after restore")
	}
}

func TestCreate_DryRunWritesNothing(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "f.txt", "base\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "base")
	writeFile(t, dir, "f.txt", "base\nchanged\n")

	issue := newTestIssue(t, "3", dir)
	runner := newTestRunner()

	before := gitStatus(t, dir)
	snap, plan, err := Create(runner, issue, false, true)
	if err != nil {
		t.Fatalf("Create (dry-run): %v", err)
	}
	if len(plan) == 0 {
		t.Error("expected a non-empty plan")
	}
	if snap == nil {
		t.Error("expected snapshot metadata to be returned even in dry-run")
	}
	if got := gitStatus(t, dir); got != before {
		t.Errorf("dry-run mutated the working tree.\nbefore:\n%s\nafter:\n%s", before, got)
	}

	if _, err := os.Stat(store.SnapshotMetaPath(issue.Workspace, issue.Hash, snap.Hash)); err == nil {
		t.Error("dry-run should not have written snapshot.yaml to disk")
	}
}
