package config

import (
	"errors"
	"strings"

	"server/app/constant"
	"server/app/model"
)

// Group 获取group所有数据
func (s *service) Group(group string) map[string]string {
	key := constant.CacheNameWithSysConfig + group
	value := s.cache.Get(key)
	if value != nil {
		return value.(map[string]string)
	}
	temp := make(map[string]string)
	configs, err := s.repositoryConfig.FindByGroup(group)
	if err != nil {
		return temp
	}
	for _, v := range configs {
		temp[v.Key] = v.Value
	}
	s.cache.SetWithExpire(key, temp, constant.CacheTimeWithSysConfig)
	return temp
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
