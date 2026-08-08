package util

import (
	"bytes"
	"io"
)

// silentBellWriter drops ASCII BEL (0x07) bytes before they reach the terminal 
type silentBellWriter struct {
	io.Writer
}

func (w silentBellWriter) Write(p []byte) (int, error) {
	if !bytes.ContainsRune(p, 0x07) {
		return w.Writer.Write(p)
	}
	if _, err := w.Writer.Write(bytes.ReplaceAll(p, []byte{0x07}, nil)); err != nil {
		return 0, err
	}
	
	return len(p), nil
}

func (silentBellWriter) Close() error { return nil }

// SilenceBell wraps w for use as a promptui Stdout, stripping the terminal.
func SilenceBell(w io.Writer) io.WriteCloser {
	return silentBellWriter{w}
}
