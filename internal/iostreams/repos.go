package iostreams

import (
	"io"

	"github.com/mehranzand/repofleet/internal/store"
)

func PrintRepos(w io.Writer, repos []store.Repo) {
	t := NewTable()
	t.AddField("Name", Dim)
	t.AddField("Forge", Dim)
	t.AddField("URL", Dim)
	t.AddField("Path", Dim)
	t.EndRow()
	for _, r := range repos {
		t.AddField(r.Name, Cyan)
		t.AddField(r.Forge, Green)
		t.AddField(r.URL, Dim)
		t.AddField(r.Path, Dim)
		t.EndRow()
	}
	t.Render(w)
}
