package store

import (
	"os"
	"path/filepath"
)

func Initialize() error {
	base, _ := os.UserConfigDir()
	workspacesDir := filepath.Join(base, "repofleet", "workspaces")
	dirs := []string{
		workspacesDir,
		filepath.Join(base, "repofleet", "issues"),
	}

	entries, _ := os.ReadDir(workspacesDir)
	preexisting := len(entries) > 0

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	sentinel := filepath.Join(base, "repofleet", ".initialized")
	if _, err := os.Stat(sentinel); os.IsNotExist(err) {
		if !preexisting {
			ws := &Workspace{Name: "default"}
			if err := ws.Save(); err != nil {
				return err
			}
		}
		if err := os.WriteFile(sentinel, nil, 0o644); err != nil {
			return err
		}
	}

	return nil
}
