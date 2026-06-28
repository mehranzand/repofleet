package iostreams

import (
	"fmt"
	"io"
	"strings"
)

type cell struct {
	value   string
	colorFn func(string) string
}

type Table struct {
	rows    [][]cell
	current []cell
}

func NewTable() *Table { return &Table{} }

func (t *Table) AddField(value string, colorFn func(string) string) {
	t.current = append(t.current, cell{value, colorFn})
}

func (t *Table) EndRow() {
	if len(t.current) > 0 {
		t.rows = append(t.rows, t.current)
		t.current = nil
	}
}

func (t *Table) Render(w io.Writer) {
	if len(t.rows) == 0 {
		return
	}

	// compute max width per column
	cols := 0
	for _, row := range t.rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	widths := make([]int, cols)
	for _, row := range t.rows {
		for j, c := range row {
			if len(c.value) > widths[j] {
				widths[j] = len(c.value)
			}
		}
	}

	for _, row := range t.rows {
		fmt.Fprint(w, "  ")
		for j, c := range row {
			isLast := j == len(row)-1
			val := c.value
			if !isLast {
				val = val + strings.Repeat(" ", widths[j]-len(val))
			}
			if c.colorFn != nil {
				val = c.colorFn(val)
			}
			fmt.Fprint(w, val)
			if !isLast {
				fmt.Fprint(w, "  ")
			}
		}
		fmt.Fprintln(w)
	}
}
