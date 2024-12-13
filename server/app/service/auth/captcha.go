package auth

import (
	"server/app/constant"
	"server/app/util"
	"server/app/util/captcha"
	"strings"
)

// GetCaptcha 获取验证码
func (c *Service) GetCaptcha(key string) (*captcha.Image, string) {
	img, code := captcha.NewCaptchaRandCode(4)
	//写入redis
	util.Cache.SetWithExpire(constant.CacheNameWithCaptcha+key, strings.ToLower(code), constant.CacheTimeWithCaptcha)
	return img, code
}

// CheckCaptcha 判断验证码是否正确
func (c *Service) CheckCaptcha(key string, code string) bool {
	value := util.Cache.Get(constant.CacheNameWithCaptcha + key)
	util.Cache.Delete(constant.CacheNameWithCaptcha + key)
	if value != nil {
		if value.(string) == strings.ToLower(code) {
			return true
		}
		return false
	}
	return false
}
