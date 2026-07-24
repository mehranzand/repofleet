package store

import "time"

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
	Hash             string          `yaml:"hash"`
	CreatedAt        time.Time       `yaml:"created_at"`
	Name             string          `yaml:"name,omitempty"`
	ShortDescription string          `yaml:"short_description,omitempty"`
	Kind             IssueKind       `yaml:"kind,omitempty"`
	ChangeType       IssueChangeType `yaml:"change_type,omitempty"`
	Workspace        string          `yaml:"workspace"`
	BranchSlug       string          `yaml:"branch_slug"`
	Repos            []Repo          `yaml:"repos"`
	Status           IssueStatus     `yaml:"status"`
}

type RepoSnapshot struct {
	Path            string   `yaml:"path"`
	Branch          string   `yaml:"branch"`
	BaseSHA         string   `yaml:"base_sha"`
	StagedPatch     string   `yaml:"staged_patch,omitempty"`
	UnstagedPatch   string   `yaml:"unstaged_patch,omitempty"`
	UntrackedDir    string   `yaml:"untracked_dir,omitempty"`
	UntrackedFiles  []string `yaml:"untracked_files,omitempty"`
	ConflictedDir   string   `yaml:"conflicted_dir,omitempty"`
	ConflictedFiles []string `yaml:"conflicted_files,omitempty"`
}

type Snapshot struct {
	Hash      string         `yaml:"hash"`
	IssueID   string         `yaml:"issue_id"`
	IssueHash string         `yaml:"issue_hash"`
	Workspace string         `yaml:"workspace"`
	CreatedAt string         `yaml:"created_at"`
	Name      string         `yaml:"name,omitempty"`
	Repos     []RepoSnapshot `yaml:"repos"`
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
