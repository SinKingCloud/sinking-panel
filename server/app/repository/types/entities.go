package types

// SelectType 类型查询条件
type SelectType struct {
	Module *string
	Name   *string
}

// UpdateType 类型更新
type UpdateType struct {
	Module  *string
	Name    *string
	Sort    *int64
	Reorder *bool // 按传入ID顺序写入连续排序值。
}
