package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"server/app/util/page"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// SelectPage 查询页码或游标分页数据。
func (r *Repository[T]) SelectPage(query *gorm.DB, queryPage *page.Query, customCursorLastField ...string) (*page.Result[T], error) {
	if r == nil || r.Database == nil || r.Database.Db == nil || query == nil || queryPage == nil || queryPage.PageSize <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if query.Statement == nil || query.Statement.Model == nil {
		return nil, errors.New("分页查询模型无效")
	}
	queryPage.HasMore = false
	queryPage.NextCursorId = ""
	queryPage.NextCursorLastId = ""
	stmt := &gorm.Statement{DB: r.Database.Db}
	if err := stmt.Parse(query.Statement.Model); err != nil {
		return nil, err
	}
	orderField := stmt.Schema.LookUpField(queryPage.OrderByField)
	cursorLastField := stmt.Schema.PrioritizedPrimaryField
	if len(customCursorLastField) > 0 && customCursorLastField[0] != "" {
		cursorLastField = stmt.Schema.LookUpField(customCursorLastField[0])
	}
	if orderField == nil {
		return nil, errors.New("分页排序字段无效")
	}
	if cursorLastField == nil {
		return nil, errors.New("分页定位字段无效")
	}
	field := orderField.DBName
	cursorLastFieldName := cursorLastField.DBName
	orderType := strings.ToLower(queryPage.OrderByType)
	if orderType != "asc" {
		orderType = "desc"
	}
	desc := orderType == "desc"
	limit := queryPage.PageSize
	cursorMode := queryPage.IsCursor()
	cursorSet := queryPage.CursorId != "" || queryPage.CursorLastId != ""
	if cursorMode && cursorSet {
		if (field == cursorLastFieldName && (queryPage.CursorId == "" || queryPage.CursorLastId != "")) || (field != cursorLastFieldName && (queryPage.CursorId == "" || queryPage.CursorLastId == "")) {
			return nil, errors.New("分页游标无效")
		}
		modelValue := reflect.New(stmt.Schema.ModelType).Elem()
		ctx := context.Background()
		if field == cursorLastFieldName {
			if err := orderField.Set(ctx, modelValue, queryPage.CursorId); err != nil {
				return nil, errors.Join(errors.New("分页游标无效"), err)
			}
			cursorId, _ := orderField.ValueOf(ctx, modelValue)
			if desc {
				query = query.Where(clause.Lt{Column: clause.Column{Name: field}, Value: cursorId})
			} else {
				query = query.Where(clause.Gt{Column: clause.Column{Name: field}, Value: cursorId})
			}
		} else {
			if err := cursorLastField.Set(ctx, modelValue, queryPage.CursorLastId); err != nil {
				return nil, errors.Join(errors.New("分页定位值无效"), err)
			}
			cursorLastId, _ := cursorLastField.ValueOf(ctx, modelValue)
			if err := orderField.Set(ctx, modelValue, queryPage.CursorId); err != nil {
				return nil, errors.Join(errors.New("分页游标无效"), err)
			}
			cursorId, _ := orderField.ValueOf(ctx, modelValue)
			if desc {
				query = query.Where(clause.Or(
					clause.Lt{Column: clause.Column{Name: field}, Value: cursorId},
					clause.And(
						clause.Eq{Column: clause.Column{Name: field}, Value: cursorId},
						clause.Lt{Column: clause.Column{Name: cursorLastFieldName}, Value: cursorLastId},
					),
				))
			} else {
				query = query.Where(clause.Or(
					clause.Gt{Column: clause.Column{Name: field}, Value: cursorId},
					clause.And(
						clause.Eq{Column: clause.Column{Name: field}, Value: cursorId},
						clause.Gt{Column: clause.Column{Name: cursorLastFieldName}, Value: cursorLastId},
					),
				))
			}
		}
	}
	if cursorMode {
		limit++
	}
	var total int64
	if !cursorMode {
		const maxPageOffset int64 = 500000
		pageIndex := int64(queryPage.Page - 1)
		if pageIndex > maxPageOffset/int64(queryPage.PageSize) {
			return nil, errors.New("分页位置过深，请使用游标分页")
		}
		if err := query.Count(&total).Error; err != nil {
			return nil, err
		}
		query = query.Offset(int(pageIndex * int64(queryPage.PageSize)))
	}
	if field != cursorLastFieldName {
		query = query.Clauses(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: field}, Desc: desc},
			{Column: clause.Column{Name: cursorLastFieldName}, Desc: desc},
		}})
	} else {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: field}, Desc: desc})
	}
	list := make([]T, 0)
	if err := query.Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	if !cursorMode || len(list) <= queryPage.PageSize {
		return page.New(queryPage, total, list), nil
	}
	queryPage.HasMore = true
	list = list[:queryPage.PageSize]
	last := reflect.ValueOf(list[len(list)-1])
	if !last.IsValid() {
		return nil, errors.New("分页数据无效")
	}
	for last.Kind() == reflect.Interface || last.Kind() == reflect.Ptr {
		if last.IsNil() {
			return nil, errors.New("分页数据无效")
		}
		last = last.Elem()
	}
	resultStmt := &gorm.Statement{DB: r.Database.Db}
	if err := resultStmt.Parse(last.Interface()); err != nil {
		return nil, err
	}
	resultCursorLastField := resultStmt.Schema.LookUpField(cursorLastFieldName)
	resultOrderField := resultStmt.Schema.LookUpField(field)
	if resultCursorLastField == nil || resultOrderField == nil {
		return nil, errors.New("分页结果缺少排序字段")
	}
	values := []struct {
		field  *schema.Field
		target *string
	}{{field: resultOrderField, target: &queryPage.NextCursorId}}
	if field != cursorLastFieldName {
		values = append(values, struct {
			field  *schema.Field
			target *string
		}{field: resultCursorLastField, target: &queryPage.NextCursorLastId})
	}
	for _, cursor := range values {
		fieldValue, _ := cursor.field.ValueOf(context.Background(), last)
		for fieldValue != nil {
			if valuer, ok := fieldValue.(driver.Valuer); ok {
				var err error
				fieldValue, err = valuer.Value()
				if err != nil {
					return nil, err
				}
				break
			}
			value := reflect.ValueOf(fieldValue)
			if value.Kind() != reflect.Interface && value.Kind() != reflect.Ptr {
				break
			}
			if value.IsNil() {
				fieldValue = nil
				break
			}
			fieldValue = value.Elem().Interface()
		}
		if fieldValue == nil {
			return nil, errors.New("分页结果游标无效")
		}
		switch value := fieldValue.(type) {
		case time.Time:
			*cursor.target = value.Format(time.RFC3339Nano)
		case []byte:
			*cursor.target = string(value)
		default:
			*cursor.target = fmt.Sprint(value)
		}
	}
	return page.New(queryPage, total, list), nil
}
