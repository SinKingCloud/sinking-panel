package elasticsearch

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// unsupportedRequest 仅对缺失端点或明确不支持的请求回退，保留认证及连接错误。
func unsupportedRequest(status int, err error) bool {
	if err == nil {
		return false
	}
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed || status == http.StatusNotImplemented {
		return true
	}
	if status != http.StatusBadRequest {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no handler found") || strings.Contains(text, "unknown parameter") || strings.Contains(text, "unrecognized parameter") || strings.Contains(text, "unsupported parameter")
}

// indexStats 0.90 使用布尔统计参数，1.x 以后使用指标路径；受限代理可回退完整统计。
func (m *Manager) indexStats(ctx context.Context, index string) (map[string]interface{}, error) {
	path := ""
	if index != "" {
		if err := ValidateIndexName(index, false); err != nil {
			return nil, err
		}
		path = "/" + url.PathEscape(index)
	}
	endpoint := path + "/_stats/docs,store"
	if m.major == 0 {
		endpoint = path + "/_stats?clear=true&docs=true&store=true"
	}
	data, status, err := m.doRequest(ctx, http.MethodGet, endpoint, nil)
	if unsupportedRequest(status, err) {
		data, _, err = m.doRequest(ctx, http.MethodGet, path+"/_stats", nil)
	}
	if err != nil {
		return nil, err
	}
	indices := object(object(data)["indices"])
	if indices == nil {
		return nil, errors.New("Elasticsearch 统计响应缺少 indices")
	}
	return indices, nil
}

// metadataIndices 仅用于旧版缺失索引详情接口的补充，0.90 支持按索引过滤元数据。
func (m *Manager) metadataIndices(ctx context.Context, index string) (map[string]interface{}, error) {
	query := url.Values{"filter_nodes": {"true"}, "filter_routing_table": {"true"}, "filter_blocks": {"true"}}
	if index != "" {
		if err := ValidateIndexName(index, false); err != nil {
			return nil, err
		}
		query.Set("filter_indices", index)
	}
	endpoint := "/_cluster/state/metadata"
	if index != "" {
		endpoint += "/" + url.PathEscape(index)
	}
	if m.major == 0 {
		endpoint = "/_cluster/state?" + query.Encode()
	}
	data, status, err := m.doRequest(ctx, http.MethodGet, endpoint, nil)
	if m.major != 0 && unsupportedRequest(status, err) {
		data, _, err = m.doRequest(ctx, http.MethodGet, "/_cluster/state?"+query.Encode(), nil)
	}
	if err != nil {
		return nil, err
	}
	indices := object(object(object(data)["metadata"])["indices"])
	if indices == nil {
		return nil, errors.New("Elasticsearch 元数据响应缺少 indices")
	}
	return indices, nil
}

// indexSetting 兼容 0.90 的 index.xxx 平面键和后续版本的 index 对象。
func indexSetting(settings map[string]interface{}, name string) interface{} {
	if value, exists := object(settings["index"])[name]; exists {
		return value
	}
	if value, exists := settings["index."+name]; exists {
		return value
	}
	return settings[name]
}

// legacyIndices 在 CAT 缺失或不支持 JSON 时组合设置、统计及健康信息，包含关闭的索引。
func (m *Manager) legacyIndices(ctx context.Context, keyword string) ([]Index, error) {
	data, status, err := m.doRequest(ctx, http.MethodGet, "/_settings", nil)
	settings := object(data)
	if unsupportedRequest(status, err) {
		settings, err = m.metadataIndices(ctx, "")
	}
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.New("Elasticsearch 索引设置响应格式不正确")
	}
	stats, err := m.indexStats(ctx, "")
	if err != nil {
		return nil, err
	}
	healthData, _, err := m.doRequest(ctx, http.MethodGet, "/_cluster/health?level=indices", nil)
	if err != nil {
		return nil, err
	}
	health := object(object(healthData)["indices"])
	list := make([]Index, 0, len(settings))
	for name, value := range settings {
		if !visibleIndex(name, keyword) {
			continue
		}
		entry := object(value)
		state, _ := entry["state"].(string)
		if state == "close" {
			state = "closed"
		}
		if state == "" {
			state = "open"
		}
		indexHealth, _ := object(health[name])["status"].(string)
		if object(stats[name]) == nil && health[name] == nil {
			state = "closed"
		}
		indexSettings := object(entry["settings"])
		primary := object(object(stats[name])["primaries"])
		total := object(object(stats[name])["total"])
		list = append(list, Index{Name: name, Health: indexHealth, Status: state,
			Documents: number(object(primary["docs"])["count"]), SizeBytes: number(object(total["store"])["size_in_bytes"]),
			PrimaryShards: int(number(indexSetting(indexSettings, "number_of_shards"))), Replicas: int(number(indexSetting(indexSettings, "number_of_replicas")))})
	}
	return list, nil
}

func visibleIndex(name, keyword string) bool {
	return ValidateIndexName(name, false) == nil && !strings.HasPrefix(name, ".") && strings.Contains(strings.ToLower(name), strings.ToLower(keyword))
}
