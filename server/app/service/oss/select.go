package oss

import (
	"fmt"
	repositoryOss "server/app/repository/oss"
	repositorySecret "server/app/repository/secret"
	"server/app/util/page"
)

// Select 分页查询对象存储，批量补充关联密钥的名称和厂商。
func (s *service) Select(where *repositoryOss.SelectOss, query *page.Query) (*page.Result[*OssListItem], error) {
	result, err := s.repository.Select(where, query)
	if err != nil {
		return nil, err
	}

	secretIds := make([]int64, 0, len(result.List))
	secrets := make(map[int64]*repositorySecret.Secret, len(result.List))
	for _, item := range result.List {
		if item.SecretId > 0 {
			if _, exists := secrets[item.SecretId]; !exists {
				secretIds = append(secretIds, item.SecretId)
				secrets[item.SecretId] = nil
			}
		}
	}
	if len(secretIds) > 0 {
		records, err := s.repositorySecret.SelectByIds(secretIds)
		if err != nil {
			return nil, fmt.Errorf("查询关联密钥失败: %w", err)
		}
		for _, record := range records {
			secrets[record.Id] = record
		}
	}

	list := make([]*OssListItem, len(result.List))
	for i, item := range result.List {
		list[i] = &OssListItem{Oss: item}
		if secret := secrets[item.SecretId]; secret != nil {
			list[i].SecretName = secret.Name
			list[i].Provider = &secret.Provider
		}
	}
	return &page.Result[*OssListItem]{
		Total:            result.Total,
		Page:             result.Page,
		PageSize:         result.PageSize,
		List:             list,
		HasMore:          result.HasMore,
		NextCursorId:     result.NextCursorId,
		NextCursorLastId: result.NextCursorLastId,
	}, nil
}
