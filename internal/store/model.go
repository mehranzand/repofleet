package store

type Workspace struct {
	Name          string `yaml:"name"`
	Repos         []Repo `yaml:"repos"`
	BranchPattern string `yaml:"branch_pattern,omitempty"`
}

type Repo struct {
	Name   string    `yaml:"name"`
	Path   string    `yaml:"path"`
	Forge  RepoForge `yaml:"forge"`
	Remote string    `yaml:"remote"`
}

type Issue struct {
	ID               string          `yaml:"id"`
	Name             string          `yaml:"name,omitempty"`
	ShortDescription string          `yaml:"short_description,omitempty"`
	Kind             IssueKind       `yaml:"kind,omitempty"`
	ChangeType       IssueChangeType `yaml:"change_type,omitempty"`
	Workspace        string          `yaml:"workspace"`
	BranchSlug       string          `yaml:"branch_slug"`
	Repos            []Repo          `yaml:"repos"`
	Status           IssueStatus     `yaml:"status"`
}

// enums
type RepoForge string

const (
	RepoForgeGitHub RepoForge = "github"
	RepoForgeGitLab RepoForge = "gitlab"
)

type IssueStatus string

const (
	IssueStatusActive   IssueStatus = "active"
	IssueStatusArchived IssueStatus = "archived"
)

type IssueKind string

const (
	IssueKindBug     IssueKind = "bug"
	IssueKindFeature IssueKind = "feature"
	IssueKindTask    IssueKind = "task"
	IssueKindStory   IssueKind = "story"
)

type IssueChangeType string

const (
	IssueChangeTypeFeat     IssueChangeType = "feat"
	IssueChangeTypeFix      IssueChangeType = "fix"
	IssueChangeTypeChore    IssueChangeType = "chore"
	IssueChangeTypeDocs     IssueChangeType = "docs"
	IssueChangeTypeRefactor IssueChangeType = "refactor"
	IssueChangeTypeTest     IssueChangeType = "test"
)
