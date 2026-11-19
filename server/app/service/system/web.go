package system

import (
	"cmp"
	"server/app/constant"
	"strconv"
	"time"
)

type webLogo struct {
	path, key string
	expires   time.Time
}

// GetWebInfo 返回面板公开的网站与界面设置，Logo 只返回文件预览签名。
func (s *service) GetWebInfo() map[string]interface{} {
	web := s.configService.Group(constant.WebGroup)
	var ui map[string]string
	var key string
	for {
		old := s.logo.Load()
		ui = s.configService.Group(constant.UiGroup)
		path := ui[constant.UiLogo]
		if old == nil && path == "" {
			break
		}
		if old != nil && old.path == path && time.Now().Before(old.expires) {
			key = old.key
			break
		}
		var next *webLogo
		if path != "" && s.fileService != nil {
			expires := time.Now().Add(constant.CacheTimeWithFilePreview)
			if key, err := s.fileService.CreatePreviewSign(path, false, constant.CacheTimeWithFilePreview); err == nil {
				next = &webLogo{path: path, key: key, expires: expires}
			}
		}
		// 只保留成功发布的签名，并发请求不返回各自生成的临时签名。
		if !s.logo.CompareAndSwap(old, next) {
			if next != nil {
				s.cache.Delete(constant.CacheNameWithFilePreview + next.key)
			}
			continue
		}
		if old != nil {
			s.cache.Delete(constant.CacheNameWithFilePreview + old.key)
		}
		if next == nil {
			break
		}
	}
	radius, err := strconv.Atoi(ui[constant.UiRadius])
	if err != nil || radius < 0 {
		radius = 8
	}
	enabled := func(value string) bool { return value == "1" || value == "true" }
	return map[string]interface{}{
		"title": cmp.Or(web[constant.WebTitle], "云上豁者"),
		"name":  cmp.Or(web[constant.WebName], "云上豁者"),
		"ui": map[string]interface{}{
			"logo":      key,
			"layout":    cmp.Or(ui[constant.UiLayout], "left"),
			"watermark": enabled(ui[constant.UiWaterMark]),
			"theme":     cmp.Or(ui[constant.UiTheme], "light"),
			"compact":   enabled(ui[constant.UiCompact]),
			"color":     cmp.Or(ui[constant.UiColor], "rgb(0, 83, 251)"),
			"radius":    radius,
		},
	}
}
