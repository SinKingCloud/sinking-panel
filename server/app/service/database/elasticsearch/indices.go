package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"server/app/util/page"
	"sort"
	"time"
)

// Indices 返回普通索引及统计，系统索引保持只读并不加入工作区目录。
func (m *Manager) Indices(ctx context.Context, keyword string, pageNumber, pageSize int, orderByType string) (*page.Result[Index], error) {
	if pageNumber == 0 {
		pageNumber = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if pageNumber < 1 || pageNumber > 1000000 || pageSize < 1 || pageSize > 1000 {
		return nil, errors.New("页码或每页数量不合法")
	}
	if len(keyword) > 4096 {
		return nil, errors.New("索引名称关键词过长")
	}
	if orderByType == "" {
		orderByType = "asc"
	}
	if orderByType != "asc" && orderByType != "desc" {
		return nil, errors.New("排序类型不合法")
	}
	var list []Index
	if m.major == 0 {
		var err error
		list, err = m.legacyIndices(ctx, keyword)
		if err != nil {
			return nil, err
		}
	} else {
		data, status, err := m.doRequest(ctx, http.MethodGet, "/_cat/indices?format=json&bytes=b&h=index,health,status,docs.count,store.size,pri,rep", nil)
		_, textFormat := data.(string)
		if unsupportedRequest(status, err) || err == nil && textFormat {
			list, err = m.legacyIndices(ctx, keyword)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			values, ok := data.([]interface{})
			if !ok {
				return nil, errors.New("Elasticsearch 索引列表格式不正确")
			}
			list = make([]Index, 0, len(values))
			for _, value := range values {
				item := object(value)
				name, _ := item["index"].(string)
				if !visibleIndex(name, keyword) {
					continue
				}
				health, _ := item["health"].(string)
				status, _ := item["status"].(string)
				list = append(list, Index{Name: name, Health: health, Status: status, Documents: number(item["docs.count"]), SizeBytes: number(item["store.size"]), PrimaryShards: int(number(item["pri"])), Replicas: int(number(item["rep"]))})
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if orderByType == "desc" {
			return list[i].Name > list[j].Name
		}
		return list[i].Name < list[j].Name
	})
	// CAT 目录没有分页参数；在兼容旧版本的目录请求后统一排序和分页。
	total := int64(len(list))
	if total == 0 {
		pageNumber = 1
	} else if lastPage := (total-1)/int64(pageSize) + 1; int64(pageNumber) > lastPage {
		pageNumber = int(lastPage)
	}
	start := (pageNumber - 1) * pageSize
	end := min(start+pageSize, len(list))
	return page.New(&page.Query{Page: pageNumber, PageSize: pageSize}, total, append([]Index{}, list[start:end]...)), nil
}

// Info 返回选中具体索引的 Mapping、Settings 和可读统计。
func (m *Manager) Info(ctx context.Context) (map[string]interface{}, error) {
	path, err := m.indexPath(false)
	if err != nil {
		return nil, err
	}
	var index map[string]interface{}
	if m.atLeast(1, 4) {
		data, status, requestErr := m.doRequest(ctx, http.MethodGet, path, nil)
		if requestErr != nil && !unsupportedRequest(status, requestErr) {
			return nil, requestErr
		}
		if requestErr == nil {
			index = object(object(data)[m.index])
			if index == nil {
				return nil, errors.New("未返回选中索引的详情")
			}
		}
	}
	if index == nil {
		mapping, mappingErr := m.Mapping(ctx)
		if mappingErr != nil {
			return nil, mappingErr
		}
		settings, settingsErr := m.Settings(ctx)
		if settingsErr != nil {
			return nil, settingsErr
		}
		metadata, metadataErr := m.metadataIndices(ctx, m.index)
		if metadataErr != nil {
			return nil, metadataErr
		}
		entry := object(metadata[m.index])
		if entry == nil {
			return nil, errors.New("未返回选中索引的元数据")
		}
		index = map[string]interface{}{"mappings": mapping, "settings": settings, "aliases": entry["aliases"]}
	}
	result := map[string]interface{}{"name": m.index, "mappings": index["mappings"], "settings": index["settings"], "aliases": index["aliases"]}
	stats, statsErr := m.indexStats(ctx, m.index)
	if statsErr == nil {
		result["stats"] = stats[m.index]
		primaries := object(object(result["stats"])["primaries"])
		result["documents"] = object(primaries["docs"])["count"]
		result["size_bytes"] = object(primaries["store"])["size_in_bytes"]
	} else {
		result["stats_error"] = statsErr.Error()
	}
	return result, nil
}

// CreateIndex 创建单个普通索引，可选 Settings 与 Mapping 保持原生 JSON 表达。
func (m *Manager) CreateIndex(ctx context.Context, name, settings, mapping string) (map[string]interface{}, error) {
	if name == "" {
		name = m.index
	}
	if err := ValidateIndexName(name, true); err != nil {
		return nil, err
	}
	settingsValue, err := decodeObject(settings, "索引设置", true)
	if err != nil {
		return nil, err
	}
	mappingValue, err := decodeObject(mapping, "字段映射", true)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{}
	if len(settingsValue) > 0 {
		body["settings"] = settingsValue
	}
	if len(mappingValue) > 0 {
		if m.major < 7 && mappingDefinition(mappingValue) {
			typeName := "_doc"
			if !m.atLeast(6, 2) {
				typeName = "doc"
			}
			mappingValue = map[string]interface{}{typeName: mappingValue}
		}
		body["mappings"] = mappingValue
	}
	data, _, err := m.doRequest(ctx, http.MethodPut, "/"+url.PathEscape(name), body)
	return object(data), err
}

// DropIndex 删除一个普通索引，不接受通配符、别名列表或系统索引。
func (m *Manager) DropIndex(ctx context.Context, name string) (map[string]interface{}, error) {
	if name == "" {
		name = m.index
	}
	if err := ValidateIndexName(name, true); err != nil {
		return nil, err
	}
	data, _, err := m.doRequest(ctx, http.MethodDelete, "/"+url.PathEscape(name), nil)
	return object(data), err
}

// DropIndices 先校验整批索引名称，再逐个删除并保留部分成功信息。
func (m *Manager) DropIndices(ctx context.Context, names []string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if len(names) == 0 {
		return errors.New("请选择索引")
	}
	for _, name := range names {
		if err := ValidateIndexName(name, true); err != nil {
			return err
		}
	}
	for completed, name := range names {
		if _, err := m.DropIndex(ctx, name); err != nil {
			return fmt.Errorf("已删除 %d 个索引，删除 %s 失败：%w", completed, name, err)
		}
	}
	return nil
}
