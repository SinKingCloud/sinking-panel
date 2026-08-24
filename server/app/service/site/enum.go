package site

import "server/app/constant"

// GetIdNameMap 获取可选默认站点的 ID 名称映射。
func (s *service) GetIdNameMap(refresh bool) (map[int64]string, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if !refresh {
		if result, ok := s.cache.Get(constant.CacheNameWithSiteNameEnum).(map[int64]string); ok {
			return result, nil
		}
	}
	result, err := s.repositorySite.SelectIdNameMap()
	if err != nil {
		return nil, err
	}
	s.cache.SetWithExpire(constant.CacheNameWithSiteNameEnum, result, constant.CacheTimeWithSiteNameEnum)
	return result, nil
}
