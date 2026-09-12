package secret

import (
	repositorySecret "server/app/repository/secret"
	"server/app/util/page"
)

// Select 分页查询密钥元数据。
func (s *service) Select(where *repositorySecret.SelectSecret, queryPage *page.Query) (*page.Result[*repositorySecret.Secret], error) {
	return s.repositorySecret.Select(where, queryPage)
}
