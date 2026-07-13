package giturl

import (
	"regexp"
	"strings"
)

var (
	sshRe   = regexp.MustCompile(`^(?:ssh://)?git@([^:/]+)[:/](.+)$`)
	httpsRe = regexp.MustCompile(`^https?://(?:[^@/]+@)?([^/]+)/(.+)$`)
)

func Normalize(raw string) string {
	host, path, ok := parse(raw)
	if !ok {
		return raw
	}
	return "git@" + host + ":" + path
}

func parse(raw string) (host, path string, ok bool) {
	if m := sshRe.FindStringSubmatch(raw); m != nil {
		return m[1], cleanPath(m[2]), true
	}
	if m := httpsRe.FindStringSubmatch(raw); m != nil {
		return m[1], cleanPath(m[2]), true
	}
	return "", "", false
}

func cleanPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if !strings.HasSuffix(path, ".git") {
		path += ".git"
	}
	return path
}
