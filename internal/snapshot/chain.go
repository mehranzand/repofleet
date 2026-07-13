package snapshot

import "fmt"

// Command is one reversible step in building a snapshot on disk.
type Command interface {
	Execute() error
	Undo() error
	Describe() string
}

// Chain runs Commands in order, undoing prior ones on failure.
type Chain struct {
	cmds []Command
}

func NewChain(cmds ...Command) *Chain {
	return &Chain{cmds: cmds}
}

func (c *Chain) Add(cmd Command) *Chain {
	c.cmds = append(c.cmds, cmd)
	return c
}

// Display plan for --dry-run
func (c *Chain) Plan() []string {
	plan := make([]string, len(c.cmds))
	for i, cmd := range c.cmds {
		plan[i] = cmd.Describe()
	}
	return plan
}

func (c *Chain) Run() error {
	for i, cmd := range c.cmds {
		if err := cmd.Execute(); err != nil {
			for j := i - 1; j >= 0; j-- {
				if uerr := c.cmds[j].Undo(); uerr != nil {
					return fmt.Errorf("%s: %w (rollback also failed: %s)", cmd.Describe(), err, uerr)
				}
			}
			return fmt.Errorf("%s: %w", cmd.Describe(), err)
		}
	}
	return nil
}
