package page

// IsCursor 是否使用游标分页。
func (q *Query) IsCursor() bool {
	return q != nil && q.Page <= 0
}
