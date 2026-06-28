package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Forge string `yaml:"forge"` // "github" | "gitlab"
	URL   string `yaml:"url"`
}

type Workspace struct {
	Name  string `yaml:"name"`
	Repos []Repo `yaml:"repos"`
}

type Config struct {
	CurrentWorkspace string      `yaml:"current_workspace"`
	Workspaces       []Workspace `yaml:"workspaces"`
}

func configPath() string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "repofleet", "config.yaml")
}

func Load() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{CurrentWorkspace: "default", Workspaces: []Workspace{{Name: "default"}}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c *Config) CurrentWS() *Workspace {
	for i := range c.Workspaces {
		if c.Workspaces[i].Name == c.CurrentWorkspace {
			return &c.Workspaces[i]
		}
	}
	// auto-create if missing
	c.Workspaces = append(c.Workspaces, Workspace{Name: c.CurrentWorkspace})
	return &c.Workspaces[len(c.Workspaces)-1]
}

func (c *Config) AddRepo(wsName string, repo Repo) {
	for i := range c.Workspaces {
		if c.Workspaces[i].Name == wsName {
			c.Workspaces[i].Repos = append(c.Workspaces[i].Repos, repo)
			return
		}
	}
	c.Workspaces = append(c.Workspaces, Workspace{Name: wsName, Repos: []Repo{repo}})
}

func (c *Config) RemoveRepo(wsName, repoName string) bool {
	for i := range c.Workspaces {
		if c.Workspaces[i].Name != wsName {
			continue
		}
		repos := c.Workspaces[i].Repos
		for j, r := range repos {
			if r.Name == repoName {
				c.Workspaces[i].Repos = append(repos[:j], repos[j+1:]...)
				return true
			}
		}
	}
	return false
}
