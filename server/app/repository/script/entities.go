package script

// SelectScript 常用脚本查询条件
type SelectScript struct {
	TypeId  string
	Keyword string
	Name    string
	Script  string
}

// UpdateScript 常用脚本更新
type UpdateScript struct {
	TypeId interface{}
	Name   interface{}
	Script interface{}
}
