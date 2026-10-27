package mongo

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseInfo 确认当前数据库存在，并读取真实集合、文档和存储统计。
// 不发送 MongoDB 4.0.5 才加入的 authorizedDatabases，兼容驱动支持的 3.6。
func (m *Manager) DatabaseInfo(ctx context.Context) (result map[string]interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	name := m.database.Name()
	names, err := m.client.ListDatabaseNames(ctx, bson.D{{
		Key:   "name",
		Value: name,
	}}, options.ListDatabases().SetNameOnly(true))
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("MongoDB 数据库不存在或当前账号无权访问")
	}
	var stats bson.M
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "dbStats",
		Value: 1,
	}}).Decode(&stats); err != nil {
		return nil, err
	}
	result = map[string]interface{}{
		"name":   name,
		"exists": true,
	}
	for source, target := range map[string]string{
		"collections": "collections",
		"views":       "views",
		"objects":     "objects",
		"dataSize":    "data_size",
		"storageSize": "storage_size",
		"indexes":     "indexes",
		"indexSize":   "index_size",
		"avgObjSize":  "avg_object_size",
	} {
		if value, ok := stats[source]; ok {
			result[target] = value
		}
	}
	return result, nil
}

// CreateDatabase 通过创建初始集合生成真实数据库，不接受已有数据库。
func (m *Manager) CreateDatabase(ctx context.Context, collectionName string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	name := m.database.Name()
	if name == "" || len(name) >= 64 || strings.ContainsAny(name, "/\\. \"$*<>:|?\x00") {
		return errors.New("MongoDB 数据库名称必须少于 64 字节，且不包含空格或特殊字符")
	}
	if strings.EqualFold(name, "admin") || strings.EqualFold(name, "local") || strings.EqualFold(name, "config") {
		return errors.New("不能创建 MongoDB 系统数据库")
	}
	if collectionName == "" || len(name)+1+len(collectionName) > 255 || strings.ContainsAny(collectionName, "$\x00") || strings.HasPrefix(collectionName, "system.") || strings.Contains(collectionName, ".system.") {
		return errors.New("请填写有效的初始集合名称，不能使用系统集合")
	}
	// MongoDB 不允许仅用大小写区分数据库；在创建前检查同名数据库。
	filter := bson.D{{
		Key: "name",
		Value: bson.D{{
			Key:   "$regex",
			Value: "^" + regexp.QuoteMeta(name) + "$",
		}, {
			Key:   "$options",
			Value: "i",
		}},
	}}
	names, err := m.client.ListDatabaseNames(ctx, filter, options.ListDatabases().SetNameOnly(true))
	if err != nil {
		return err
	}
	if len(names) > 0 {
		return errors.New("MongoDB 数据库已存在，请同步数据库列表")
	}
	return m.database.CreateCollection(ctx, collectionName)
}

// DropDatabase 删除当前真实数据库，不存在时允许重复删除。
func (m *Manager) DropDatabase(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	name := m.database.Name()
	if name == "" || strings.EqualFold(name, "admin") || strings.EqualFold(name, "local") || strings.EqualFold(name, "config") {
		return errors.New("不能删除 MongoDB 系统数据库")
	}
	return m.database.Drop(ctx)
}

// DropDatabases 先校验整批目标，再复用当前客户端逐个删除业务数据库。
func (m *Manager) DropDatabases(ctx context.Context, names []string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if len(names) == 0 {
		return errors.New("请选择数据库")
	}
	for _, name := range names {
		if strings.TrimSpace(name) == "" || len(name) >= 64 || strings.ContainsAny(name, "/\\. \"$*<>:|?\x00") {
			return errors.New("数据库名称不合法")
		}
		if strings.EqualFold(name, "admin") || strings.EqualFold(name, "local") || strings.EqualFold(name, "config") {
			return errors.New("不能删除系统数据库")
		}
	}
	for completed, name := range names {
		if err = m.client.Database(name).Drop(ctx); err != nil {
			return fmt.Errorf("已删除 %d 个数据库，删除 %s 失败：%w", completed, name, err)
		}
	}
	return nil
}
