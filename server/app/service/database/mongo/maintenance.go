package mongo

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	client "go.mongodb.org/mongo-driver/mongo"
)

// Validate 校验集合，禁止操作系统集合。
func (m *Manager) Validate(ctx context.Context, collectionName string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(collection.Name(), "system.") {
		return nil, errors.New("不能维护或删除 MongoDB 系统集合")
	}
	var document bson.Raw
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "validate",
		Value: collection.Name(),
	}, {
		Key:   "maxTimeMS",
		Value: 25000,
	}}).Decode(&document); err != nil {
		return nil, err
	}
	value, length, err := m.decode(document)
	if err != nil {
		return nil, err
	}
	if length > 8<<20 {
		return map[string]interface{}{
			"message":   "维护命令已执行，结果超过 8 MB",
			"truncated": true,
		}, nil
	}
	return map[string]interface{}{
		"message":   "集合校验已完成",
		"result":    value,
		"truncated": false,
	}, nil
}

// Compact 压缩集合，禁止操作系统集合。
func (m *Manager) Compact(ctx context.Context, collectionName string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(collection.Name(), "system.") {
		return nil, errors.New("不能维护或删除 MongoDB 系统集合")
	}
	var document bson.Raw
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "compact",
		Value: collection.Name(),
	}, {
		Key:   "maxTimeMS",
		Value: 25000,
	}}).Decode(&document); err != nil {
		return nil, err
	}
	value, length, err := m.decode(document)
	if err != nil {
		return nil, err
	}
	if length > 8<<20 {
		return map[string]interface{}{
			"message":   "维护命令已执行，结果超过 8 MB",
			"truncated": true,
		}, nil
	}
	return map[string]interface{}{
		"message":   "集合压缩已完成",
		"result":    value,
		"truncated": false,
	}, nil
}

// DropCollection 删除集合，保留重复删除的幂等行为并保护系统集合。
func (m *Manager) DropCollection(ctx context.Context, collectionName string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(collection.Name(), "system.") {
		return nil, errors.New("不能维护或删除 MongoDB 系统集合")
	}
	var document bson.Raw
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "drop",
		Value: collection.Name(),
	}, {
		Key:   "maxTimeMS",
		Value: 25000,
	}}).Decode(&document); err != nil {
		var commandError client.CommandError
		if errors.As(err, &commandError) && commandError.Code == 26 {
			return map[string]interface{}{
				"message":   "集合已删除",
				"result":    map[string]interface{}{"already_absent": true},
				"truncated": false,
			}, nil
		}
		return nil, err
	}
	value, length, err := m.decode(document)
	if err != nil {
		return nil, err
	}
	if length > 8<<20 {
		return map[string]interface{}{
			"message":   "维护命令已执行，结果超过 8 MB",
			"truncated": true,
		}, nil
	}
	return map[string]interface{}{
		"message":   "集合已删除",
		"result":    value,
		"truncated": false,
	}, nil
}
