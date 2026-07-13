package issuecmd

import (
	"fmt"
	"os"
	"strings"

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
