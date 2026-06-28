package issuecmd

import "github.com/mehranzand/repofleet/internal/config"

func repoPaths(repos []config.Repo) []string {
	paths := make([]string, len(repos))
	for i, r := range repos {
		paths[i] = r.Path
	}
	return paths
}

func filterRepos(repos []config.Repo, names []string) []config.Repo {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	var out []config.Repo
	for _, r := range repos {
		if set[r.Name] {
			out = append(out, r)
		}
	}
	return out
}
