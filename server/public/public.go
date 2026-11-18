package public

import (
	"embed"
	"net/http"
)

var (
	//go:embed dist/*
	Static embed.FS
	//go:embed sql/install.sql
	Sql string
	//go:embed all:plugin
	Plugin embed.FS

	FileServer = http.FileServer(http.FS(Static))
)

func Path() string {
	return "/dist"
}
