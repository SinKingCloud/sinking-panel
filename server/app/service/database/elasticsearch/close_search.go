package elasticsearch

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// CloseSearch 仅释放指定分页会话，不允许关闭其他查询或使用 _all。
func (m *Manager) CloseSearch(ctx context.Context, pagination, searchID string) error {
	if err := m.validateSearchID(searchID); err != nil {
		return err
	}
	path := "/_search/scroll"
	var body interface{} = map[string]interface{}{"scroll_id": []string{searchID}}
	switch pagination {
	case "pit":
		if !m.atLeast(7, 12) {
			return errors.New("当前版本不支持 PIT 分页")
		}
		path, body = "/_pit", map[string]interface{}{"id": searchID}
	case "scroll":
		if m.major < 2 {
			path, body = path+"/"+url.PathEscape(searchID), nil
		}
	default:
		return errors.New("只允许关闭 PIT 或 scroll 分页会话")
	}
	_, status, err := m.doRequest(ctx, http.MethodDelete, path, body)
	if status == http.StatusNotFound {
		return nil // 会话已过期时关闭同样视为完成。
	}
	return err
}

func (m *Manager) validateSearchID(id string) error {
	if id == "" || len(id) > 131072 || id == "_all" {
		return errors.New("分页会话不合法，请刷新后重新查询")
	}
	for _, c := range id {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '+' || c == '/' || c == '=' || c == '-' || c == '_' {
			continue
		}
		return errors.New("分页会话不合法，请刷新后重新查询")
	}
	return nil
}
