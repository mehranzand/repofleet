package store

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func issuePath(id string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "issues", id+".yaml")
}

func activePath(wsName string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "active", wsName)
}

func LoadIssue(id string) (*Issue, error) {
	data, err := os.ReadFile(issuePath(id))
	if err != nil {
		return nil, err
	}
	var issue Issue
	return &issue, yaml.Unmarshal(data, &issue)
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

func CurrentIssueID(wsName string) string {
	data, err := os.ReadFile(activePath(wsName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SetCurrentIssue(wsName, id string) error {
	path := activePath(wsName)
	if id == "" {
		_ = os.Remove(path)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(id), 0o644)
}
