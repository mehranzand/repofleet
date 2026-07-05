package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	Bash       = "bash"
	Zsh        = "zsh"
	Fish       = "fish"
	PowerShell = "powershell"
)

const InstallMarker = "# repofleet-shell-integration"

const bashZshSnippet = InstallMarker + `
rf() {
  if [[ "$1" == "issue" && "$2" == "status" ]]; then
    local tmp p
    tmp=$(mktemp) || return 1
    command rf issue status --go-to --out "$tmp" "${@:3}"
    p=$(cat "$tmp" 2>/dev/null)
    rm -f "$tmp"
    [ -n "$p" ] && cd "$p"
  else
    command rf "$@"
  fi
}`

const fishSnippet = InstallMarker + `
function rf
  if test "$argv[1]" = "issue" -a "$argv[2]" = "status"
    set tmp (mktemp)
    command rf issue status --go-to --out $tmp $argv[3..]
    set p (cat $tmp 2>/dev/null)
    rm -f $tmp
    test -n "$p" && cd $p
  else
    command rf $argv
  end
end`

const powershellSnippet = InstallMarker + `
function rf {
  $bin = (Get-Command rf -CommandType Application -ErrorAction SilentlyContinue).Source
  if (-not $bin) { Write-Error "rf binary not found in PATH"; return }
  if ($args[0] -eq "issue" -and $args[1] -eq "status") {
    $tmp = [System.IO.Path]::GetTempFileName()
    & $bin issue status --go-to --out $tmp @($args | Select-Object -Skip 2)
    $p = Get-Content $tmp -ErrorAction SilentlyContinue
    Remove-Item $tmp -ErrorAction SilentlyContinue
    if ($p) { Set-Location $p }
  } else {
    & $bin @args
  }
}`

type shellDef struct {
	snippet string
	rcFile  func(home string) string
}

var shellDefs = map[string]shellDef{
	Bash:       {bashZshSnippet, func(home string) string { return filepath.Join(home, ".bashrc") }},
	Zsh:        {bashZshSnippet, func(home string) string { return filepath.Join(home, ".zshrc") }},
	Fish:       {fishSnippet, func(home string) string { return filepath.Join(home, ".config", "fish", "config.fish") }},
	// PowerShell 7+
	PowerShell: {powershellSnippet, func(home string) string { return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1") }},
	// Windows PowerShell 5.x
	"powershell5": {powershellSnippet, func(home string) string { return filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1") }},
}

var aliases = map[string]string{"pwsh": PowerShell}

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
	if strings.Contains(string(existing), InstallMarker) {
		return false, nil
	}
	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n%s\n", snip)
	return err == nil, err
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

	// For PowerShell, also install to the PS5 profile (WindowsPowerShell)
	if sh == PowerShell {
		ps5def := shellDefs["powershell5"]
		ps5Path := ps5def.rcFile(home)
		installTo(def.snippet, ps5Path) //nolint — best effort, ignore error
	}

	ok2, err := installTo(def.snippet, rcPath)
	return ok2, rcPath, err
}
