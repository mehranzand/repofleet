package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mehranzand/repofleet/internal/util"
)

type AmbiguousIssueError struct {
	Query   string
	Matches []*Issue
}

func issueDir(wsName string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "issues", wsName)
}

func issuePath(wsName, hash string) string {
	return filepath.Join(issueDir(wsName), hash+".yaml")
}

func NewIssueHash() (string, error) {
	return util.ShortHash()
}

func LoadIssueByHash(wsName, hash string) (*Issue, error) {
	data, err := os.ReadFile(issuePath(wsName, hash))
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := yaml.Unmarshal(data, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func LoadIssuesForWorkspace(wsName string, status IssueStatus) ([]*Issue, error) {
	dir := issueDir(wsName)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var issues []*Issue
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		issue, err := LoadIssueByHash(wsName, strings.TrimSuffix(e.Name(), ".yaml"))
		if err != nil {
			continue
		}
		if status != "" && issue.Status != status {
			continue
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

func (e *AmbiguousIssueError) Error() string {
	hashes := make([]string, len(e.Matches))
	for i, m := range e.Matches {
		hashes[i] = m.Hash
	}
	return fmt.Sprintf("issue %q is ambiguous — %d separate issues share this ID across different repos; retry with one of these hashes: %s (or run `rf issue switch` with no argument to choose interactively)", e.Query, len(e.Matches), strings.Join(hashes, ", "))
}

func FindIssueByKeyword(wsName, keyword string, status IssueStatus) (*Issue, error) {
	issues, err := LoadIssuesForWorkspace(wsName, status)
	if err != nil {
		return nil, err
	}

	var idMatches []*Issue
	for _, issue := range issues {
		if issue.ID == keyword {
			idMatches = append(idMatches, issue)
		}
	}
	if len(idMatches) == 1 {
		return idMatches[0], nil
	}

	if len(idMatches) == 0 {
		for _, issue := range issues {
			if strings.EqualFold(issue.Name, keyword) {
				return issue, nil
			}
		}
	}

	if issue, err := LoadIssueByHash(wsName, keyword); err == nil {
		return issue, nil
	}

	if len(idMatches) > 1 {
		return nil, &AmbiguousIssueError{Query: keyword, Matches: idMatches}
	}
	return nil, fmt.Errorf("issue %q not found by ID, name, or hash", keyword)
}

func (i *Issue) Save() error {
	path := issuePath(i.Workspace, i.Hash)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(i)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func DeleteIssue(wsName, hash string) error {
	if err := os.Remove(issuePath(wsName, hash)); err != nil {
		return err
	}
	if CurrentIssueHash(wsName) == hash {
		_ = SetCurrentIssueHash(wsName, "")
	}
	return nil
}

func DeleteIssuesForWorkspace(wsName string) error {
	return os.RemoveAll(issueDir(wsName))
}
