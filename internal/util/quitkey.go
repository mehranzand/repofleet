package util

import "io"

type quitOnQReader struct {
	io.Reader
}

func (r quitOnQReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	for i := 0; i < n; i++ {
		if p[i] == 'q' {
			p[i] = 0x03
		}
	}
	return n, err
}

func (quitOnQReader) Close() error { return nil }

func QuitOnQ(r io.Reader) io.ReadCloser {
	return quitOnQReader{r}
}
