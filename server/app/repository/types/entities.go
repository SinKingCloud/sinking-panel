package types

// SelectType 类型查询条件
type SelectType struct {
	Module string
	Name   string
}

// UpdateType 类型更新
type UpdateType struct {
	Module interface{}
	Name   interface{}
	Sort   interface{}
}
