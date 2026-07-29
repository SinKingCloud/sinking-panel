package config

import (
	"errors"
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
	temp := s.selectByGroup(group)
	s.cache.SetWithExpire(key, temp, constant.CacheTimeWithSysConfig)
	return temp
}

// Set 设置数据
func (s *service) Set(group string, key string, value string) error {
	return s.Sets(group, map[string]string{key: value})
}

// Sets 批量设置数据
func (s *service) Sets(group string, configs map[string]string) error {
	if len(configs) == 0 {
		return errors.New("设置数据不能为空")
	}
	lock := constant.LockConfigSet + group
	if !s.cache.Lock(lock, constant.LockTimeConfigSet) {
		return errors.New("获取并发锁失败")
	}
	defer s.cache.UnLock(lock)
	defer s.cache.Delete(constant.CacheNameWithSysConfig + group)
	list := make([]*model.Config, 0, len(configs))
	for key, value := range configs {
		list = append(list, &model.Config{
			Key:   key,
			Value: value,
		})
	}
	return s.repositoryConfig.Save(list)
}

// Get 获取数据
func (s *service) Get(group string, key string) string {
	temp := s.Group(group)
	return temp[key]
}
