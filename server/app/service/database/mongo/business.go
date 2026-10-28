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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Collections 列举当前数据库的集合。
func (m *Manager) Collections(ctx context.Context) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	// 内部集合不作为业务数据浏览，避免自动查询 system.sessions、system.buckets 等特殊集合。
	// system.js 和 system.profile 支持普通文档查询，保留函数与性能分析数据入口。
	filter := bson.D{{
		Key: "$or",
		Value: bson.A{
			bson.D{{
				Key: "name",
				Value: bson.D{{
					Key: "$not",
					Value: bson.D{{
						Key:   "$regex",
						Value: "^system\\.",
					}},
				}},
			}},
			bson.D{{
				Key: "name",
				Value: bson.D{{
					Key:   "$in",
					Value: bson.A{"system.js", "system.profile"},
				}},
			}},
		},
	}}
	settings := options.ListCollections().SetNameOnly(true).SetAuthorizedCollections(true).SetBatchSize(100)
	cursor, e := m.database.ListCollections(ctx, filter, settings)
	if e != nil {
		var commandError client.CommandError
		// nameOnly/authorizedCollections 从 4.0 引入；旧版本只移除不支持的参数。
		// 不把权限、认证或连接错误当作版本差异重试。
		if errors.As(e, &commandError) && (commandError.Code == 2 || commandError.Code == 9 || commandError.Code == 40415) {
			text := strings.ToLower(commandError.Message)
			if strings.Contains(text, "nameonly") || strings.Contains(text, "authorizedcollections") {
				cursor, e = m.database.ListCollections(ctx, filter, options.ListCollections().SetBatchSize(100))
			}
		}
	}
	if e != nil {
		return nil, e
	}
	defer cursor.Close(ctx)
	rows := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		var item struct {
			Name string `bson:"name"`
			Type string `bson:"type"`
		}
		if e = cursor.Decode(&item); e != nil {
			return nil, e
		}
		if len(rows) >= 10000 {
			return nil, errors.New("集合数量超过 10000，请使用命令查询")
		}
		rows = append(rows, map[string]interface{}{
			"name": item.Name,
			"type": item.Type,
		})
	}
	return rows, cursor.Err()
}

// Indexes 读取集合索引。
func (m *Manager) Indexes(ctx context.Context, collectionName string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	cursor, e := collection.Indexes().List(ctx)
	if e != nil {
		return nil, e
	}
	defer cursor.Close(ctx)
	rows := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		var item struct {
			Name               string `bson:"name"`
			Key                bson.D `bson:"key"`
			Unique             bool   `bson:"unique"`
			Sparse             bool   `bson:"sparse"`
			ExpireAfterSeconds *int64 `bson:"expireAfterSeconds"`
		}
		if e = cursor.Decode(&item); e != nil {
			return nil, e
		}
		keys, _, e := m.decode(item.Key)
		if e != nil {
			return nil, e
		}
		row := map[string]interface{}{
			"name":   item.Name,
			"keys":   keys,
			"unique": item.Unique,
			"sparse": item.Sparse,
		}
		if item.ExpireAfterSeconds != nil {
			row["expire_after_seconds"] = *item.ExpireAfterSeconds
		}
		rows = append(rows, row)
	}
	return rows, cursor.Err()
}

