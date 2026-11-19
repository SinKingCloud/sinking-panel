package system

import (
	"cmp"
	"server/app/constant"
	"strconv"
	"time"
)

// GetWebInfo 返回面板公开的网站与界面设置，Logo 只返回文件预览签名。
func (s *service) GetWebInfo(sign func(string, bool, time.Duration) (string, error)) map[string]interface{} {
	web := s.configService.Group(constant.WebGroup)
	ui := s.configService.Group(constant.UiGroup)
	type webLogo struct {
		path, key string
		expires   time.Time
	}
	valid := func(path string, logo webLogo) bool {
		return path == logo.path && (path == "" || time.Now().Before(logo.expires))
	}
	CacheNameWithWebLogo := "WebLogo" //面板 Logo 的签名记录
	s.logoMu.RLock()
	logo, _ := s.cache.Get(CacheNameWithWebLogo).(webLogo)
	s.logoMu.RUnlock()
	if !valid(ui[constant.UiLogo], logo) {
		s.logoMu.Lock()
		defer s.logoMu.Unlock()
		ui = s.configService.Group(constant.UiGroup)
		logo, _ = s.cache.Get(CacheNameWithWebLogo).(webLogo)
		if path := ui[constant.UiLogo]; !valid(path, logo) {
			if logo.key != "" {
				s.cache.Delete(constant.CacheNameWithFilePreview + logo.key)
				s.cache.Delete(CacheNameWithWebLogo)
			}
			logo = webLogo{}
			if path != "" && sign != nil {
				expires := time.Now().Add(constant.CacheTimeWithFilePreview)
				if key, err := sign(path, false, constant.CacheTimeWithFilePreview); err == nil && key != "" {
					logo = webLogo{path: path, key: key, expires: expires}
					s.cache.SetWithExpire(CacheNameWithWebLogo, logo, -1)
				}
			}
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
			"logo":      logo.key,
			"layout":    cmp.Or(ui[constant.UiLayout], "left"),
			"watermark": enabled(ui[constant.UiWaterMark]),
			"theme":     cmp.Or(ui[constant.UiTheme], "light"),
			"compact":   enabled(ui[constant.UiCompact]),
			"color":     cmp.Or(ui[constant.UiColor], "rgb(0, 83, 251)"),
			"radius":    radius,
		},
	}
}
