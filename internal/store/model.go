package store

type Settings struct {
	CurrentWorkspace string `yaml:"current_workspace"`
	CurrentIssue     string `yaml:"current_issue,omitempty"`
}

type Workspace struct {
	Name          string `yaml:"name"`
	Repos         []Repo `yaml:"repos"`
	BranchPattern string `yaml:"branch_pattern,omitempty"`
}

type Repo struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Forge Forge  `yaml:"forge"`
	URL   string `yaml:"url"`
}

type Issue struct {
	ID               string      `yaml:"id"`
	Name             string      `yaml:"name,omitempty"`
	ShortDescription string      `yaml:"short_description,omitempty"`
	Workspace        string      `yaml:"workspace"`
	BranchSlug       string      `yaml:"branch_slug"`
	Repos            []Repo      `yaml:"repos"`
	Status           IssueStatus `yaml:"status"`
}

// enums
type Forge string
const (
	ForgeGitHub Forge = "github"
	ForgeGitLab Forge = "gitlab"
)

type IssueStatus string
const (
	IssueStatusActive   IssueStatus = "active"
	IssueStatusArchived IssueStatus = "archived"
)