package iostreams

import (
	"fmt"
	"io"

	"github.com/mehranzand/repofleet/internal/store"
)

func PrintReposOrEmpty(w io.Writer, repos []store.Repo, emptyMsg string) {
	if len(repos) == 0 {
		fmt.Fprintf(w, "%s\n", Dim(emptyMsg))
		return
	}
	PrintRepos(w, repos)
}

func PrintRepos(w io.Writer, repos []store.Repo) {
	t := NewTable()
	t.AddField("Name", Dim)
	t.AddField("Forge", Dim)
	t.AddField("URL", Dim)
	t.AddField("Path", Dim)
	t.EndRow()
	for _, r := range repos {
		t.AddField(r.Name, Cyan)
		t.AddField(string(r.Forge), Green)
		t.AddField(r.URL, Dim)
		t.AddField(r.Path, Dim)
		t.EndRow()
	}
	t.Render(w)
}
