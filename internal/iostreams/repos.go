package iostreams

import (
	"io"
	"os"

	"github.com/mehranzand/repofleet/internal/store"
)

func PrintRepos(w io.Writer, repos []store.Repo) {
	t := NewTable()
	t.AddField("", Dim)
	t.AddField("Name", Dim)
	t.AddField("Forge", Dim)
	t.AddField("Remote", Dim)
	t.AddField("Path", Dim)
	t.EndRow()
	for _, r := range repos {
		marker, markerColor := "", Dim
		if _, err := os.Stat(r.Path); err != nil {
			marker, markerColor = "!", Red
		}
		t.AddField(marker, markerColor)
		t.AddField(r.Name, Cyan)
		t.AddField(string(r.Forge), Green)
		t.AddField(r.Remote, Dim)
		t.AddField(r.Path, Dim)
		t.EndRow()
	}
	t.Render(w)
}
