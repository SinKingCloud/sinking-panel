package secret

import "server/app/util/str"

// Secret 返回密钥详情，Data 只包含当前厂商的脱敏字段。
type Secret struct {
	Id         int64             `json:"id"`
	Name       string            `json:"name"`
	Provider   int               `json:"provider"`
	Data       map[string]string `json:"data"`
	CreateTime str.DateTime      `json:"create_time"`
	UpdateTime str.DateTime      `json:"update_time"`
}

// TencentCloud 腾讯云密钥。
type TencentCloud struct {
	SecretId  string `json:"tencent_secret_id"`
	SecretKey string `json:"tencent_secret_key"`
}

// Aliyun 阿里云密钥。
type Aliyun struct {
	AccessKeyId     string `json:"aliyun_access_key_id"`
	AccessKeySecret string `json:"aliyun_access_key_secret"`
}

// HuaweiCloud 华为云密钥。
type HuaweiCloud struct {
	AccessKeyId     string `json:"huawei_access_key_id"`
	SecretAccessKey string `json:"huawei_secret_access_key"`
}

// Volcengine 火山云密钥。
type Volcengine struct {
	AccessKeyId     string `json:"volcengine_access_key_id"`
	AccessKeySecret string `json:"volcengine_access_key_secret"`
}

// BaiduCloud 百度云密钥。
type BaiduCloud struct {
	AccessKeyId     string `json:"baidu_access_key_id"`
	SecretAccessKey string `json:"baidu_secret_access_key"`
}
