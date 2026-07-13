package snapshot

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type WriteFileCmd struct {
	Path string
	Data []byte
}

func (c *WriteFileCmd) Execute() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.Path, c.Data, 0o644)
}

func (c *WriteFileCmd) Undo() error {
	err := os.Remove(c.Path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (c *WriteFileCmd) Describe() string {
	return fmt.Sprintf("write %s", c.Path)
}

type CopyFileCmd struct {
	Src string
	Dst string
}

func (c *CopyFileCmd) Execute() error {
	if err := os.MkdirAll(filepath.Dir(c.Dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(c.Src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(c.Dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func (c *CopyFileCmd) Undo() error {
	err := os.Remove(c.Dst)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (c *CopyFileCmd) Describe() string {
	return fmt.Sprintf("copy %s -> %s", c.Src, c.Dst)
}
