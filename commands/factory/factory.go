package factory

import (
	"github.com/mehranzand/repofleet/internal/git"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
)

type Factory struct {
	Settings  *store.Settings
	Workspace *store.Workspace
	GitRunner *git.Runner
	IO        *iostreams.IOStreams
}

func New() (*Factory, error) {
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	settings, err := store.LoadSettings()
	if err != nil {
		return nil, err
	}
	ws, err := store.LoadWorkspace(settings.CurrentWorkspace)
	if err != nil {
		return nil, err
	}
	return &Factory{
		Settings:  settings,
		Workspace: ws,
		GitRunner: git.NewRunner(),
		IO:        iostreams.System(),
	}, nil
}
