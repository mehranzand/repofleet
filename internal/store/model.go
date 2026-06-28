package store

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

// Config is the top-level global config persisted to ~/.config/repofleet/config.yaml.
type Config struct {
	CurrentWorkspace string      `yaml:"current_workspace"`
	Workspaces       []Workspace `yaml:"workspaces"`
}

// IssueStatus represents the lifecycle state of an issue context.
type IssueStatus string

const (
	IssueStatusActive   IssueStatus = "active"
	IssueStatusArchived IssueStatus = "archived"
)

// Issue is a cross-repo issue context persisted to ~/.config/repofleet/issues/<id>.yaml.
type Issue struct {
	ID         string      `yaml:"id"`
	BranchSlug string      `yaml:"branch_slug"`
	Repos      []Repo      `yaml:"repos"`
	Status     IssueStatus `yaml:"status"`
}
