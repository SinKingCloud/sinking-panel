package public

import (
	"embed"
	"net/http"
	"path"
	"server/app/util/context"
	"strings"
)

var (
	//go:embed dist/*
	dist embed.FS
	//go:embed sql/install.sql
	Sql string
	//go:embed all:plugin
	Plugin embed.FS

	distFileServer = http.FileServer(http.FS(dist))
)

// Static 静态资源
func Static(c *context.Context) {
	requestPath, _, _ := strings.Cut(c.Request.URL.Path, "#")
	switch strings.ToLower(path.Ext(requestPath)) {
	case ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".svg", ".ico", ".bmp", ".tif", ".tiff":
		c.SetHeader("Cache-Control", "public, max-age=86400")
	case "", ".html", ".htm":
		c.SetHeader("Cache-Control", "no-cache, must-revalidate")
	}
	c.Request.URL.Path = "/dist" + requestPath
	distFileServer.ServeHTTP(c.Writer, c.Request)
}
