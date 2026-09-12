package secret

import (
	"encoding/json"
	"server/app/constant"
	repositorySecret "server/app/repository/secret"
)

// GetIdNameMap 按查询条件缓存密钥 ID 和名称映射。
func (s *service) GetIdNameMap(where *repositorySecret.SelectSecret) (map[int64]string, error) {
	s.enumMu.Lock()
	defer s.enumMu.Unlock()
	var query repositorySecret.SelectSecret
	if where != nil {
		query = *where
	}
	key, _ := json.Marshal(query)
	cached, _ := s.cache.Get(constant.CacheNameWithSecretNameEnum).(map[string]map[int64]string)
	if result, ok := cached[string(key)]; ok {
		return result, nil
	}
	result, err := s.repositorySecret.SelectIdNameMap(where)
	if err != nil {
		return nil, err
	}
	if cached == nil {
		cached = make(map[string]map[int64]string)
	}
	cached[string(key)] = result
	s.cache.SetWithExpire(constant.CacheNameWithSecretNameEnum, cached, constant.CacheTimeWithSecretNameEnum)
	return result, nil
}
