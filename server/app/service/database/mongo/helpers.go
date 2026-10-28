package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	client "go.mongodb.org/mongo-driver/v2/mongo"
)

func (m *Manager) collection(name string) (*client.Collection, error) {
	if name == "" || len(name) > 255 || strings.ContainsRune(name, 0) {
		return nil, errors.New("请选择有效集合")
	}
	return m.database.Collection(name), nil
}
func (m *Manager) parse(text, label string, required bool) (bson.D, error) {
	if strings.TrimSpace(text) == "" {
		if required {
			return nil, fmt.Errorf("请填写%s", label)
		}
		return bson.D{}, nil
	}
	if len(text) > 8<<20 {
		return nil, fmt.Errorf("%s不能超过 8 MB", label)
	}
	var document bson.D
	if bson.UnmarshalExtJSON([]byte(text), false, &document) != nil || document == nil {
		return nil, fmt.Errorf("%s必须是有效的 JSON 对象", label)
	}
	if required && len(document) == 0 {
		return nil, fmt.Errorf("%s不能为空对象", label)
	}
	return document, nil
}

func (m *Manager) decode(document interface{}) (interface{}, int, error) {
	encoded, e := bson.MarshalExtJSON(document, true, false)
	if e != nil {
		return nil, 0, e
	}
	var value interface{}
	if e = json.Unmarshal(encoded, &value); e != nil {
		return nil, 0, e
	}
	return value, len(encoded), nil
}

func (m *Manager) readCursor(ctx context.Context, cursor *client.Cursor, max, page int) (map[string]interface{}, error) {
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = cursor.Close(closeCtx)
	}()
	documents := make([]interface{}, 0, max)
	bytes := 0
	more, truncated := false, false
	for cursor.Next(ctx) {
		if len(documents) >= max {
			more = true
			break
		}
		value, length, e := m.decode(cursor.Current)
		if e != nil {
			return nil, e
		}
		if bytes+length > 8<<20 {
			more, truncated = true, true
			break
		}
		bytes += length
		documents = append(documents, value)
	}
	if e := cursor.Err(); e != nil {
		return nil, e
	}
	return map[string]interface{}{
		"documents": documents,
		"has_more":  more,
		"page":      page,
		"page_size": max,
		"truncated": truncated,
	}, nil
}
