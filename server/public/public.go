package public

import (
	"embed"
	"net/http"
)

var (
	//go:embed dist/*
	Static embed.FS
	//go:embed sql/*
	sql embed.FS
	//go:embed captcha/*
	captcha    embed.FS
	FileServer = http.FileServer(http.FS(Static))
)

func Path() string {
	return "/dist"
}

func Index() []byte {
	bytes, err := Static.ReadFile("dist/index.html")
	if err != nil || bytes == nil {
		return nil
	}
	return bytes
}

func Sql() string {
	bytes, err := sql.ReadFile("sql/install.sql")
	if err != nil || bytes == nil {
		return ""
	}
	return string(bytes)
}

func Captcha() []byte {
	bytes, err := captcha.ReadFile("captcha/tencent.js")
	if err != nil || bytes == nil {
		return nil
	}
	return bytes
}
