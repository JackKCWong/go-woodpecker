package woodpecker

import (
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/spi"
)

type Woodpecker struct {
	GitClient spi.GitClient
	GitServer spi.GitServer
	DepMgr    api.DependencyManager
	OSSIndex  spi.OSSIndex
}

type Opts struct {
	Verbose          bool
	BranchNamePrefix string
}
