package snapshot

import (
	"fmt"
	"strings"

	"github.com/mehranzand/repofleet/internal/util/git"
)

type GitCommand struct {
	Runner   *git.Runner
	RepoPath string
	Args     []string
	UndoArgs []string
}

func (c *GitCommand) Execute() error {
	_, err := runOne(c.Runner, c.RepoPath, c.Args...)
	return err
}

func (c *GitCommand) Undo() error {
	if c.UndoArgs == nil {
		return nil
	}
	_, err := runOne(c.Runner, c.RepoPath, c.UndoArgs...)
	return err
}

func (c *GitCommand) Describe() string {
	return fmt.Sprintf("git %s (%s)", strings.Join(c.Args, " "), c.RepoPath)
}
