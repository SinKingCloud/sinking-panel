package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Documents 普通分页保留 from/size；深分页使用 PIT 快照，旧版自动使用 scroll。
// PIT 游标由原始 sort 值组成，始终以 JSON 字符串传递，避免浏览器丢失大整数精度。
func (m *Manager) Documents(ctx context.Context, query, pagination, searchID, cursor string, page, pageSize int) (*DocumentPage, error) {
	path, err := m.indexPath(false)
	if err != nil {
		return nil, err
	}
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}
	if page < 1 || page > 1000000000 || pageSize < 1 || pageSize > 500 {
		return nil, errors.New("页码应为正整数，分页容量应为 1 至 500")
	}
	if pagination == "" {
		pagination = "offset"
	}
	if pagination != "offset" && pagination != "cursor" && pagination != "pit" && pagination != "scroll" {
		return nil, errors.New("分页方式不合法")
	}
	if pagination == "offset" && (page > 10000/pageSize || searchID != "" || cursor != "") {
		return nil, errors.New("普通分页查询窗口不能超过 10000 条，请使用游标分页")
	}
	if pagination == "cursor" {
		if searchID != "" || cursor != "" || page != 1 {
			return nil, errors.New("请从第一页创建新的分页会话")
		}
		pagination = "scroll"
		// PIT 在 7.10 提供，但稳定的 _shard_doc 排序在 7.12 才提供。
		if m.atLeast(7, 12) {
			pagination = "pit"
		}
	} else if pagination != "offset" {
		if err := validateSearchID(searchID); err != nil {
			return nil, err
		}
		if pagination == "scroll" && (cursor != "" || page < 2) {
			return nil, errors.New("旧版滚动分页只能依次获取下一页，请刷新后重新查询")
		}
	}
	if pagination == "pit" && !m.atLeast(7, 12) {
		return nil, errors.New("当前版本不支持稳定的 PIT 分页，请重新查询")
	}
	body, err := decodeObject(query, "搜索 DSL", true)
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"search_after", "pit", "scroll"} {
		if _, exists := body[key]; exists {
			return nil, errors.New("搜索 DSL 不能设置 " + key + "，请使用分页参数")
		}
	}
	if pagination != "offset" {
		for _, key := range []string{"from", "size", "collapse", "rescore", "slice", "terminate_after", "knn", "retriever"} {
			if _, exists := body[key]; exists {
				return nil, errors.New("游标分页不支持搜索 DSL 中的 " + key + "，请移除该参数或使用请求执行")
			}
		}
	}
	// 额外请求路由元信息，但不改变调用者原有的字段投影或 _source 选择。
	fieldKey := "stored_fields"
	if m.major < 5 {
		fieldKey = "fields"
	}
	stored, exists := body[fieldKey]
	fields := make([]interface{}, 0)
	if exists {
		switch value := stored.(type) {
		case string:
			fields = append(fields, value)
		case []interface{}:
			fields = append(fields, value...)
		default:
			return nil, errors.New(fieldKey + " 必须是字段名或字段名数组")
		}
	} else if _, exists := body["_source"]; !exists {
		body["_source"] = true
	}
	hasRouting := false
	for _, field := range fields {
		name, ok := field.(string)
		if !ok || name == "_none_" {
			return nil, errors.New(fieldKey + " 必须保留文档元信息，不能使用 _none_")
		}
		if name == "_routing" {
			hasRouting = true
		}
	}
	if !hasRouting {
		fields = append(fields, "_routing")
	}
	body[fieldKey], body["size"] = fields, pageSize
	if m.major >= 7 {
		body["track_total_hits"] = true
	}
	if body["timeout"] == nil {
		body["timeout"] = "25s"
	}
	if pagination == "offset" {
		body["from"] = (page - 1) * pageSize
	}
	if pagination == "pit" {
		sorts := make([]interface{}, 0)
		switch value := body["sort"].(type) {
		case nil:
			sorts = append(sorts, map[string]interface{}{"_score": "desc"})
		case string:
			sorts = append(sorts, value)
		case map[string]interface{}:
			// 多字段对象的键顺序无法作为稳定排序依据，请用数组显式指定。
			if len(value) != 1 {
				return nil, errors.New("多个排序字段请使用 sort 数组指定先后顺序")
			}
			sorts = append(sorts, value)
		case []interface{}:
			sorts = append(sorts, value...)
		default:
			return nil, errors.New("sort 必须是字段名、排序对象或数组")
		}
		if len(sorts) == 0 {
			sorts = append(sorts, map[string]interface{}{"_score": "desc"})
		}
		hasTieBreaker := false
		for _, sort := range sorts {
			if sort == "_shard_doc" || object(sort)["_shard_doc"] != nil {
				hasTieBreaker = true
			}
		}
		if !hasTieBreaker {
			sorts = append(sorts, map[string]interface{}{"_shard_doc": "asc"})
		}
		if len(sorts) > 64 {
			return nil, errors.New("排序字段不能超过 64 个")
		}
		body["sort"] = sorts
		if cursor != "" {
			var values []interface{}
			decoder := json.NewDecoder(strings.NewReader(cursor))
			decoder.UseNumber()
			if len(cursor) > 65536 || !json.Valid([]byte(cursor)) || decoder.Decode(&values) != nil || len(values) != len(sorts) {
				return nil, errors.New("分页游标不合法，请重新查询")
			}
			for _, value := range values {
				switch value.(type) {
				case nil, string, json.Number, bool:
				default:
					return nil, errors.New("分页游标只允许排序值，请重新查询")
				}
			}
			body["search_after"] = values
		}
	}
	// 失败的 scroll 可能已经前移，必须关闭，不能把不完整页当作下一页继续查询。
	completed := false
	defer func() {
		if !completed && searchID != "" {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = m.CloseSearch(cleanupCtx, pagination, searchID)
		}
	}()
	var data interface{}
	var status int
	switch pagination {
	case "pit":
		if searchID == "" {
			data, _, err = m.doRequest(ctx, http.MethodPost, path+"/_pit?keep_alive=2m", nil)
			if err != nil {
				return nil, err
			}
			searchID, _ = object(data)["id"].(string)
			if err = validateSearchID(searchID); err != nil {
				return nil, errors.New("Elasticsearch 未返回有效的分页会话")
			}
			if number(object(object(data)["_shards"])["failed"]) > 0 {
				return nil, errors.New("部分分片无法创建快照，请恢复分片后重新查询")
			}
		}
		body["pit"] = map[string]interface{}{"id": searchID, "keep_alive": "2m"}
		data, status, err = m.doRequest(ctx, http.MethodPost, "/_search", body)
	case "scroll":
		if searchID == "" {
			data, status, err = m.doRequest(ctx, http.MethodPost, path+"/_search?scroll=2m", body)
		} else if m.major < 2 {
			data, status, err = m.doRequest(ctx, http.MethodPost, "/_search/scroll?"+url.Values{"scroll": {"2m"}, "scroll_id": {searchID}}.Encode(), nil)
		} else {
			data, status, err = m.doRequest(ctx, http.MethodPost, "/_search/scroll", map[string]interface{}{"scroll": "2m", "scroll_id": searchID})
		}
	default:
		data, status, err = m.doRequest(ctx, http.MethodPost, path+"/_search", body)
	}
	result := object(data)
	if id, ok := result["pit_id"].(string); ok && pagination == "pit" {
		searchID = id
	}
	if id, ok := result["_scroll_id"].(string); ok && pagination == "scroll" {
		searchID = id
	}
	if err != nil {
		if pagination != "offset" && (status == http.StatusNotFound || strings.Contains(err.Error(), "search_context_missing_exception")) {
			return nil, errors.New("分页会话已过期或索引已变动，请刷新后重新查询")
		}
		return nil, err
	}
	if result["timed_out"] == true || result["terminated_early"] == true {
		return nil, errors.New("搜索未完整完成，请缩小查询范围后刷新重试")
	}
	if number(object(result["_shards"])["failed"]) > 0 {
		return nil, errors.New("搜索存在失败的分片，无法返回完整分页数据，请刷新重试")
	}
	hits := object(result["hits"])
	if hits == nil {
		return nil, errors.New("Elasticsearch 搜索响应缺少 hits")
	}
	values, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, errors.New("Elasticsearch 搜索响应缺少文档列表")
	}
	documents := make([]map[string]interface{}, 0, len(values))
	for _, value := range values {
		item := object(value)
		if item == nil {
			return nil, errors.New("Elasticsearch 返回了不完整的文档数据")
		}
		routing := item["_routing"]
		if routing == nil {
			routing = object(item["fields"])["_routing"]
		}
		if values, ok := routing.([]interface{}); ok && len(values) > 0 {
			routing = values[0]
		}
		if routing != nil {
			item["_routing"] = routing
		}
		documents = append(documents, item)
	}
	total, relation := number(hits["total"]), "eq"
	if count := object(hits["total"]); count != nil {
		total = number(count["value"])
		if text, ok := count["relation"].(string); ok {
			relation = text
		}
	}
	nextCursor := ""
	if pagination == "pit" && len(documents) > 0 {
		sort, ok := documents[len(documents)-1]["sort"].([]interface{})
		if !ok || len(sort) != len(body["sort"].([]interface{})) {
			return nil, errors.New("Elasticsearch 未返回完整排序游标，请重新查询")
		}
		encoded, err := json.Marshal(sort)
		if err != nil {
			return nil, err
		}
		nextCursor = string(encoded)
	}
	if pagination != "offset" {
		if err := validateSearchID(searchID); err != nil {
			return nil, errors.New("Elasticsearch 未返回有效的分页会话")
		}
	}
	hasMore := len(documents) == pageSize && (int64(page)*int64(pageSize) < total || relation == "gte")
	if pagination == "offset" {
		hasMore = hasMore && (page+1)*pageSize <= 10000
	}
	completed = true
	return &DocumentPage{Documents: documents, Total: total, TotalRelation: relation, Page: page, PageSize: pageSize,
		HasMore: hasMore, ElapsedMS: number(result["took"]), Pagination: pagination, SearchID: searchID, NextCursor: nextCursor}, nil
}

