package issuecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

func relativeAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy", int(d.Hours()/(24*365)))
	}
}

func openTTY() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("windows: use default stdio")
	}
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

type repoItem struct {
	Name     string
	Branch   string
	path     string
}

func newStatusCmd(f *factory.Factory) *cobra.Command {
	var goTo bool
	var outFile string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status dashboard for all repos in the current issue",
		Long:  "Show the current branch and uncommitted changes for every repo in the active issue context.",
		RunE: func(cmd *cobra.Command, args []string) error {
			id := store.CurrentIssueID(f.Workspace.Name)
			if id == "" {
				return fmt.Errorf("no active issue — switch to one with: rf issue switch <id>")
			}

			ctx, err := store.LoadIssue(id)
			if err != nil {
				return err
			}

			paths := repoPaths(ctx.Repos)
			branchResults := f.GitRunner.Run(paths, "rev-parse", "--abbrev-ref", "HEAD")
			statusResults := f.GitRunner.Run(paths, "status", "--short")
			commitResults := f.GitRunner.Run(paths, "rev-parse", "--short", "HEAD")
			timeResults   := f.GitRunner.Run(paths, "log", "-1", "--format=%ct")
			diffResults   := f.GitRunner.Run(paths, "diff", "HEAD", "--numstat")

			type repoData struct {
				branch  string
				commit  string
				age     string
				added   int
				deleted int
				clean   bool
			}

			data := make([]repoData, len(ctx.Repos))
			for i := range ctx.Repos {
				d := &data[i]
				d.branch = "?"
				if branchResults[i].Err == nil {
					d.branch = strings.TrimSpace(branchResults[i].Stdout)
				}
				if commitResults[i].Err == nil {
					d.commit = strings.TrimSpace(commitResults[i].Stdout)
				}
				if timeResults[i].Err == nil {
					if ts, err := strconv.ParseInt(strings.TrimSpace(timeResults[i].Stdout), 10, 64); err == nil {
						d.age = relativeAge(time.Since(time.Unix(ts, 0)))
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
			}

			if !goTo {
				fmt.Fprintf(f.IO.Out, "%s %s %s %s\n\n",
					iostreams.Dim("Repos for issue"), iostreams.Cyan("#"+ctx.ID),
					iostreams.Dim("on branch"), iostreams.Cyan(ctx.BranchSlug),
				)

				t := iostreams.NewTable()
				t.AddField("Repo", iostreams.Dim)
				t.AddField("Checkout", iostreams.Dim)
				t.AddField("Commit", iostreams.Dim)
				t.AddField("Age", iostreams.Dim)
				t.AddField("HEAD±", iostreams.Dim)
				t.EndRow()

				for i, r := range ctx.Repos {
					d := data[i]
					var headPM string
					if d.clean {
						headPM = iostreams.Dim("clean")
					} else {
						headPM = iostreams.Green(fmt.Sprintf("↑%d", d.added)) + " " + iostreams.Red(fmt.Sprintf("↓%d", d.deleted))
					}
					t.AddField(r.Name, iostreams.Cyan)
					t.AddField(d.branch, iostreams.Dim)
					t.AddField(d.commit, iostreams.Dim)
					t.AddField(d.age, iostreams.Dim)
					t.AddField(headPM, nil)
					t.EndRow()
				}

				t.Render(f.IO.Out)
				return nil
			}

			// --- select mode ---
			maxName, maxBranch := 0, 0
			for i, r := range ctx.Repos {
				if len(r.Name) > maxName {
					maxName = len(r.Name)
				}
				if len(data[i].branch) > maxBranch {
					maxBranch = len(data[i].branch)
				}
			}

			items := make([]repoItem, len(ctx.Repos))
			for i, r := range ctx.Repos {
				d := data[i]
				items[i] = repoItem{
					Name:   r.Name + strings.Repeat(" ", maxName-len(r.Name)),
					Branch: d.branch + strings.Repeat(" ", maxBranch-len(d.branch)),
					path:   r.Path,
				}
			}

			templates := &promptui.SelectTemplates{
				Label:    `{{ "Go to repo:" | faint }}`,
				Active:   `> {{ .Name | cyan }}  {{ .Branch | faint }}`,
				Inactive: `  {{ .Name }}  {{ .Branch | faint }}`,
				Selected: `{{ "✓" | green }} {{ .Name | cyan }}`,
			}

			tty, ttyErr := openTTY()
			if ttyErr == nil {
				defer tty.Close()
			}

			prompt := promptui.Select{
				Label:     fmt.Sprintf("Issue %s", ctx.ID),
				Items:     items,
				Templates: templates,
				Size:      15,
			}
			if ttyErr == nil {
				prompt.Stdin = tty
				prompt.Stdout = tty
			}

			idx, _, err := prompt.Run()
			if err != nil {
				return nil // cancelled
			}

			absPath := items[idx].path
			cwd, _ := filepath.Abs(".")
			rel, relErr := filepath.Rel(cwd, absPath)
			if relErr != nil {
				rel = absPath
			}

			if outFile != "" {
				return os.WriteFile(outFile, []byte(rel), 0o600)
			}
			fmt.Fprintln(f.IO.Out, rel)
			return nil
		},
	}

	cmd.Flags().BoolVar(&goTo, "go-to", false, "interactive repo selector — cd to selection via shell wrapper")
	cmd.Flags().StringVar(&outFile, "out", "", "write selected path to this file (used by shell wrapper)")
	return cmd
}
