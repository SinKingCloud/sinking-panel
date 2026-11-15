package czdb

import (
	"embed"
)

var (
	//go:embed data/*.gz
	Ip embed.FS
)
