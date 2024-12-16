package ip

import (
	"errors"
	"net"
	"server/app/constant"
	file2 "server/app/util/file"
	"server/app/util/ip/czdb"
	"server/public"
	"strings"
)

type Info struct {
	IP        string `json:"ip"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Districts string `json:"districts"`
	Isp       string `json:"isp"`
}

func Query(ip string) (*Info, error) {
	parse := net.ParseIP(ip)
	if parse == nil {
		return nil, errors.New("ip格式不正确")
	}
	var txt string
	var err error
	if parse.To4() != nil {
		if v4Search == nil {
			return nil, errors.New("ipv4数据库初始化失败")
		}
		txt, err = v4Search.Search(ip)
	} else {
		if v6Search == nil {
			return nil, errors.New("ipv6数据库初始化失败")
		}
		txt, err = v6Search.Search(ip)
	}
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
	v4Search *czdb.DbSearcher //ipv4查询
	v6Search *czdb.DbSearcher //ipv6查询
)

const key = "SFdbti7sFbLvmjxB/W179A==" //纯真IP库key

func init() {
	v4, err := initIPFile("cz88_public_v4.czdb")
	if err == nil && v4 != "" {
		v4Search, err = czdb.NewDbSearcher(v4, "MEMORY", key)
	}
	v6, err := initIPFile("cz88_public_v6.czdb")
	if err == nil && v6 != "" {
		v6Search, err = czdb.NewDbSearcher(v6, "MEMORY", key)
	}
}

// initIPFile 初始化ip文件
func initIPFile(name string) (string, error) {
	path := constant.DBPath
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	file := file2.New(path)
	_ = file.EnsureDirectory(path, 0777)
	if !file.Has(name) {
		open, err := public.Ip.Open("ip/" + name)
		if err == nil {
			defer func() {
				_ = open.Close()
			}()
			_, err = file.WriteStream(name, open, 0777)
		} else {
			_, err = file.Write(name, "", 0777)
		}
		if err != nil {
			return "", err
		}
	}
	return strings.ReplaceAll(path+name, "//", ""), nil
}
