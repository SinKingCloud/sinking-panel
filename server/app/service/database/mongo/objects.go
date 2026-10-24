package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	client "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// viewPipeline 保留 Extended JSON 类型；空数组表示不转换源数据。
func (m *Manager) viewPipeline(text string) (client.Pipeline, error) {
	if strings.TrimSpace(text) == "" || len(text) > 1<<20 {
		return nil, errors.New("视图管道必须是非空 JSON 字符串，且不能超过 1 MB")
	}
	var values []json.RawMessage
	if json.Unmarshal([]byte(text), &values) != nil || values == nil || len(values) > 100 {
		return nil, errors.New("视图管道必须是 JSON 数组，最多 100 个阶段")
	}
	pipeline := make(client.Pipeline, 0, len(values))
	for _, value := range values {
		stage, err := m.parse(string(value), "视图阶段", true)
		if err != nil || len(stage) != 1 || !strings.HasPrefix(stage[0].Key, "$") {
			return nil, errors.New("每个视图阶段必须包含一个聚合操作")
		}
		if stage[0].Key == "$out" || stage[0].Key == "$merge" {
			return nil, errors.New("视图不支持写入型聚合阶段")
		}
		pipeline = append(pipeline, stage)
	}
	return pipeline, nil
}

// CreateView 创建真实只读视图，源集合或视图必须位于当前数据库。
func (m *Manager) CreateView(ctx context.Context, viewName, sourceCollection, pipelineJSON string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	database := m.database.Name()
	if strings.EqualFold(database, "admin") || strings.EqualFold(database, "config") || strings.EqualFold(database, "local") {
		return nil, errors.New("不能在系统数据库中创建视图")
	}
	for _, name := range []string{viewName, sourceCollection} {
		if _, err = m.collection(name); err != nil {
			return nil, err
		}
		if len(database)+1+len(name) > 255 || strings.Contains(name, "$") || strings.HasPrefix(name, "system.") || strings.Contains(name, ".system.") {
			return nil, errors.New("不能使用系统集合或无效视图名称")
		}
	}
	if viewName == sourceCollection {
		return nil, errors.New("视图不能引用自身")
	}
	pipeline, err := m.viewPipeline(pipelineJSON)
	if err != nil {
		return nil, err
	}
	cursor, err := m.database.ListCollections(ctx, bson.D{{Key: "name", Value: bson.D{{Key: "$in", Value: bson.A{viewName, sourceCollection}}}}}, options.ListCollections().SetBatchSize(2))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	sourceExists := false
	var collation *options.Collation
	for cursor.Next(ctx) {
		document := cursor.Current
		name, _ := document.Lookup("name").StringValueOK()
		kind, _ := document.Lookup("type").StringValueOK()
		if name == viewName {
			return nil, errors.New("视图名称已被集合或视图使用")
		}
		if name != sourceCollection {
			continue
		}
		if kind != "collection" && kind != "view" && kind != "timeseries" {
			return nil, errors.New("源对象不是可查询的集合或视图")
		}
		sourceExists = true
		if kind == "view" {
			if value := document.Lookup("options", "collation"); value.Type != 0 {
				collation = &options.Collation{}
				if err = value.Unmarshal(collation); err != nil {
					return nil, errors.New("源视图排序规则格式不正确")
				}
			}
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	if !sourceExists {
		return nil, errors.New("源集合或视图不存在，或当前账号无权访问")
	}
	settings := options.CreateView()
	if collation != nil {
		settings.SetCollation(collation)
	}
	if err = m.database.CreateView(ctx, viewName, sourceCollection, pipeline, settings); err != nil {
		return nil, err
	}
	decoded, _, err := m.decode(bson.D{{Key: "pipeline", Value: pipeline}})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": viewName, "source": sourceCollection, "pipeline": decoded.(map[string]interface{})["pipeline"]}, nil
}

// ViewDefinition 返回现有视图的源名称和管道，普通集合不会被误认为视图。
func (m *Manager) ViewDefinition(ctx context.Context, viewName string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	database := m.database.Name()
	if strings.EqualFold(database, "admin") || strings.EqualFold(database, "config") || strings.EqualFold(database, "local") {
		return nil, errors.New("不能管理系统数据库中的视图")
	}
	if _, err = m.collection(viewName); err != nil {
		return nil, err
	}
	if len(database)+1+len(viewName) > 255 || strings.Contains(viewName, "$") || strings.HasPrefix(viewName, "system.") || strings.Contains(viewName, ".system.") {
		return nil, errors.New("不能访问系统视图或无效视图名称")
	}
	cursor, err := m.database.ListCollections(ctx, bson.D{{Key: "name", Value: viewName}}, options.ListCollections().SetBatchSize(1))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if !cursor.Next(ctx) {
		if err = cursor.Err(); err != nil {
			return nil, err
		}
		return nil, errors.New("视图不存在或当前账号无权访问")
	}
	document := cursor.Current
	kind, _ := document.Lookup("type").StringValueOK()
	if kind != "view" {
		return nil, errors.New("当前对象是集合，不能作为视图管理")
	}
	source, ok := document.Lookup("options", "viewOn").StringValueOK()
	if !ok || source == "" {
		return nil, errors.New("服务器未返回完整的视图定义")
	}
	if strings.HasPrefix(source, "system.") || strings.Contains(source, ".system.") {
		return nil, errors.New("不能管理引用系统集合的视图")
	}
	pipeline := make(client.Pipeline, 0)
	if value := document.Lookup("options", "pipeline"); value.Type != 0 {
		if err = value.Unmarshal(&pipeline); err != nil {
			return nil, errors.New("服务器返回的视图管道格式不正确")
		}
	}
	decoded, length, err := m.decode(bson.D{{Key: "pipeline", Value: pipeline}})
	if err != nil {
		return nil, err
	}
	if length > 1<<20 {
		return nil, errors.New("视图定义超过 1 MB")
	}
	return map[string]interface{}{"name": viewName, "source": source, "pipeline": decoded.(map[string]interface{})["pipeline"]}, nil
}

// UpdateView 使用 collMod 修改原视图管道，保留源对象及原排序规则。
func (m *Manager) UpdateView(ctx context.Context, viewName, pipelineJSON string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	pipeline, err := m.viewPipeline(pipelineJSON)
	if err != nil {
		return nil, err
	}
	current, err := m.ViewDefinition(ctx, viewName)
	if err != nil {
		return nil, err
	}
	source := current.(map[string]interface{})["source"].(string)
	var document bson.Raw
	if err = m.database.RunCommand(ctx, bson.D{{Key: "collMod", Value: viewName}, {Key: "viewOn", Value: source}, {Key: "pipeline", Value: pipeline}, {Key: "maxTimeMS", Value: 25000}}).Decode(&document); err != nil {
		return nil, err
	}
	decoded, _, err := m.decode(bson.D{{Key: "pipeline", Value: pipeline}})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": viewName, "source": source, "pipeline": decoded.(map[string]interface{})["pipeline"]}, nil
}
