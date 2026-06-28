package issuecmd

import "github.com/mehranzand/repofleet/internal/store"

func repoPaths(repos []store.Repo) []string {
	paths := make([]string, len(repos))
	for i, r := range repos {
		paths[i] = r.Path
	}
	return paths
}

func filterRepos(repos []store.Repo, names []string) []store.Repo {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	var out []store.Repo
	for _, r := range repos {
		if set[r.Name] {
			out = append(out, r)
		}
	}
	return out
}
