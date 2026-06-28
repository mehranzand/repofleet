package factory

import (
	"github.com/mehranzand/repofleet/internal/git"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/store"
)

type Factory struct {
	Config    *store.Config
	GitRunner *git.Runner
	IO        *iostreams.IOStreams
}

func New() (*Factory, error) {
	cfg, err := store.Load()
	if err != nil {
		return nil, err
	}
	return &Factory{
		Config:    cfg,
		GitRunner: git.NewRunner(),
		IO:        iostreams.System(),
	}, nil
}