func validateDocumentID(id string) error {
	if id == "" || len(id) > 512 || !utf8.ValidString(id) || strings.IndexFunc(id, unicode.IsControl) >= 0 {
		return errors.New("文档 ID 应为 1 至 512 字节，不能包含控制字符")
	}
	return nil
}

func validateRouting(routing string) error {
	if len(routing) > 512 || !utf8.ValidString(routing) || strings.IndexFunc(routing, unicode.IsControl) >= 0 {
		return errors.New("文档路由值不能超过 512 字节或包含控制字符")
	}
	return nil
}

// documentType 旧版索引使用实际类型；新版仅使用 _doc，避免盲目写入另一文档类型。
func (m *Manager) documentType(ctx context.Context, requested string) (string, error) {
	if m.major >= 7 {
		return "_doc", nil
	}
	if requested != "" {
		if len(requested) > 255 || strings.ContainsAny(requested, "/\\?*%,# ") || strings.IndexFunc(requested, unicode.IsControl) >= 0 || requested == "." || requested == ".." || strings.HasPrefix(requested, "_") && (requested != "_doc" || !m.atLeast(6, 2)) {
			return "", errors.New("文档类型不合法")
		}
		return requested, nil
	}
	mappings, err := m.Mapping(ctx)
	if err != nil {
		return "", err
	}
	if mappingDefinition(mappings) {
		return "_doc", nil
	}
	types := make([]string, 0)
	for name := range mappings {
		if name != "_default_" {
			types = append(types, name)
		}
	}
	if len(types) > 1 {
		return "", errors.New("旧版索引包含多个文档类型，请指定 document_type")
	}
	if len(types) == 1 {
		return m.documentType(ctx, types[0])
	}
	if !m.atLeast(6, 2) {
		return "doc", nil
	}
	return "_doc", nil
}

