package jwt

import (
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"server/app/constant"
	"server/app/util/str"
	"time"
)

// MyClaims jwt载体
type MyClaims struct {
	User *User
	jwt.StandardClaims
}

// User 会员用户表
type User struct {
	Id         int          `gorm:"column:id;PRIMARY_KEY" json:"id"`
	LoginToken string       `gorm:"column:login_token" json:"login_token"`
	LoginIp    string       `gorm:"column:login_ip" json:"login_ip"`
	LoginTime  str.DateTime `gorm:"column:login_time" json:"login_time"`
	Status     int          `gorm:"column:status" json:"status"`
}

// getKey 获取加密key
func getKey() []byte {
	return []byte(constant.JwtKey)
}

// GetLoginToken 获取token
func GetLoginToken(token string) map[string]string {
	temp := make(map[string]string)
	_ = json.Unmarshal([]byte(token), &temp)
	return temp
}

// GetToken 生成token user 用户信息
func GetToken(user *User, expireTime int) string {
	if expireTime == 0 {
		expireTime = constant.JwtExpireTime
	}
	setClaim := MyClaims{
		User: user,
	}
	if expireTime >= 0 {
		setClaim.StandardClaims = jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(expireTime) * time.Second).Unix(), //过期时间（时间戳）
		}
	}
	reqClaim := jwt.NewWithClaims(jwt.SigningMethodHS256, setClaim) //生成token
	token, err := reqClaim.SignedString(getKey())                   //转换为字符串
	if err != nil {
		return ""
	}
	return token
}

// CheckToken 验证token
func CheckToken(token string) *MyClaims {
	setToken, err := jwt.ParseWithClaims(token, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getKey(), nil
	})
	if err != nil {
		return nil
	}
	if key, success := setToken.Claims.(*MyClaims); setToken.Valid && success {
		return key
	} else {
		return nil
	}
}
