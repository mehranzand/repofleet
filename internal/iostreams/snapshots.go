package iostreams

import (
	"fmt"
	"io"

	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util"
)

func PrintSnapshots(w io.Writer, snaps []*store.Snapshot) {
	var order []string
	groups := map[string][]*store.Snapshot{}
	for _, s := range snaps {
		if _, ok := groups[s.IssueHash]; !ok {
			order = append(order, s.IssueHash)
		}
		groups[s.IssueHash] = append(groups[s.IssueHash], s)
	}

	for _, issueHash := range order {
		group := groups[issueHash]
		first := group[0]
		fmt.Fprintf(w, "%s %s\n",
			Bold("Issue #"+first.IssueID),
			Dim(fmt.Sprintf("(hash %s, %d snapshot(s))", issueHash, len(group))),
		)

		t := NewTable()
		t.AddField("Hash", Dim)
		t.AddField("Created", Dim)
		t.AddField("Size", Dim)
		t.AddField("Repos", Dim)
		t.EndRow()
		for _, s := range group {
			changed := 0
			for _, rs := range s.Repos {
				if rs.StagedPatch != "" || rs.UnstagedPatch != "" || len(rs.UntrackedFiles) > 0 || len(rs.ConflictedFiles) > 0 {
					changed++
				}
			}
			size := "-"
			if n, err := util.DirSize(store.SnapshotDir(s.Workspace, s.IssueHash, s.Hash)); err == nil {
				size = util.FormatBytes(n)
			}
			t.AddField(s.Hash, Cyan)
			t.AddField(formatDatetime(s.CreatedAt), Dim)
			t.AddField(size, nil)
			t.AddField(fmt.Sprintf("%d/%d changed", changed, len(s.Repos)), Green)
			t.EndRow()
		}
		t.Render(w)
	}
}
