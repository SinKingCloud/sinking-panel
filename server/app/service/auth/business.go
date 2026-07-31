package auth

import (
	"errors"
	"server/app/constant"
	"server/app/util/jwt"
	"server/app/util/str"
	"strconv"
	"time"
)

// Login 账号登录
func (s *service) Login(account string, pwd string, types string, ip string) (string, error) {
	sUser := s.configService.Get(constant.LoginGroup, constant.LoginAccount)
	sPwd := s.configService.Get(constant.LoginGroup, constant.LoginPassword)
	if sUser == "" && sPwd == "" {
		password := str.NewStringTool(pwd).GetPassword()
		if password == "" {
			return "", errors.New("密码加密失败")
		}
		if err := s.configService.Sets(map[string]string{
			constant.LoginAccount:  account,
			constant.LoginPassword: password,
		}); err != nil {
			return "", err
		}
		sUser = account
		sPwd = password
	}
	if (sUser == "") != (sPwd == "") {
		return "", errors.New("登录配置异常")
	}
	if sUser != account || !str.NewStringTool(sPwd).CheckPassword(pwd) {
		return "", errors.New("用户名或密码错误")
	}

	token := str.NewStringTool(strconv.FormatInt(time.Now().UnixMilli(), 10)).Md5()
	loginTime := str.DateTime(time.Now())
	if token != "" {
		if err := s.configService.Set(constant.LoginToken+"."+types, token); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("生成token失败")
	}
	expire, _ := strconv.Atoi(s.configService.Get(constant.LoginGroup, constant.LoginExpire))
	if expire > 0 && expire <= 7200 {
		expire = 7200
	}
	tokenValue := jwt.GetToken(&jwt.User{
		LoginToken: token,
		LoginIp:    ip,
		LoginTime:  loginTime,
	}, expire)
	return tokenValue, nil
}

// Logout 注销登录
func (s *service) Logout(types string) error {
	return s.configService.Set(constant.LoginToken+"."+types, "")
}

// UpdateAccount 修改账户
func (s *service) UpdateAccount(account string, password string) error {
	configs := make(map[string]string)
	if account != "" {
		configs[constant.LoginAccount] = account
	}
	if password != "" {
		value := str.NewStringTool(password).GetPassword()
		if value == "" {
			return errors.New("密码加密失败")
		}
		configs[constant.LoginPassword] = value
	}
	if len(configs) == 0 {
		return errors.New("账户信息不能为空")
	}
	if err := s.configService.Sets(configs); err != nil {
		return errors.New("修改账户信息失败")
	}
	return nil
}
