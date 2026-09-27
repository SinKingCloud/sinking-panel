package container

import (
	"cmp"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"server/app/util/page"
)

// selectPage 按回调筛选并分页，list 传入切片指针，成功后只保留当前页，返回筛选后的总数。
// match 返回 true 时保留该项，nil 表示不过滤；fields 按当前索引返回 ID、名称、创建时间。
// queryPage 为 nil 时取第一页，每页默认 20 条；Page <= 0 时使用游标分页。
func (m *Manager) selectPage(queryPage *page.Query, list interface{}, match func(int) bool, fields func(int) (string, string, int64)) (int64, error) {
	data := reflect.ValueOf(list)
	if data.Kind() != reflect.Ptr || data.IsNil() || data.Elem().Kind() != reflect.Slice || fields == nil {
		return 0, errors.New("分页列表或排序回调无效")
	}
	data = data.Elem()
	if queryPage == nil {
		queryPage = &page.Query{Page: 1}
	}
	if queryPage.PageSize <= 0 || queryPage.PageSize > 1000 {
		queryPage.PageSize = 20
	}
	if queryPage.OrderByField == "" {
		queryPage.OrderByField = "create_time"
	}
	queryPage.HasMore, queryPage.NextCursorId, queryPage.NextCursorLastId = false, "", ""
	switch queryPage.OrderByField {
	case "id", "name", "create_time", "created_at":
	default:
		return 0, errors.New("分页排序字段无效")
	}
	desc := strings.ToLower(queryPage.OrderByType) != "asc"
	compare := func(index int, id, name string, createdAt int64) int {
		itemID, itemName, itemCreatedAt := fields(index)
		var result int
		switch queryPage.OrderByField {
		case "name":
			result = cmp.Compare(itemName, name)
		case "create_time", "created_at":
			result = cmp.Compare(itemCreatedAt, createdAt)
		}
		if result == 0 {
			result = cmp.Compare(itemID, id)
		}
		if desc {
			return -result
		}
		return result
	}

	cursorMode := queryPage.IsCursor()
	cursorSet := queryPage.CursorId != "" || queryPage.CursorLastId != ""
	cursorID := queryPage.CursorId
	var cursorCreatedAt int64
	if cursorMode && cursorSet {
		if queryPage.OrderByField == "id" {
			if cursorID == "" || queryPage.CursorLastId != "" {
				return 0, errors.New("分页游标无效")
			}
		} else {
			cursorID = queryPage.CursorLastId
			if cursorID == "" {
				return 0, errors.New("分页定位值无效")
			}
			// 名称允许为空，由 CursorLastId 区分空名称游标和首次查询。
			if queryPage.OrderByField == "create_time" || queryPage.OrderByField == "created_at" {
				var err error
				cursorCreatedAt, err = strconv.ParseInt(queryPage.CursorId, 10, 64)
				if err != nil {
					return 0, errors.New("分页游标无效")
				}
			}
		}
	}
	// 筛选判断由调用方提供；向前收拢匹配项，不影响尚未处理的后续项。
	if match != nil {
		count := 0
		for i := 0; i < data.Len(); i++ {
			if match(i) {
				if count != i {
					data.Index(count).Set(data.Index(i))
				}
				count++
			}
		}
		data.Slice(count, data.Len()).Clear()
		data.SetLen(count)
	}
	total := data.Len()
	sort.Slice(data.Interface(), func(i, j int) bool {
		id, name, createdAt := fields(j)
		return compare(i, id, name, createdAt) < 0
	})

	start := 0
	if cursorMode {
		if cursorSet {
			start = sort.Search(total, func(i int) bool {
				return compare(i, cursorID, queryPage.CursorId, cursorCreatedAt) > 0
			})
		}
	} else {
		// 先判断是否超出末页，避免大页码乘法溢出；超出时返回空列表。
		start = total
		if total > 0 && queryPage.Page-1 <= (total-1)/queryPage.PageSize {
			start = (queryPage.Page - 1) * queryPage.PageSize
		}
	}
	end := start + min(queryPage.PageSize, total-start)
	if cursorMode && end < total {
		queryPage.HasMore = true
		id, name, createdAt := fields(end - 1)
		switch queryPage.OrderByField {
		case "id":
			queryPage.NextCursorId = id
		case "name":
			queryPage.NextCursorId, queryPage.NextCursorLastId = name, id
		case "create_time", "created_at":
			queryPage.NextCursorId, queryPage.NextCursorLastId = strconv.FormatInt(createdAt, 10), id
		}
	}
	// 复制当前页，释放全量快照的切片引用；空结果保持 []。
	result := reflect.MakeSlice(data.Type(), end-start, end-start)
	reflect.Copy(result, data.Slice(start, end))
	data.Set(result)
	return int64(total), nil
}
