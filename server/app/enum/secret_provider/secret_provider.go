package secret_provider

const (
	TencentCloud = iota // 腾讯云
	Aliyun              // 阿里云
	HuaweiCloud         // 华为云
	Volcengine          // 火山引擎
	BaiduCloud          // 百度云
)

// Map 密钥服务商数据。
func Map() map[int]string {
	return map[int]string{
		TencentCloud: "腾讯云",
		Aliyun:       "阿里云",
		HuaweiCloud:  "华为云",
		Volcengine:   "火山云",
		BaiduCloud:   "百度云",
	}
}
