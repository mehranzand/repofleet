package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mehranzand/repofleet/internal/giturl"
)

func workspacesDir() string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "workspaces")
}

func workspacePath(name string) string {
	return filepath.Join(workspacesDir(), name+".yaml")
}

func currentWorkspaceFile() string {
	return filepath.Join(workspacesDir(), ".current")
}

func currentIssueFile(wsName string) string {
	return filepath.Join(workspacesDir(), wsName+".current")
}

func readPointerFile(path, def string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return def
	}
	return strings.TrimSpace(string(data))
}

func CurrentWorkspaceName() string {
	return readPointerFile(currentWorkspaceFile(), "default")
}

func SetCurrentWorkspace(name string) error {
	return os.WriteFile(currentWorkspaceFile(), []byte(name), 0o644)
}

func CurrentIssueHash(wsName string) string {
	return readPointerFile(currentIssueFile(wsName), "")
}

func SetCurrentIssueHash(wsName, hash string) error {
	path := currentIssueFile(wsName)
	if hash == "" {
		_ = os.Remove(path)
		return nil
	}
	return os.WriteFile(path, []byte(hash), 0o644)
}

func LoadWorkspace(name string) (*Workspace, error) {
	data, err := os.ReadFile(workspacePath(name))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := yaml.Unmarshal(data, &ws); err != nil {
		return nil, err
	}
	for i := range ws.Repos {
		ws.Repos[i].Remote = giturl.Normalize(ws.Repos[i].Remote)
	}
	return &ws, nil
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
	dir := workspacesDir()
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
		if err != nil || ws == nil {
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
