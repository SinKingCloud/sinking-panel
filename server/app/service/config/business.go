package config

import (
	"errors"
	"strings"

	"server/app/constant"
	"server/app/model"
)

// Group 获取 group 所有数据，优先读取缓存，查询失败时返回空映射。
func (s *service) Group(group string) map[string]string {
	key := constant.CacheNameWithSysConfig + group
	if value := s.cache.Get(key); value != nil {
		return value.(map[string]string)
	}
	configs, err := s.GroupWithError(group)
	if err != nil {
		return make(map[string]string)
	}
	return configs
}

// GroupWithError 直接读取配置表，区分“没有配置”和“查询失败”，供业务初始化使用。
// 成功后刷新缓存，避免初始化依据陈旧的空缓存覆盖数据库中已有的设置。
func (s *service) GroupWithError(group string) (map[string]string, error) {
	configs, err := s.repositoryConfig.FindByGroup(group)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(configs))
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	s.cache.SetWithExpire(constant.CacheNameWithSysConfig+group, result, constant.CacheTimeWithSysConfig)
	return result, nil
}

// Set 设置数据
func (s *service) Set(key string, value string) error {
	return s.Sets(map[string]string{key: value})
}

// Sets 批量设置数据
func (s *service) Sets(configs map[string]string) error {
	if len(configs) == 0 {
		return errors.New("设置数据不能为空")
	}
	if !s.cache.Lock(constant.LockConfigSet, constant.LockTimeConfigSet) {
		return errors.New("获取并发锁失败")
	}
	defer s.cache.UnLock(constant.LockConfigSet)
	list := make([]*model.Config, 0, len(configs))
	groups := make(map[string]struct{})
	for key, value := range configs {
		group := key
		if index := strings.IndexByte(key, '.'); index > 0 {
			group = key[:index]
		}
		groups[group] = struct{}{}
		list = append(list, &model.Config{
			Key:   key,
			Value: value,
		})
	}
	defer func() {
		for group := range groups {
			s.cache.Delete(constant.CacheNameWithSysConfig + group)
		}
	}()
	return s.repositoryConfig.Save(list)
}

// Get 获取数据
func (s *service) Get(group string, key string) string {
	temp := s.Group(group)
	return temp[key]
}

// DeleteByGroup 删除整组配置并清理缓存，供业务卸载时回收自有配置。
func (s *service) DeleteByGroup(group string) error {
	if strings.TrimSpace(group) == "" {
		return errors.New("配置组不能为空")
	}
	if !s.cache.Lock(constant.LockConfigSet, constant.LockTimeConfigSet) {
		return errors.New("获取并发锁失败")
	}
	defer s.cache.UnLock(constant.LockConfigSet)
	defer s.cache.Delete(constant.CacheNameWithSysConfig + group)
	return s.repositoryConfig.DeleteByGroup(group)
}
