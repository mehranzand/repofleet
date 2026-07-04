package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func CurrentIssueID(_ string) string {
	s, err := LoadSettings()
	if err != nil {
		return ""
	}
	return s.CurrentIssue
}

func SetCurrentIssue(_ string, id string) error {
	s, err := LoadSettings()
	if err != nil {
		return err
	}
	s.CurrentIssue = id
	return s.Save()
}

func issuePath(id string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "issues", id+".yaml")
}


func LoadIssue(id string) (*Issue, error) {
	data, err := os.ReadFile(issuePath(id))
	if err != nil {
		return nil, err
	}
	var issue Issue
	return &issue, yaml.Unmarshal(data, &issue)
}

func LoadIssueByIDOrName(wsName, query string) (*Issue, error) {
	if issue, err := LoadIssue(query); err == nil && issue.Workspace == wsName {
		return issue, nil
	}
	issues, err := LoadIssuesForWorkspace(wsName)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		if strings.EqualFold(issue.Name, query) {
			return issue, nil
		}
	}
	return nil, fmt.Errorf("issue %q not found", query)
}

func (i *Issue) Save() error {
	path := issuePath(i.ID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(i)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadIssuesForWorkspace(wsName string) ([]*Issue, error) {
	base, _ := os.UserConfigDir()
	dir := filepath.Join(base, "repofleet", "issues")
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
		issue, err := LoadIssue(strings.TrimSuffix(e.Name(), ".yaml"))
		if err != nil {
			continue
		}
		if issue.Workspace == wsName {
			issues = append(issues, issue)
		}
	}
	return issues, nil
}

func DeleteIssue(wsName, id string) error {
	if err := os.Remove(issuePath(id)); err != nil {
		return err
	}
	if CurrentIssueID(wsName) == id {
		_ = SetCurrentIssue(wsName, "")
	}
	return nil
}
