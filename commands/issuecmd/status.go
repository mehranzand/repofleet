package issuecmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func countLines(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	n := bytes.Count(b, []byte("\n"))
	if b[len(b)-1] != '\n' {
		n++
	}
	return n
}

func openTTY() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("windows: use default stdio")
	}
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

type repoItem struct {
	Name         string
	Branch       string
	SwitchBranch string
	Current      bool
	Missing      bool
	path         string
}

func newStatusCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status dashboard for all repos in the current issue",
		Long: "Show the current branch and uncommitted changes for every repo in the active issue context.\n\n" +
			"HEAD± counts additions/deletions from both tracked changes and untracked (new) files.\n" +
			"Repos whose path no longer exists on disk show a red ! marker instead of status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			hash := store.CurrentIssueHash(f.Workspace.Name)
			if hash == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch <id>")
			}

			ctx, err := store.LoadIssueByHash(f.Workspace.Name, hash)
			if err != nil {
				return err
			}

			paths := repoPaths(ctx.Repos)
			branchResults := f.GitRunner.Run(paths, "rev-parse", "--abbrev-ref", "HEAD")
			statusResults := f.GitRunner.Run(paths, "status", "--short")
			commitResults := f.GitRunner.Run(paths, "rev-parse", "--short", "HEAD")
			timeResults := f.GitRunner.Run(paths, "log", "-1", "--format=%ct")
			diffResults := f.GitRunner.Run(paths, "diff", "HEAD", "--numstat")
			untrackedResults := f.GitRunner.Run(paths, "ls-files", "--others", "--exclude-standard")

			type repoData struct {
				branch  string
				commit  string
				age     string
				added   int
				deleted int
				clean   bool
				missing bool
			}

			data := make([]repoData, len(ctx.Repos))
			for i := range ctx.Repos {
				d := &data[i]
				if _, err := os.Stat(ctx.Repos[i].Path); err != nil {
					d.missing = true
					continue
				}
				d.branch = "?"
				if branchResults[i].Err == nil {
					d.branch = strings.TrimSpace(branchResults[i].Stdout)
				}
				if commitResults[i].Err == nil {
					d.commit = strings.TrimSpace(commitResults[i].Stdout)
				}
				if timeResults[i].Err == nil {
					if ts, err := strconv.ParseInt(strings.TrimSpace(timeResults[i].Stdout), 10, 64); err == nil {
						d.age = iostreams.RelativeAge(time.Since(time.Unix(ts, 0)))
					}
				}
				d.clean = statusResults[i].Err == nil && strings.TrimSpace(statusResults[i].Stdout) == ""
				if !d.clean && diffResults[i].Err == nil {
					for _, line := range strings.Split(strings.TrimSpace(diffResults[i].Stdout), "\n") {
						parts := strings.Fields(line)
						if len(parts) >= 2 {
							if n, e := strconv.Atoi(parts[0]); e == nil {
								d.added += n
							}
							if n, e := strconv.Atoi(parts[1]); e == nil {
								d.deleted += n
							}
						}
					}
				}
				if !d.clean && untrackedResults[i].Err == nil {
					for _, rel := range strings.Split(strings.TrimSpace(untrackedResults[i].Stdout), "\n") {
						rel = strings.TrimSpace(rel)
						if rel == "" {
							continue
						}
						content, err := os.ReadFile(filepath.Join(ctx.Repos[i].Path, rel))
						if err != nil {
							continue
						}
						d.added += countLines(content)
					}
				}
			}

			fmt.Fprintf(f.IO.Out, "%s %s %s %s\n\n",
				iostreams.Dim("Repos for issue"), iostreams.Cyan("#"+ctx.ID),
				iostreams.Dim("on branch"), iostreams.Cyan(ctx.BranchSlug),
			)

			t := iostreams.NewTable()
			t.AddField("", iostreams.Dim)
			t.AddField("Repo", iostreams.Dim)
			t.AddField("Checkout", iostreams.Dim)
			t.AddField("Commit", iostreams.Dim)
			t.AddField("Age", iostreams.Dim)
			t.AddField("HEAD±", iostreams.Dim)
			t.EndRow()

			for i, r := range ctx.Repos {
				d := data[i]
				if d.missing {
					t.AddField("!", iostreams.Red)
					t.AddField(r.Name, iostreams.Cyan)
					t.AddField("-", iostreams.Dim)
					t.AddField("-", iostreams.Dim)
					t.AddField("-", iostreams.Dim)
					t.AddField(iostreams.Red("path missing"), nil)
					t.EndRow()
					continue
				}
				var headPM string
				if d.clean {
					headPM = iostreams.Dim("clean")
				} else {
					headPM = iostreams.Green(fmt.Sprintf("↑%d", d.added)) + " " + iostreams.Red(fmt.Sprintf("↓%d", d.deleted))
				}
				t.AddField("", iostreams.Dim)
				t.AddField(r.Name, iostreams.Cyan)
				t.AddField(d.branch, iostreams.Dim)
				t.AddField(d.commit, iostreams.Dim)
				t.AddField(d.age, iostreams.Dim)
				t.AddField(headPM, nil)
				t.EndRow()
			}

			t.Render(f.IO.Out)
			return nil
		},
	}
}
