package giturl

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"git@github.com:mehranzand/repofleet.git", "git@github.com:mehranzand/repofleet.git"},
		{"https://github.com/mehranzand/pulseup.git", "git@github.com:mehranzand/pulseup.git"},
		{"https://github.com/mehranzand/pulseup", "git@github.com:mehranzand/pulseup.git"},
		{"git@gitlab.com:group/sub/repo", "git@gitlab.com:group/sub/repo.git"},
		{"ssh://git@github.com/owner/repo.git", "git@github.com:owner/repo.git"},
		{"", ""},
		{"not-a-url", "not-a-url"},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
