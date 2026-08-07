package issuecmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/store"
)

func checkRepoPath(r store.Repo) error {
	info, err := os.Stat(r.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("repo %q path %q no longer exists — was it moved or deleted?", r.Name, r.Path)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("repo %q path %q is not a directory", r.Name, r.Path)
	}
	return nil
}

func repoPaths(repos []store.Repo) []string {
	paths := make([]string, len(repos))
	for i, r := range repos {
		paths[i] = r.Path
	}
	return paths
}

// branchAction describes which path checkoutOrCreateBranch took to reach the branch.
type branchAction string

const (
	branchCheckedOut branchAction = "checked out existing local branch"
)

func branchTrackedFrom(remote string) branchAction {
	return branchAction(fmt.Sprintf("fetched and tracked from %s", remote))
}

func branchCreatedFrom(base string) branchAction {
	return branchAction(fmt.Sprintf("created new local branch from %s", base))
}

func mainOrMasterBranch(f *factory.Factory, path string) (string, error) {
	for _, candidate := range []string{"main", "master"} {
		exists := f.GitRunner.Run([]string{path}, "rev-parse", "--verify", "--quiet", "refs/heads/"+candidate)
		if len(exists) > 0 && exists[0].Err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("neither %q nor %q branch found locally", "main", "master")
}

func repoRemotes(f *factory.Factory, path string) []string {
	results := f.GitRunner.Run([]string{path}, "remote")
	if len(results) == 0 || results[0].Err != nil {
		return nil
	}
	var remotes []string
	for _, line := range strings.Split(results[0].Stdout, "\n") {
		if name := strings.TrimSpace(line); name != "" {
			remotes = append(remotes, name)
		}
	}
	for i, r := range remotes {
		if r == "origin" {
			remotes[0], remotes[i] = remotes[i], remotes[0]
			break
		}
	}
	return remotes
}

func checkoutOrCreateBranch(f *factory.Factory, path, branch string) (branchAction, error) {
	localExists := f.GitRunner.Run([]string{path}, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if len(localExists) > 0 && localExists[0].Err == nil {
		results := f.GitRunner.Run([]string{path}, "checkout", branch)
		return branchCheckedOut, results[0].Err
	}

	for _, remote := range repoRemotes(f, path) {
		remoteExists := f.GitRunner.Run([]string{path}, "ls-remote", "--exit-code", "--heads", remote, branch)
		if len(remoteExists) == 0 || remoteExists[0].Err != nil || strings.TrimSpace(remoteExists[0].Stdout) == "" {
			continue
		}
		if fetchResults := f.GitRunner.Run([]string{path}, "fetch", remote, branch); fetchResults[0].Err != nil {
			return branchTrackedFrom(remote), fetchResults[0].Err
		}
		results := f.GitRunner.Run([]string{path}, "checkout", "-b", branch, "--track", remote+"/"+branch)
		return branchTrackedFrom(remote), results[0].Err
	}

	base, err := mainOrMasterBranch(f, path)
	if err != nil {
		return "", err
	}
	results := f.GitRunner.Run([]string{path}, "checkout", "-b", branch, base)
	return branchCreatedFrom(base), results[0].Err
}

func checkoutRemoteQualifiedBranch(f *factory.Factory, path, remote, branchName string) (branchAction, error) {
	localExists := f.GitRunner.Run([]string{path}, "rev-parse", "--verify", "--quiet", "refs/heads/"+branchName)
	if len(localExists) > 0 && localExists[0].Err == nil {
		upstream := f.GitRunner.Run([]string{path}, "rev-parse", "--abbrev-ref", branchName+"@{upstream}")
		if len(upstream) > 0 && upstream[0].Err == nil && strings.TrimSpace(upstream[0].Stdout) == remote+"/"+branchName {
			results := f.GitRunner.Run([]string{path}, "checkout", branchName)
			return branchCheckedOut, results[0].Err
		}
		return "", fmt.Errorf(
			"local branch %q already exists but does not track %s/%s — delete/rename it first, or use a different --branch name",
			branchName, remote, branchName,
		)
	}

	remoteExists := f.GitRunner.Run([]string{path}, "ls-remote", "--exit-code", "--heads", remote, branchName)
	if len(remoteExists) == 0 || remoteExists[0].Err != nil || strings.TrimSpace(remoteExists[0].Stdout) == "" {
		detail := ""
		if len(remoteExists) > 0 && remoteExists[0].Err != nil {
			detail = fmt.Sprintf(": %s", remoteExists[0].Err)
		}
		return "", fmt.Errorf("branch %q not found on remote %q%s", branchName, remote, detail)
	}

	if fetchResults := f.GitRunner.Run([]string{path}, "fetch", remote, branchName); fetchResults[0].Err != nil {
		return branchTrackedFrom(remote), fetchResults[0].Err
	}
	results := f.GitRunner.Run([]string{path}, "checkout", "-b", branchName, "--track", remote+"/"+branchName)
	return branchTrackedFrom(remote), results[0].Err
}

func filterRepos(repos []store.Repo, names []string) ([]store.Repo, error) {
	index := make(map[string]store.Repo, len(repos))
	for _, r := range repos {
		index[r.Name] = r
	}

	var out []store.Repo
	var unknown []string
	for _, n := range names {
		if r, ok := index[n]; ok {
			out = append(out, r)
		} else {
			unknown = append(unknown, n)
		}
	}

	if len(unknown) > 0 {
		available := make([]string, 0, len(repos))
		for _, r := range repos {
			available = append(available, r.Name)
		}
		return nil, fmt.Errorf(
			"repo %q not found in current workspace — available: %s",
			strings.Join(unknown, ", "),
			strings.Join(available, ", "),
		)
	}
	return out, nil
}
