package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func workspacePath(name string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "workspaces", name+".yaml")
}

func currentWorkspaceFile() string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "workspaces", ".current")
}

func currentIssueFile(wsName string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "workspaces", wsName+".current")
}

func CurrentWorkspaceName() string {
	data, err := os.ReadFile(currentWorkspaceFile())
	if err != nil {
		return "default"
	}
	return strings.TrimSpace(string(data))
}

func SetCurrentWorkspace(name string) error {
	return os.WriteFile(currentWorkspaceFile(), []byte(name), 0o644)
}

func CurrentIssueID(wsName string) string {
	data, err := os.ReadFile(currentIssueFile(wsName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SetCurrentIssue(wsName, id string) error {
	if id == "" {
		_ = os.Remove(currentIssueFile(wsName))
		return nil
	}
	return os.WriteFile(currentIssueFile(wsName), []byte(id), 0o644)
}

func LoadWorkspace(name string) (*Workspace, error) {
	data, err := os.ReadFile(workspacePath(name))
	if os.IsNotExist(err) {
		return &Workspace{}, nil
	}
	if err != nil {
		return nil, err
	}
	var ws Workspace
	return &ws, yaml.Unmarshal(data, &ws)
}

func DeleteWorkspace(name string) error {
	_ = os.Remove(currentIssueFile(name))
	path := workspacePath(name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workspace %q not found", name)
		}
		return err
	}
	return nil
}

func (w *Workspace) Save() error {
	path := workspacePath(w.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(w)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadWorkspaces() ([]*Workspace, error) {
	base, _ := os.UserConfigDir()
	dir := filepath.Join(base, "repofleet", "workspaces")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var workspaces []*Workspace
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		ws, err := LoadWorkspace(strings.TrimSuffix(e.Name(), ".yaml"))
		if err != nil {
			continue
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (w *Workspace) AddRepo(repo Repo) {
	w.Repos = append(w.Repos, repo)
}

func (w *Workspace) RemoveRepo(name string) bool {
	for i, r := range w.Repos {
		if r.Name == name {
			w.Repos = append(w.Repos[:i], w.Repos[i+1:]...)
			return true
		}
	}
	return false
}