// Insert 添加文档，指定 ID 时使用 op_type=create，防止覆盖已有文档。
func (m *Manager) Insert(ctx context.Context, documentID, document, documentType, routing string) (map[string]interface{}, error) {
	path, err := m.indexPath(true)
	if err != nil {
		return nil, err
	}
	if err = validateRouting(routing); err != nil {
		return nil, err
	}
	body, err := decodeObject(document, "文档内容", false)
	if err != nil {
		return nil, err
	}
	typeName, err := m.documentType(ctx, documentType)
	if err != nil {
		return nil, err
	}
	path += "/" + url.PathEscape(typeName)
	method := http.MethodPost
	if documentID != "" {
		if err := validateDocumentID(documentID); err != nil {
			return nil, err
		}
		path += "/" + url.PathEscape(documentID)
		method = http.MethodPut
	}
	params := url.Values{"op_type": {"create"}, "refresh": {"true"}}
	if routing != "" {
		params.Set("routing", routing)
	}
	data, _, err := m.doRequest(ctx, method, path+"?"+params.Encode(), body)
	return normalizeDocumentResult(data, "created"), err
}

// Update 完整替换已存在的 _source；使用读取到的版本防止并发覆盖或无意创建文档。
func (m *Manager) Update(ctx context.Context, documentID, document, documentType, routing, seqNo, primaryTerm, version string) (map[string]interface{}, error) {
	path, err := m.indexPath(true)
	if err != nil {
		return nil, err
	}
	if err = validateDocumentID(documentID); err != nil {
		return nil, err
	}
	if err = validateRouting(routing); err != nil {
		return nil, err
	}
	body, err := decodeObject(document, "文档内容", false)
	if err != nil {
		return nil, err
	}
	typeName, err := m.documentType(ctx, documentType)
	if err != nil {
		return nil, err
	}
	path += "/" + url.PathEscape(typeName) + "/" + url.PathEscape(documentID)
	params := url.Values{"refresh": []string{"true"}}
	if routing != "" {
		params.Set("routing", routing)
	}
	if m.atLeast(6, 7) && (seqNo != "" || primaryTerm != "") {
		seq, seqErr := strconv.ParseInt(seqNo, 10, 64)
		term, termErr := strconv.ParseInt(primaryTerm, 10, 64)
		if seqErr != nil || termErr != nil || seq < 0 || term < 1 {
			return nil, errors.New("文档并发版本不合法，请重新打开编辑")
		}
		params.Set("if_seq_no", strconv.FormatInt(seq, 10))
		params.Set("if_primary_term", strconv.FormatInt(term, 10))
	} else if version != "" {
		value, err := strconv.ParseInt(version, 10, 64)
		if err != nil || value < 1 {
			return nil, errors.New("文档版本不合法，请重新打开编辑")
		}
		params.Set("version", strconv.FormatInt(value, 10))
	} else {
		return nil, errors.New("缺少读取时的文档版本，请重新打开编辑")
	}
	data, _, err := m.doRequest(ctx, http.MethodPut, path+"?"+params.Encode(), body)
	return normalizeDocumentResult(data, "updated"), err
}

