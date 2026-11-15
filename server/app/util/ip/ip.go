package ip

import (
	"compress/gzip"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"server/app/constant"
	"server/app/util/ip/czdb"
)

func Query(ip string) (*Info, error) {
	parse := net.ParseIP(ip)
	if parse == nil {
		return nil, errors.New("ip格式不正确")
	}
	family, name := 0, "ipv4"
	if parse.To4() == nil {
		family, name = 1, "ipv6"
	}
	// 按地址族首次加载索引，查询时只读取命中的数据块。
	initialize[family].Do(func() {
		if path, err := initIPFile(name); err == nil && path != "" {
			searchers[family], _ = czdb.NewDbSearcher(path, "BTREE", key)
		}
	})
	if searchers[family] == nil {
		return nil, errors.New(name + "数据库初始化失败")
	}
	txt, err := searchers[family].Search(ip)
	if err != nil {
		return nil, err
	}
	if txt == "" {
		return nil, errors.New("未查询到数据")
	}
	index := strings.LastIndexByte(txt, byte(9))
	isp := ""
	if index > 0 {
		isp = txt[index+1:]
	} else {
		index = len(txt)
	}
	arr := strings.Split(txt[:index], "–")
	num := len(arr)
	info := &Info{
		IP: ip,
	}
	if isp != "" {
		info.Isp = isp
	}
	if num >= 1 {
		info.Country = arr[0]
	}
	if num >= 2 {
		info.Province = arr[1]
	}
	if num >= 3 {
		info.City = arr[2]
	}
	if num >= 4 {
		info.Districts = arr[3]
	}
	return info, nil
}

var (
	initialize [2]sync.Once
	searchers  [2]*czdb.DbSearcher // IPv4、IPv6 搜索器
)

const key = "SFdbti7sFbLvmjxB/W179A==" //纯真IP库key

// initIPFile 初始化ip文件
func initIPFile(name string) (string, error) {
	path := filepath.Join(constant.IPPath, name)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(constant.IPPath, 0755); err != nil {
		return "", err
	}
	source, err := czdb.Ip.Open("data/" + name + ".gz")
	if err != nil {
		return "", err
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	target, err := os.CreateTemp(constant.IPPath, "."+name+"-*")
	if err != nil {
		return "", err
	}
	defer target.Close()
	defer os.Remove(target.Name())
	if _, err = io.Copy(target, reader); err != nil {
		return "", err
	}
	if err = target.Chmod(0644); err != nil {
		return "", err
	}
	if err = target.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(target.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}
