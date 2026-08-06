package iostreams

import (
	"fmt"
	"io"
	"strings"

	"github.com/mehranzand/repofleet/internal/store"
)

func IssueLabel(id, name string) string {
	label := "#" + id
	if name != "" {
		label = fmt.Sprintf("%s (%s)", label, name)
	}
	return label
}

func PrintIssues(w io.Writer, issues []*store.Issue, currentHash string) {
	t := NewTable()
	t.AddField("", Dim)
	t.AddField("ID", Dim)
	t.AddField("Hash", Dim)
	t.AddField("Branch", Dim)
	t.AddField("Status", Dim)
	t.AddField("Repos", Dim)
	t.EndRow()
	for _, i := range issues {
		marker, markerColor := "", Dim
		if i.Hash == currentHash {
			marker, markerColor = "*", Green
		}
		names := make([]string, len(i.Repos))
		for j, r := range i.Repos {
			names[j] = r.Name
		}
		statusColor := Green
		if i.Status == store.IssueStatusArchived {
			statusColor = Dim
		}
		t.AddField(marker, markerColor)
		t.AddField(IssueLabel(i.ID, i.Name), Cyan)
		t.AddField(i.Hash, Dim)
		t.AddField(i.BranchSlug, Dim)
		t.AddField(string(i.Status), statusColor)
		t.AddField(strings.Join(names, ", "), Dim)
		t.EndRow()
	}
	t.Render(w)
}
