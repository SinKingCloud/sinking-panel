package log

// SelectLog 日志查询条件
type SelectLog struct {
	Keyword         string
	Type            string
	Ip              string
	Location        string
	Title           string
	Content         string
	CreateTimeStart string
	CreateTimeEnd   string
	UpdateTimeStart string
	UpdateTimeEnd   string
}
