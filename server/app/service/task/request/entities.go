package request

// Config HTTP 请求配置。
type Config struct {
	Method  string            `json:"method"`  // 请求方式
	Url     string            `json:"url"`     // 请求地址
	Headers map[string]string `json:"headers"` // 请求头
	Body    string            `json:"body"`    // 请求体
}
