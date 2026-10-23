package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"server/app/util/str"
)

// MyClaims jwt载体
type MyClaims struct {
	User *User
	jwt.RegisteredClaims
}

// User 会员用户表
type User struct {
	LoginToken string       ` json:"login_token"`
	LoginIp    string       ` json:"login_ip"`
	LoginTime  str.DateTime ` json:"login_time"`
}
