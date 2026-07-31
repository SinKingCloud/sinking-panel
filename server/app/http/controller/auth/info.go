package auth

import (
	"server/app/constant"
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Info 网站信息
func Info(c *context.Context) {
	web := service.Config.Group(constant.WebGroup)
	ui := service.Config.Group(constant.UiGroup)
	c.SuccessWithData("获取成功", sinking_web.H{
		"title": c.GetStringWithDefault(web[constant.WebTitle], "云上豁者"),
		"name":  c.GetStringWithDefault(web[constant.WebName], "云上豁者"),
		"ui": sinking_web.H{
			"layout":    c.GetStringWithDefault(ui[constant.UiLayout], "left"),
			"watermark": c.GetBoolWithDefault(ui[constant.UiWaterMark], false),
			"theme":     c.GetStringWithDefault(ui[constant.UiTheme], "dark"),
			"compact":   c.GetBoolWithDefault(ui[constant.UiCompact], false),
			"color":     c.GetStringWithDefault(ui[constant.UiColor], "rgb(0,81,235)"),
			"radius":    c.GetIntWithDefault(ui[constant.UiRadius], 0),
		},
	})
}
