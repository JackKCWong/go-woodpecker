package woodpecker

import (
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/spi"
)

type Woodpecker struct {
	GitClient spi.GitClient
	GitServer spi.GitServer
	DepMgr    api.DependencyManager
}

type Opts struct {
	BranchNamePrefix string
}
