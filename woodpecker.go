package woodpecker

import (
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/spi"
)

type Woodpecker struct {
	Client spi.GitClient
	Server spi.GitServer
	DepMgr api.DependencyManager
}
