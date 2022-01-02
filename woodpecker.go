package woodpecker

import (
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/spi"
)

type Woodpecker struct {
	GitClient spi.GitClient
	GitServer spi.GitServer
	DepMgr    api.DependencyManager
	Opts      Opts
}

type Opts struct {
	BranchNamePrefix string
}

func (w Woodpecker) Dig() error {
	return nil
}
