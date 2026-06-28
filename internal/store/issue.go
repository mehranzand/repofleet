package store

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

func CurrentIssueID() string {
	base, _ := os.UserConfigDir()
	data, err := os.ReadFile(filepath.Join(base, "repofleet", "current_issue"))
	if err != nil {
		return ""
	}
	return string(data)
}

func SetCurrentIssue(id string) error {
	base, _ := os.UserConfigDir()
	dir := filepath.Join(base, "repofleet")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "current_issue"), []byte(id), 0o644)
}
