package types

import (
	"errors"
	"server/app/constant"
	"server/app/enum/type_module"
)

// GetEnum 获取指定模块的类型枚举
func (s *service) GetEnum(module string) (map[int64]string, error) {
	if _, ok := type_module.Map()[module]; !ok {
		return nil, errors.New("所属模块不合法")
	}
	cacheKey := constant.CacheNameWithTypeEnum + module
	if data := s.cache.Get(cacheKey); data != nil {
		if result, ok := data.(map[int64]string); ok {
			return result, nil
		}
	}
	result, err := s.repositoryTypes.SelectIdNameMap(module)
	if err != nil {
		return nil, err
	}
	s.cache.SetWithExpire(cacheKey, result, constant.CacheTimeWithTypeEnum)
	return result, nil
}

// clearTypeEnumCache 清理指定模块的类型枚举缓存
func (s *service) clearTypeEnumCache(modules ...string) {
	for _, module := range modules {
		if module != "" {
			s.cache.Delete(constant.CacheNameWithTypeEnum + module)
		}
	}
}
