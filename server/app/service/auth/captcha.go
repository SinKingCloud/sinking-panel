package auth

import (
	"encoding/json"
	"errors"
	"server/app/constant"
	"server/app/util/captcha"

	"github.com/wenlng/go-captcha/v2/slide"
)

// GetCaptcha 获取验证码
func (s *service) GetCaptcha(key string) (map[string]interface{}, error) {
	result, masterImage, tileImage, height, width, err := captcha.GenSlide()
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, errors.New("获取验证码失败")
	}
	s.cache.SetWithExpire(constant.CacheNameWithCaptcha+key, string(bytes), constant.CacheTimeWithCaptcha)
	ret := map[string]interface{}{
		"key":          key,
		"image_base64": masterImage,
		"width":        width,
		"height":       height,
		"tile_base64":  tileImage,
		"tile_width":   result.Width,
		"tile_height":  result.Height,
		"tile_x":       result.DX,
		"tile_y":       result.DY,
	}
	return ret, nil
}

// CheckCaptcha 判断验证码是否正确
func (s *service) CheckCaptcha(key string, x int, y int) bool {
	value := s.cache.Get(constant.CacheNameWithCaptcha + key)
	s.cache.Delete(constant.CacheNameWithCaptcha + key)
	valueString, ok := value.(string)
	if !ok || valueString == "" {
		return false
	}
	var dct *slide.Block
	if err := json.Unmarshal([]byte(valueString), &dct); err != nil || dct == nil {
		return false
	}
	return captcha.CheckSlide(x, y, dct)
}
