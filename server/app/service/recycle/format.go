package recycle

import (
	"encoding/base64"
	rand2 "math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// encodeName 加密名称
// name 文件路径
func (s *service) encodeName(name string, deleteTime int64, id int) string {
	if deleteTime <= 0 {
		deleteTime = time.Now().Unix()
	}
	if id <= 0 {
		id = 1000000 + rand2.Intn(8999999)
	}
	name = "v2_" + base64.RawURLEncoding.EncodeToString([]byte(name)) + "_t_" + strconv.FormatInt(deleteTime, 10) + "." + strconv.Itoa(id)
	return name
}

// decodeName 还原名称
// name 文件名称
func (s *service) decodeName(name string) (path string, deleteTime int64, id int) {
	index := strings.LastIndex(name, "_t_")
	if index <= 0 {
		return "", 0, 0
	}
	encodedPath := name[:index]
	arr2 := strings.Split(name[index+3:], ".")
	if len(arr2) == 2 {
		num, err := strconv.ParseInt(arr2[0], 10, 64)
		if err == nil && num > 0 {
			deleteTime = num
		}
		num, err = strconv.ParseInt(arr2[1], 10, 32)
		if err == nil && num > 0 {
			id = int(num)
		}
	}
	if strings.HasPrefix(encodedPath, "v2_") {
		decodedPath, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encodedPath, "v2_"))
		if err != nil || len(decodedPath) == 0 {
			return "", 0, 0
		}
		path = string(decodedPath)
	} else {
		path = filepath.FromSlash(strings.ReplaceAll(encodedPath, "_sk_", "/"))
	}
	return path, deleteTime, id
}
