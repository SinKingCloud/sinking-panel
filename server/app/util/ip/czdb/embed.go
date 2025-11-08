package czdb

import (
	"embed"
)

var (
	//go:embed data/*
	Ip embed.FS
)
