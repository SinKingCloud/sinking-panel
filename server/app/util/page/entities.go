package page

// Result 分页查询结果。
type Result[T any] struct {
	Total            *int64  `json:"total,omitempty"`
	Page             *int    `json:"page,omitempty"`
	PageSize         int     `json:"page_size"`
	List             []T     `json:"list"`
	HasMore          *bool   `json:"has_more,omitempty"`
	NextCursorId     *string `json:"next_cursor_id,omitempty"`
	NextCursorLastId *string `json:"next_cursor_last_id,omitempty"`
}

// Query 分页和排序查询参数。
type Query struct {
	Page         int
	PageSize     int
	OrderByField string
	OrderByType  string
	CursorId     string
	CursorLastId string

	HasMore          bool
	NextCursorId     string
	NextCursorLastId string
}
