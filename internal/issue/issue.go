package issue

import (
	"os"
	"path/filepath"

	"github.com/mehranzand/repofleet/internal/config"
	"gopkg.in/yaml.v3"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

type Context struct {
	ID         string       `yaml:"id"`
	BranchSlug string       `yaml:"branch_slug"`
	Repos      []config.Repo `yaml:"repos"`
	Status     Status       `yaml:"status"`
}

func statePath(id string) string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "issues", id+".yaml")
}

func Load(id string) (*Context, error) {
	data, err := os.ReadFile(statePath(id))
	if err != nil {
		return nil, err
	}
	var ctx Context
	return &ctx, yaml.Unmarshal(data, &ctx)
}

func (c *Context) Save() error {
	path := statePath(c.ID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func CurrentID() string {
	base, _ := os.UserConfigDir()
	data, err := os.ReadFile(filepath.Join(base, "repofleet", "current_issue"))
	if err != nil {
		return ""
	}
	return string(data)
}

func SetCurrent(id string) error {
	base, _ := os.UserConfigDir()
	dir := filepath.Join(base, "repofleet")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "current_issue"), []byte(id), 0o644)
}
