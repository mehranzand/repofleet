package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Result struct {
	RepoPath string
	Stdout   string
	Err      error
}

type indexedResult struct {
	idx int
	res Result
}

type Runner struct{}

func NewRunner() *Runner { return &Runner{} }


func (r *Runner) Run(repoPaths []string, args ...string) []Result {
	ch := make(chan indexedResult, len(repoPaths))

	for i, path := range repoPaths {
		go func(idx int, p string) {
			ch <- indexedResult{idx, run(p, args...)}
		}(i, path)
	}

	results := make([]Result, len(repoPaths))
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
