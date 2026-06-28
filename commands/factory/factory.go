package factory

import (
	"github.com/mehranzand/repofleet/internal/config"
	"github.com/mehranzand/repofleet/internal/git"
	"github.com/mehranzand/repofleet/internal/iostreams"
)

type Factory struct {
	Config    *config.Config
	GitRunner *git.Runner
	IO        *iostreams.IOStreams
}

func New() (*Factory, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Factory{
		Config:    cfg,
		GitRunner: git.NewRunner(),
		IO:        iostreams.System(),
	}, nil
}
