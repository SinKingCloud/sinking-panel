package task

type Type int //日志类型

const (
	Script Type = iota //用户操作
)

// Types 类型数据
func (s *Service) Types() map[Type]string {
	return map[Type]string{
		Script: "系统脚本",
	}
}

type Status int

const (
	Running Status = iota //运行
	Stop                  //暂停
)

// Status 状态数据
func (s *Service) Status() map[Status]string {
	return map[Status]string{
		Running: "运行",
		Stop:    "暂停",
	}
}
