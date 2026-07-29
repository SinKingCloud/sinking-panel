package recycle

import (
	rand2 "math/rand"
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
	name = strings.ReplaceAll(name, "/", "_sk_") + "_t_" + strconv.FormatInt(deleteTime, 10) + "." + strconv.Itoa(id)
	return name
}

// decodeName 还原名称
// name 文件名称
func (s *service) decodeName(name string) (path string, deleteTime int64, id int) {
	arr := strings.Split(name, "_t_")
	if len(arr) != 2 {
		return "", 0, 0
	}
	arr2 := strings.Split(arr[1], ".")
	if len(arr2) == 2 {
		num, _ := strconv.Atoi(arr2[0])
		if num > 0 {
			deleteTime = int64(num)
		}
		num, _ = strconv.Atoi(arr2[1])
		if num > 0 {
			id = num
		}
	}
	path = strings.ReplaceAll(arr[0], "_sk_", "/")
	return path, deleteTime, id
}
