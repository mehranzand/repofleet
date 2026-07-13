package store

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mehranzand/repofleet/internal/util"
)

func snapshotsRootDir() string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "snapshots")
}

func snapshotWorkspaceDir(wsName string) string {
	return filepath.Join(snapshotsRootDir(), wsName)
}

func snapshotIssueDir(wsName, issueHash string) string {
	return filepath.Join(snapshotWorkspaceDir(wsName), issueHash)
}

func snapshotDir(wsName, issueHash, hash string) string {
	return filepath.Join(snapshotIssueDir(wsName, issueHash), hash)
}

func SnapshotDir(wsName, issueHash, hash string) string {
	return snapshotDir(wsName, issueHash, hash)
}

func NewSnapshotHash() (string, error) {
	return util.ShortHash()
}

func SnapshotMetaPath(wsName, issueHash, hash string) string {
	return filepath.Join(snapshotDir(wsName, issueHash, hash), "snapshot.yaml")
}

func SnapshotStagedPatchPath(wsName, issueHash, hash, repoName string) string {
	return filepath.Join(snapshotDir(wsName, issueHash, hash), repoName+".staged.patch")
}

func SnapshotUnstagedPatchPath(wsName, issueHash, hash, repoName string) string {
	return filepath.Join(snapshotDir(wsName, issueHash, hash), repoName+".unstaged.patch")
}

func SnapshotUntrackedDir(wsName, issueHash, hash, repoName string) string {
	return filepath.Join(snapshotDir(wsName, issueHash, hash), repoName+".untracked")
}

func SnapshotConflictedDir(wsName, issueHash, hash, repoName string) string {
	return filepath.Join(snapshotDir(wsName, issueHash, hash), repoName+".conflicted")
}

func LoadSnapshot(wsName, issueHash, hash string) (*Snapshot, error) {
	data, err := os.ReadFile(SnapshotMetaPath(wsName, issueHash, hash))
	if err != nil {
		return nil, err
	}
	var s Snapshot
	return &s, yaml.Unmarshal(data, &s)
}

func ListSnapshotsForIssue(wsName, issueHash string) ([]*Snapshot, error) {
	entries, err := os.ReadDir(snapshotIssueDir(wsName, issueHash))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var snaps []*Snapshot
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		snap, err := LoadSnapshot(wsName, issueHash, e.Name())
		if err != nil {
			continue
		}
		snaps = append(snaps, snap)
	}
	return snaps, nil
}

func LatestSnapshot(wsName, issueHash string) (*Snapshot, error) {
	snaps, err := ListSnapshotsForIssue(wsName, issueHash)
	if err != nil {
		return nil, err
	}
	var latest *Snapshot
	var latestTime time.Time
	for _, s := range snaps {
		t, err := time.Parse(time.RFC3339, s.CreatedAt)
		if err != nil {
			continue
		}
		if latest == nil || t.After(latestTime) {
			latest, latestTime = s, t
		}
	}
	return latest, nil
}


func DeleteSnapshot(wsName, issueHash, hash string) error {
	return os.RemoveAll(snapshotDir(wsName, issueHash, hash))
}

func DeleteSnapshotsForIssue(wsName, issueHash string) error {
	return os.RemoveAll(snapshotIssueDir(wsName, issueHash))
}

func DeleteSnapshotsForWorkspace(wsName string) error {
	return os.RemoveAll(snapshotWorkspaceDir(wsName))
}

func ListSnapshots(wsName string) ([]*Snapshot, error) {
	entries, err := os.ReadDir(snapshotWorkspaceDir(wsName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var snaps []*Snapshot
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		issueSnaps, err := ListSnapshotsForIssue(wsName, e.Name())
		if err != nil {
			continue
		}
		snaps = append(snaps, issueSnaps...)
	}

	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].CreatedAt > snaps[j].CreatedAt
	})
	return snaps, nil
}