// Delete 删除一个具体文档，不提供批量匹配删除入口。
func (m *Manager) Delete(ctx context.Context, documentID, documentType, routing string) (map[string]interface{}, error) {
	path, err := m.indexPath(true)
	if err != nil {
		return nil, err
	}
	if err = validateDocumentID(documentID); err != nil {
		return nil, err
	}
	if err = validateRouting(routing); err != nil {
		return nil, err
	}
	if m.major < 7 && legacyDeleteEndpoint(documentID) {
		return nil, errors.New("旧版文档删除不允许 _query、_mapping 或 _mappings 作为 ID，避免批量删除")
	}
	typeName, err := m.documentType(ctx, documentType)
	if err != nil {
		return nil, err
	}
	params := url.Values{"refresh": {"true"}}
	if routing != "" {
		params.Set("routing", routing)
	}
	data, status, err := m.doRequest(ctx, http.MethodDelete, path+"/"+url.PathEscape(typeName)+"/"+url.PathEscape(documentID)+"?"+params.Encode(), nil)
	result := normalizeDocumentResult(data, "deleted")
	if (status == http.StatusNotFound || err == nil) && (result["found"] == false || result["result"] == "not_found") {
		return result, errors.New("Elasticsearch 文档不存在或已被删除")
	}
	return result, err
}

func legacyDeleteEndpoint(id string) bool {
	return id == "_query" || id == "_mapping" || id == "_mappings"
}

// normalizeDocumentResult 保留旧版 ok/created/found 并补齐现代 result，调用方无需猜响应年代。
func normalizeDocumentResult(data interface{}, operation string) map[string]interface{} {
	result := object(data)
	if result == nil || result["result"] != nil {
		return result
	}
	if found, exists := result["found"].(bool); exists && operation == "deleted" {
		if !found {
			operation = "not_found"
		}
	} else if created, exists := result["created"].(bool); exists {
		if created {
			operation = "created"
		} else {
			operation = "updated"
		}
	} else if result["ok"] != true && result["_version"] == nil {
		return result
	}
	result["result"] = operation
	return result
}
