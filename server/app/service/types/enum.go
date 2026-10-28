package types

import (
	"server/app/constant"
	"strconv"
	"strings"
)

// GetEnum 获取分类枚举；实例范围包含全局分类，也支持 script_database% 前缀查询。
func (s *service) GetEnum(module string) (map[int64]string, error) {
	// 模糊查询直接读取，避免任意前缀缓存无法随分类变更精确失效。
	if strings.Contains(module, "%") {
		return s.repositoryTypes.SelectIdNameMap(module)
	}
	modules := []string{module}
	if index := strings.LastIndexByte(module, '_'); index > 0 {
		if _, err := strconv.ParseUint(module[index+1:], 10, 64); err == nil {
			modules = append(modules, module[:index])
		}
	}
	result := make(map[int64]string)
	for _, current := range modules {
		// 每个范围独立缓存，实例枚举实时组合，修改全局分类只需清理全局缓存。
		cacheKey := constant.CacheNameWithTypeEnum + current
		data, ok := s.cache.Get(cacheKey).(map[int64]string)
		if !ok {
			var err error
			data, err = s.repositoryTypes.SelectIdNameMap(current)
			if err != nil {
				return nil, err
			}
			s.cache.SetWithExpire(cacheKey, data, constant.CacheTimeWithTypeEnum)
		}
		for id, name := range data {
			result[id] = name
		}
	}
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
