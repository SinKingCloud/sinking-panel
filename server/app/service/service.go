package service

import (
	"server/app/service/auth"
	"server/app/service/config"
)

var (
	Config = config.GetIns()
	Auth   = auth.GetIns()
)
