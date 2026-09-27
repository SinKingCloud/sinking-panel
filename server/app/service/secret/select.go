package secret

import (
	"errors"
	"server/app/enum/secret_provider"
	repositorySecret "server/app/repository/secret"
	"server/app/util/page"
)

// Select 分页查询密钥元数据。
func (s *service) Select(where *repositorySecret.SelectSecret, queryPage *page.Query) (*page.Result[*repositorySecret.Secret], error) {
	if where != nil && where.Provider != nil {
		if _, ok := secret_provider.Map()[*where.Provider]; !ok {
			return nil, errors.New("服务商不合法")
		}
	}
	return s.repositorySecret.Select(where, queryPage)
}
