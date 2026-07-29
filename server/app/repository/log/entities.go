package log

// SelectLog 日志查询条件
type SelectLog struct {
	Type            string
	Ip              string
	Title           string
	Content         string
	CreateTimeStart string
	CreateTimeEnd   string
	UpdateTimeStart string
	UpdateTimeEnd   string
}
