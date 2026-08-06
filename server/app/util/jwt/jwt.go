package jwt

import (
	"crypto/sha256"
	"os"
	"os/exec"
	"runtime"
	"server/app/constant"
	"server/app/util/str"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtKey     []byte
	jwtKeyOnce sync.Once
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

// getKey 获取加密key
func getKey() []byte {
	jwtKeyOnce.Do(func() {
		machineId := ""
		switch runtime.GOOS {
		case "linux":
			for _, path := range []string{"/sys/class/dmi/id/product_uuid", "/etc/machine-id", "/var/lib/dbus/machine-id"} {
				value, err := os.ReadFile(path)
				if err == nil && strings.TrimSpace(string(value)) != "" {
					machineId = string(value)
					break
				}
			}
		case "darwin":
			value, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
			if err == nil {
				for _, line := range strings.Split(string(value), "\n") {
					if strings.Contains(line, "IOPlatformUUID") {
						fields := strings.SplitN(line, "=", 2)
						if len(fields) == 2 {
							machineId = strings.Trim(strings.TrimSpace(fields[1]), "\"")
						}
						break
					}
				}
			}
		case "windows":
			value, err := exec.Command("reg.exe", "query", `HKLM\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid").Output()
			if err == nil {
				for _, line := range strings.Split(string(value), "\n") {
					fields := strings.Fields(line)
					if len(fields) >= 3 && strings.EqualFold(fields[0], "MachineGuid") {
						machineId = fields[len(fields)-1]
						break
					}
				}
			}
		}
		if machineId == "" {
			machineId, _ = os.Hostname()
		}
		value := sha256.Sum256([]byte(constant.JwtKey + ":" + strings.TrimSpace(strings.ToLower(machineId))))
		jwtKey = value[:]
	})
	return jwtKey
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
		setClaim.RegisteredClaims = jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Second)), //过期时间（时间戳）
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