// Documents 分页查询文档。
func (m *Manager) Documents(ctx context.Context, collectionName, filterJSON, sortJSON, projectionJSON string, page, pageSize int) (result interface{}, err error) {
	if page < 0 || page > 1000000 || pageSize < 0 {
		return nil, errors.New("分页和行数参数不合法")
	}
	if len(filterJSON) > 1<<20 || len(sortJSON) > 65536 || len(projectionJSON) > 65536 {
		return nil, errors.New("操作参数超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	size := pageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	if size > 1000 {
		size = 1000
	}
	filter, e := m.parse(filterJSON, "筛选条件", false)
	if e != nil {
		return nil, e
	}
	sort, e := m.parse(sortJSON, "排序条件", false)
	if e != nil {
		return nil, e
	}
	projection, e := m.parse(projectionJSON, "显示字段", false)
	if e != nil {
		return nil, e
	}
	started := time.Now()
	countCtx, cancelCount := context.WithTimeout(ctx, 25*time.Second)
	total, e := collection.CountDocuments(countCtx, filter)
	cancelCount()
	if e != nil {
		return nil, e
	}
	if maxPage := max(1, int((total+int64(size)-1)/int64(size))); page > maxPage {
		page = maxPage
	}
	if len(sort) == 0 {
		sort = bson.D{{
			Key:   "_id",
			Value: 1,
		}}
	}
	opts := options.Find().SetSkip(int64(page-1) * int64(size)).SetLimit(int64(size + 1)).SetBatchSize(int32(min(size, 50))).SetSort(sort)
	if len(projection) > 0 {
		opts.SetProjection(projection)
	}
	queryCtx, cancelQuery := context.WithTimeout(ctx, 25*time.Second)
	defer cancelQuery()
	cursor, e := collection.Find(queryCtx, filter, opts)
	if e != nil {
		return nil, e
	}
	data, e := m.readCursor(queryCtx, cursor, size, page)
	if e != nil {
		return nil, e
	}
	data["total"] = total
	data["elapsed_ms"] = time.Since(started).Milliseconds()
	return data, nil
}

// Insert 插入文档。
func (m *Manager) Insert(ctx context.Context, collectionName, documentJSON string) (result interface{}, err error) {
	if len(documentJSON) > 8<<20 {
		return nil, errors.New("文档不能超过 8 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(documentJSON)
	if text == "" || len(text) > 8<<20 {
		return nil, errors.New("请填写不超过 8 MB 的文档")
	}
	if strings.HasPrefix(text, "[") {
		var raw []json.RawMessage
		if json.Unmarshal([]byte(text), &raw) != nil || len(raw) == 0 || len(raw) > 1000 {
			return nil, errors.New("批量插入需要 1 至 1000 个文档")
		}
		documents := make([]interface{}, len(raw))
		for i, item := range raw {
			document, e := m.parse(string(item), "文档", false)
			if e != nil {
				return nil, e
			}
			documents[i] = document
		}
		inserted, e := collection.InsertMany(ctx, documents)
		if e != nil {
			return nil, e
		}
		ids, _, e := m.decode(bson.D{{
			Key:   "ids",
			Value: inserted.InsertedIDs,
		}})
		return map[string]interface{}{
			"affected": len(inserted.InsertedIDs),
			"inserted": ids,
		}, e
	}
	document, e := m.parse(text, "文档", false)
	if e != nil {
		return nil, e
	}
	inserted, e := collection.InsertOne(ctx, document)
	if e != nil {
		return nil, e
	}
	id, _, e := m.decode(bson.D{{
		Key:   "id",
		Value: inserted.InsertedID,
	}})
	return map[string]interface{}{
		"affected": 1,
		"inserted": id,
	}, e
}

// Update 按非空条件修改文档。
func (m *Manager) Update(ctx context.Context, collectionName, filterJSON, updateJSON string, many, upsert bool) (result interface{}, err error) {
	if len(filterJSON) > 1<<20 || len(updateJSON) > 8<<20 {
		return nil, errors.New("操作参数超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	filter, e := m.parse(filterJSON, "筛选条件", true)
	if e != nil {
		return nil, e
	}
	update, e := m.parse(updateJSON, "更新内容", true)
	if e != nil {
		return nil, e
	}
	var updated *client.UpdateResult
	if !strings.HasPrefix(update[0].Key, "$") {
		if many {
			return nil, errors.New("替换文档只支持单条，请关闭批量更新")
		}
		updated, e = collection.ReplaceOne(ctx, filter, update, options.Replace().SetUpsert(upsert))
	} else if many {
		updated, e = collection.UpdateMany(ctx, filter, update, options.UpdateMany().SetUpsert(upsert))
	} else {
		updated, e = collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(upsert))
	}
	if e != nil {
		return nil, e
	}
	return map[string]interface{}{
		"matched":  updated.MatchedCount,
		"affected": updated.ModifiedCount,
		"upserted": updated.UpsertedCount,
	}, nil
}

// Delete 删除指定数据。
func (m *Manager) Delete(ctx context.Context, collectionName, filterJSON string, many bool) (result interface{}, err error) {
	if len(filterJSON) > 1<<20 {
		return nil, errors.New("筛选条件超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	filter, e := m.parse(filterJSON, "筛选条件", true)
	if e != nil {
		return nil, e
	}
	var deleted *client.DeleteResult
	if many {
		deleted, e = collection.DeleteMany(ctx, filter)
	} else {
		deleted, e = collection.DeleteOne(ctx, filter)
	}
	if e != nil {
		return nil, e
	}
	return map[string]interface{}{"affected": deleted.DeletedCount}, nil
}

// Aggregate 执行读取型聚合。
func (m *Manager) Aggregate(ctx context.Context, collectionName, pipelineJSON string, limit int) (result interface{}, err error) {
	if limit < 0 {
		return nil, errors.New("返回上限不合法")
	}
	if len(pipelineJSON) > 1<<20 {
		return nil, errors.New("聚合管道超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	var raw []json.RawMessage
	if json.Unmarshal([]byte(pipelineJSON), &raw) != nil || raw == nil || len(raw) > 1000 {
		return nil, errors.New("聚合管道必须是最多包含 1000 个阶段的 JSON 数组")
	}
	pipeline := make(client.Pipeline, 0, len(raw)+1)
	for _, item := range raw {
		stage, e := m.parse(string(item), "聚合阶段", true)
		if e != nil || len(stage) != 1 {
			return nil, errors.New("每个聚合阶段必须包含一个操作")
		}
		if stage[0].Key == "$out" || stage[0].Key == "$merge" {
			return nil, errors.New("写入型聚合请通过命令功能明确执行")
		}
		pipeline = append(pipeline, stage)
	}
	pipeline = append(pipeline, bson.D{{
		Key:   "$limit",
		Value: limit + 1,
	}})
	started := time.Now()
	queryCtx, cancelQuery := context.WithTimeout(ctx, 25*time.Second)
	defer cancelQuery()
	cursor, e := collection.Aggregate(queryCtx, pipeline, options.Aggregate().SetBatchSize(50))
	if e != nil {
		return nil, e
	}
	data, e := m.readCursor(queryCtx, cursor, limit, 1)
	if e != nil {
		return nil, e
	}
	data["elapsed_ms"] = time.Since(started).Milliseconds()
	return data, nil
}

// Command 执行有界命令。
func (m *Manager) Command(ctx context.Context, commandJSON string) (result interface{}, err error) {
	if len(commandJSON) > 1<<20 {
		return nil, errors.New("命令超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	command, e := m.parse(commandJSON, "数据库命令", true)
	if e != nil {
		return nil, e
	}
	// v2 根据 context 自动附加 maxTimeMS，用户指定的上限转为 context，避免重复字段。
	fields := command[:0]
	for _, field := range command {
		if field.Key != "maxTimeMS" {
			fields = append(fields, field)
			continue
		}
		var milliseconds int64
		switch value := field.Value.(type) {
		case int32:
			milliseconds = int64(value)
		case int64:
			milliseconds = value
		case float64:
			milliseconds = int64(value)
			if float64(milliseconds) != value {
				return nil, errors.New("maxTimeMS 必须是非负整数")
			}
		default:
			return nil, errors.New("maxTimeMS 必须是非负整数")
		}
		if milliseconds < 0 {
			return nil, errors.New("maxTimeMS 必须是非负整数")
		}
		if milliseconds > 0 && milliseconds < 30000 {
			var cancelCommand context.CancelFunc
			ctx, cancelCommand = context.WithTimeout(ctx, time.Duration(milliseconds)*time.Millisecond)
			defer cancelCommand()
		}
	}
	command = fields
	if len(command) == 0 {
		return nil, errors.New("请填写数据库命令")
	}
	if strings.EqualFold(command[0].Key, "shutdown") {
		return nil, errors.New("关闭数据库请通过软件管理或终端操作")
	}
	started := time.Now()
	cursorCommand := false
	switch command[0].Key {
	case "find", "aggregate", "listCollections", "listIndexes":
		cursorCommand = true
	}
	if command[0].Key == "aggregate" {
		for _, option := range command[1:] {
			if option.Key == "explain" && option.Value == true {
				cursorCommand = false
				break
			}
		}
	}
	if cursorCommand {
		// 游标命令只执行一次，后续批次由驱动获取，返回上限与界面保持一致。
		// 不向聚合管道追加阶段，保留 $out、$merge 必须位于末尾的语义。
		cursor, e := m.database.RunCommandCursor(ctx, command)
		if e != nil {
			return nil, e
		}
		data, e := m.readCursor(ctx, cursor, 500, 1)
		if e != nil {
			return nil, e
		}
		data["elapsed_ms"] = time.Since(started).Milliseconds()
		return data, nil
	}
	var document bson.Raw
	if e = m.database.RunCommand(ctx, command).Decode(&document); e != nil {
		// 驱动会将 ok:1 中的 writeErrors/writeConcernError 转为 WriteException。
		var writeError client.WriteException
		if errors.As(e, &writeError) {
			return nil, fmt.Errorf("命令写入失败，部分写入可能已生效：%w", e)
		}
		return nil, e
	}
	more := false
	if cursor, ok := document.Lookup("cursor").DocumentOK(); ok {
		if id, ok := cursor.Lookup("id").AsInt64OK(); ok && id != 0 {
			// 未知命令仍返回原始结果，但只保留首批并关闭服务端游标，不能重复执行命令。
			namespace, _ := cursor.Lookup("ns").StringValueOK()
			databaseName, collectionName, ok := strings.Cut(namespace, ".")
			if !ok || databaseName == "" || collectionName == "" {
				return nil, errors.New("命令返回的游标缺少有效命名空间，无法继续读取")
			}
			closeCtx, closeCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer closeCancel()
			if e = m.client.Database(databaseName).RunCommand(closeCtx, bson.D{{
				Key:   "killCursors",
				Value: collectionName,
			}, {
				Key:   "cursors",
				Value: bson.A{id},
			}}).Err(); e != nil {
				return nil, fmt.Errorf("命令已执行，但关闭结果游标失败：%w", e)
			}
			more = true
		}
	}
	value, length, e := m.decode(document)
	if e != nil {
		return nil, e
	}
	if length > 8<<20 {
		return map[string]interface{}{
			"result":     "返回内容超过 8 MB，请限制命令返回数量或字段。",
			"truncated":  true,
			"elapsed_ms": time.Since(started).Milliseconds(),
		}, nil
	}
	data := map[string]interface{}{
		"result":     value,
		"truncated":  more,
		"has_more":   more,
		"elapsed_ms": time.Since(started).Milliseconds(),
	}
	if more {
		data["warning"] = "此命令仅显示首批结果，结果游标已关闭；查询文档请使用 find 或 aggregate。"
	}
	return data, nil
}

// CreateCollection 创建集合。
func (m *Manager) CreateCollection(ctx context.Context, collectionName string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	return nil, m.database.CreateCollection(ctx, collection.Name())
}

// CreateIndex 创建索引。
func (m *Manager) CreateIndex(ctx context.Context, collectionName, name, keysJSON string, unique, sparse bool, expireAfterSeconds *int32) (result interface{}, err error) {
	if len(name) > 255 || len(keysJSON) > 65536 {
		return nil, errors.New("索引参数超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	keys, e := m.parse(keysJSON, "索引字段", true)
	if e != nil {
		return nil, e
	}
	opts := options.Index().SetUnique(unique).SetSparse(sparse)
	if name != "" {
		opts.SetName(name)
	}
	if expireAfterSeconds != nil {
		if *expireAfterSeconds < 0 {
			return nil, errors.New("索引有效期不能小于 0")
		}
		opts.SetExpireAfterSeconds(*expireAfterSeconds)
	}
	createdName, e := collection.Indexes().CreateOne(ctx, client.IndexModel{
		Keys:    keys,
		Options: opts,
	})
	return map[string]interface{}{"name": createdName}, e
}

// DropIndex 删除指定索引。
func (m *Manager) DropIndex(ctx context.Context, collectionName, name string) (result interface{}, err error) {
	if len(name) > 255 {
		return nil, errors.New("索引名称超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	collection, err := m.collection(collectionName)
	if err != nil {
		return nil, err
	}
	if name == "" || name == "*" {
		return nil, errors.New("请选择具体索引")
	}
	return nil, collection.Indexes().DropOne(ctx, name)
}
