package auth

import (
	"errors"
	"server/app/constant"
	"server/app/util/jwt"
	"server/app/util/str"
	"strconv"
	"time"
)

// CheckAccount 判断账号密码
func (s *service) CheckAccount(account string, pwd string) error {
	sUser := s.configService.Get(constant.LoginGroup, constant.LoginAccount)
	sPwd := s.configService.Get(constant.LoginGroup, constant.LoginPassword)
	if sUser == "" || sPwd == "" || sUser != account || !str.NewStringTool(sPwd).CheckPassword(pwd) {
		return errors.New("用户名或密码错误")
	}
	return nil
}

// GenLoginToken 生成jwtToken
func (s *service) GenLoginToken(types string, ip string) (tokenValue string, err error) {
	token := str.NewStringTool(strconv.FormatInt(time.Now().UnixMilli(), 10)).Md5()
	loginTime := str.DateTime(time.Now())
	if token != "" {
		err = s.configService.Set(constant.LoginGroup, constant.LoginToken+"."+types, token)
	} else {
		err = errors.New("生成token失败")
	}
	if err != nil {
		return "", err
	}
	expire, _ := strconv.Atoi(s.configService.Get(constant.LoginGroup, constant.LoginExpire))
	if expire > 0 && expire <= 600 {
		expire = 600
	}
	tokenValue = jwt.GetToken(&jwt.User{
		LoginToken: token,
		LoginIp:    ip,
		LoginTime:  loginTime,
	}, expire)
	return tokenValue, nil
}

// ClearLoginToken 清理jwtToken
func (s *service) ClearLoginToken(types string) error {
	err := s.configService.Set(constant.LoginGroup, constant.LoginToken+"."+types, "")
	if err != nil {
		return err
	}
	return nil
}
