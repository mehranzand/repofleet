package iostreams

import (
	"os"
	"strings"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	green     = "\033[32m"
	boldGreen = "\033[1;32m"
	cyan      = "\033[36m"
)

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func Green(s string) string {
	if !colorEnabled() {
		return s
	}
	return green + s + reset
}

func Cyan(s string) string {
	if !colorEnabled() {
		return s
	}
	return cyan + s + reset
}

func BoldGreen(s string) string {
	if !colorEnabled() {
		return s
	}
	return boldGreen + s + reset
}

func Dim(s string) string {
	if !colorEnabled() {
		return s
	}
	return dim + s + reset
}

// Logo returns the repofleet branded logo string for terminal output.
// version is shown after the wordmark (pass empty string to omit).
func Logo(version string) string {
	ver := ""
	if version != "" && version != "dev" {
		ver = " " + version
	}

	if !colorEnabled() {
		return "█▀█ █▀▀\n█▀▄ █▀▀   ⟫ repofleet" + ver + "\n█ █ █     issue-centered multi-repo workflow"
	}

	bg := boldGreen
	cy := cyan
	di := dim
	rs := reset

	wordmark := bg + "⟫ " + cy + "repo" + bg + "fleet" + rs + di + ver + rs
	tagline  := di + "issue-centered multi-repo workflow" + rs

	return bg + "█▀█ █▀▀" + rs + "\n" +
		bg + "█▀▄ █▀▀" + rs + "   " + wordmark + "\n" +
		bg + "█ █ █  " + rs + "   " + tagline
}

func Bold(s string) string {
	if !colorEnabled() {
		return s
	}
	return bold + s + reset
}

// ColorizeFlags colorizes the -flag and --flag portions of cobra flag usage lines.
func ColorizeFlags(usage string) string {
	if !colorEnabled() {
		return usage
	}
	lines := strings.Split(usage, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmed, "-") {
			// find where the description starts (two or more spaces after flags)
			idx := strings.Index(line, "   ")
			if idx > 0 {
				flagPart := line[:idx]
				rest := line[idx:]
				lines[i] = cyan + flagPart + reset + rest
			}
		}
	}
	return strings.Join(lines, "\n")
}
