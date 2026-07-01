package store

import (
	"os"
	"path/filepath"
)

func Initialize() error {
	base, _ := os.UserConfigDir()
	dirs := []string{
		filepath.Join(base, "repofleet", "workspaces"),
		filepath.Join(base, "repofleet", "issues"),
		filepath.Join(base, "repofleet", "active"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(settingsPath()); os.IsNotExist(err) {
		s := &Settings{CurrentWorkspace: "default"}
		if err := s.Save(); err != nil {
			return err
		}
	}

	defaultPath := filepath.Join(base, "repofleet", "workspaces", "default.yaml")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		ws := &Workspace{Name: "default"}
		if err := ws.Save(); err != nil {
			return err
		}
	}

	return nil
}
