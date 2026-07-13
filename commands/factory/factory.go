package factory

import (
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
	"github.com/mehranzand/repofleet/internal/util/git"
)

type Factory struct {
	Workspace *store.Workspace
	GitRunner *git.Runner
	IO        *iostreams.IOStreams
}

func New() (*Factory, error) {
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	wsName := store.CurrentWorkspaceName()
	ws, err := store.LoadWorkspace(wsName)
	if err != nil {
		return nil, err
	}
	return &Factory{
		Workspace: ws,
		GitRunner: git.NewRunner(),
		IO:        iostreams.System(),
	}, nil
}
