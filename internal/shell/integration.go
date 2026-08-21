package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const (
	Bash       = "bash"
	Zsh        = "zsh"
	Fish       = "fish"
	PowerShell = "powershell"
)

const markerPrefix = "# repofleet-shell-integration"

const integrationVersion = 3

func startMarker(v int) string {
	return fmt.Sprintf("%s-start v%d", markerPrefix, v)
}

const endMarker = markerPrefix + "-end"

func block(body string) string {
	return startMarker(integrationVersion) + body + "\n" + endMarker
}

var legacyMarkerRe = regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(markerPrefix) + `(?: v(\d+))?$`)

var blockRe = regexp.MustCompile(`(?ms)^` + regexp.QuoteMeta(markerPrefix) + `-start v(\d+)$.*?^` + regexp.QuoteMeta(endMarker) + `$`)

// installedVersion returns the highest shell-integration version marker
func installedVersion(content string) int {
	max := -1
	for _, m := range legacyMarkerRe.FindAllStringSubmatch(content, -1) {
		v := 0
		if m[1] != "" {
			if n, err := strconv.Atoi(m[1]); err == nil {
				v = n
			}
		}
		if v > max {
			max = v
		}
	}
	for _, m := range blockRe.FindAllStringSubmatch(content, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil && n > max {
			max = n
		}
	}
	return max
}

var bashZshSnippet = block(`
rf() {
  if [[ "$1" == "issue" && ( "$2" == "goto" || "$2" == "switch" ) ]]; then
    local tmp p
    tmp=$(mktemp) || return 1
    command rf issue "$2" --out "$tmp" "${@:3}"
    p=$(cat "$tmp" 2>/dev/null)
    rm -f "$tmp"
    [ -n "$p" ] && cd "$p"
  else
    command rf "$@"
  fi
}`)

var fishSnippet = block(`
function rf
  if test "$argv[1]" = "issue" -a \( "$argv[2]" = "goto" -o "$argv[2]" = "switch" \)
    set tmp (mktemp)
    command rf issue $argv[2] --out $tmp $argv[3..]
    set p (cat $tmp 2>/dev/null)
    rm -f $tmp
    test -n "$p" && cd $p
  else
    command rf $argv
  end
end`)

var powershellSnippet = block(`
function rf {
  $bin = (Get-Command rf -CommandType Application -ErrorAction SilentlyContinue).Source
  if (-not $bin) { Write-Error "rf binary not found in PATH"; return }
  if ($args[0] -eq "issue" -and ($args[1] -eq "goto" -or $args[1] -eq "switch")) {
    $tmp = [System.IO.Path]::GetTempFileName()
    $rest = $args | Select-Object -Skip 2
    & $bin issue $args[1] --out $tmp @rest
    $p = Get-Content $tmp -ErrorAction SilentlyContinue
    Remove-Item $tmp -ErrorAction SilentlyContinue
    if ($p) { Set-Location $p }
  } else {
    & $bin @args
  }
}`)

type shellDef struct {
	snippet string
	rcFile  func(home string) string
}

var shellDefs = map[string]shellDef{
	Bash:       {bashZshSnippet, func(home string) string { return filepath.Join(home, ".bashrc") }},
	Zsh:        {bashZshSnippet, func(home string) string { return filepath.Join(home, ".zshrc") }},
	Fish:       {fishSnippet, func(home string) string { return filepath.Join(home, ".config", "fish", "config.fish") }},
	PowerShell: {powershellSnippet, func(_ string) string { return psProfilePath() }},
}

var aliases = map[string]string{"pwsh": PowerShell}

// psProfilePath queries PowerShell for the real $PROFILE path
func psProfilePath() string {
	out, err := exec.Command("powershell", "-NoProfile", "-Command", "$PROFILE").Output()
	if err == nil {
		if p := strings.TrimSpace(string(out)); p != "" {
			return p
		}
	}
	// fallback to PS5 conventional path
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
}

func normalize(sh string) string {
	if canon, ok := aliases[sh]; ok {
		return canon
	}
	return sh
}

func Detect() string {
	if runtime.GOOS == "windows" {
		return PowerShell
	}
	switch strings.ToLower(filepath.Base(os.Getenv("SHELL"))) {
	case Fish:
		return Fish
	case Zsh:
		return Zsh
	default:
		return Bash
	}
}

func Snippet(sh string) (string, error) {
	def, ok := shellDefs[normalize(sh)]
	if !ok {
		return "", fmt.Errorf("unknown shell %q — supported: bash, zsh, fish, powershell", sh)
	}
	return def.snippet, nil
}

func RCFile(sh string) (string, error) {
	def, ok := shellDefs[normalize(sh)]
	if !ok {
		return "", fmt.Errorf("unknown shell %q", sh)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return def.rcFile(home), nil
}

func installTo(snip, rcPath string) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
		return false, err
	}
	existing, _ := os.ReadFile(rcPath)
	content := string(existing)

	if installedVersion(content) >= integrationVersion {
		return false, nil
	}

	var newContent string
	if loc := blockRe.FindStringIndex(content); loc != nil {
		newContent = content[:loc[0]] + snip + content[loc[1]:]
	} else {
		newContent = content + "\n" + snip + "\n"
	}

	if err := os.WriteFile(rcPath, []byte(newContent), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func Install(sh string) (installed bool, rcPath string, err error) {
	sh = normalize(sh)
	def, ok := shellDefs[sh]
	if !ok {
		return false, "", fmt.Errorf("unknown shell %q", sh)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false, "", err
	}
	rcPath = def.rcFile(home)

	ok2, err := installTo(def.snippet, rcPath)
	return ok2, rcPath, err
}
