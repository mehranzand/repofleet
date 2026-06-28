package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Result holds the output of a git command run inside one repository.
type Result struct {
	RepoPath string
	Stdout   string
	Err      error
}

// Runner executes git commands across one or more repositories concurrently.
type Runner struct{}

func NewRunner() *Runner { return &Runner{} }

// Run executes the given git args in every repoPath concurrently and
// returns one Result per repository.
func (r *Runner) Run(repoPaths []string, args ...string) []Result {
	results := make([]Result, len(repoPaths))
	ch := make(chan struct{ idx int; res Result }, len(repoPaths))

	for i, path := range repoPaths {
		go func(idx int, p string) {
			ch <- struct {
				idx int
				res Result
			}{idx, run(p, args...)}
		}(i, path)
	}

	for range repoPaths {
		item := <-ch
		results[item.idx] = item.res
	}
	return results
}

func run(repoPath string, args ...string) Result {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%w: %s", err, stderr.String())
	}
	return Result{
		RepoPath: repoPath,
		Stdout:   stdout.String(),
		Err:      err,
	}
}
