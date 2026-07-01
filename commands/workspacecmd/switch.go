package workspacecmd

import (
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/mehranzand/repofleet/commands/factory"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/spf13/cobra"
)

type workspaceItem struct {
	Name    string
	Repos   string
	Current bool
}

func newSwitchCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "switch [name]",
		Short: "Switch to a workspace, or create one if it doesn't exist",
		Long:  "Switch workspaces interactively, or pass a name to switch directly. creates the workspace if it doesn't exist.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string

			if len(args) == 1 {
				name = args[0]
			} else {
				selected, err := promptWorkspace(f)
				if err != nil {
					return err
				}
				if selected == "" {
					return nil
				}
				name = selected
			}

			if name == f.Settings.CurrentWorkspace {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Dim(fmt.Sprintf("Already in workspace %q", name)))
				return nil
			}

			ws, err := store.LoadWorkspace(name)
			if err != nil {
				return err
			}

			created := ws.Name == ""
			if created {
				ws = &store.Workspace{Name: name}
				if err := ws.Save(); err != nil {
					return err
				}
			}

			f.Settings.CurrentWorkspace = name
			if err := f.Settings.Save(); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Created and switched to workspace %q", name)))
			} else {
				fmt.Fprintf(f.IO.Out, "%s\n", iostreams.Green("✓")+" "+iostreams.Cyan(fmt.Sprintf("Switched to workspace %q", name)))
			}

			if len(ws.Repos) > 0 {
				fmt.Fprintln(f.IO.Out)
				iostreams.PrintRepos(f.IO.Out, ws.Repos)
			}
			return nil
		},
	}
}

func promptWorkspace(f *factory.Factory) (string, error) {
	workspaces, err := store.LoadWorkspaces()
	if err != nil {
		return "", err
	}

	if len(workspaces) == 0 {
		prompt := promptui.Prompt{
			Label: "No workspaces found. Enter a name to create one",
		}
		result, err := prompt.Run()
		if err != nil || result == "" {
			return "", nil
		}
		return result, nil
	}

	items := make([]workspaceItem, len(workspaces))
	for i, ws := range workspaces {
		items[i] = workspaceItem{
			Name:    ws.Name,
			Repos:   strconv.Itoa(len(ws.Repos)) + " repo(s)",
			Current: ws.Name == f.Settings.CurrentWorkspace,
		}
	}

	activeRow := `▸ ` +
		`{{ if .Current }}{{ .Name | green }}{{ else }}{{ .Name | cyan }}{{ end }}` +
		`  {{ .Repos | faint }}` +
		`{{ if .Current }}  {{ "*" | green }}{{ end }}`

	inactiveRow := `  {{ .Name }}` +
		`  {{ .Repos | faint }}` +
		`{{ if .Current }}  {{ "*" | green }}{{ end }}`

	templates := &promptui.SelectTemplates{
		Label:    `{{ "Select workspace:" | faint }}`,
		Active:   activeRow,
		Inactive: inactiveRow,
		Selected: `{{ "✓" | green }} {{ .Name | cyan }}`,
		Help:     `{{ "↑↓ navigate  ↵ select" | faint }}`,
	}

	prompt := promptui.Select{
		Label:     "Select workspace",
		Items:     items,
		Templates: templates,
	}
	i, _, err := prompt.Run()
	if err != nil {
		return "", nil
	}
	return items[i].Name, nil
}
