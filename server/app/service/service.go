package service

import (
	"server/app/service/aliyun"
	"server/app/service/auth"
	"server/app/service/config"
	"server/app/service/notice"
	"server/app/service/pay_log"
	"server/app/service/uin"
	"server/app/service/uin_cookie"
	"server/app/service/uin_log"
	"server/app/service/uin_notify"
	"server/app/service/uin_order"
	"server/app/service/uin_server"
	"server/app/service/user"
	"server/app/service/user_log"
	"server/app/service/user_order"
)

var (
	Config    = config.GetIns()
	Notice    = notice.GetIns()
	PayLog    = pay_log.GetIns()
	UserLog   = user_log.GetIns()
	User      = user.GetIns()
	Auth      = auth.GetIns()
	AliYun    = aliyun.GetIns()
	Uin       = uin.GetIns()
	UinLog    = uin_log.GetIns()
	UinServer = uin_server.GetIns()
	UinCookie = uin_cookie.GetIns()
	UinOrder  = uin_order.GetIns()
	UserOrder = user_order.GetIns()
	UinNotify = uin_notify.GetIns()
)
