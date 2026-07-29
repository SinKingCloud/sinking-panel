package config

import (
	"errors"
	"server/app/constant"
	"server/app/model"
	"server/app/util/str"
	"time"
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
	lock := constant.LockConfigSet + group + key
	if !s.cache.Lock(lock, constant.LockTimeConfigSet) {
		return errors.New("获取并发锁失败")
	}
	defer s.cache.UnLock(lock)
	defer s.cache.Delete(constant.CacheNameWithSysConfig + group)
	if n, e := s.repositoryConfig.CountByKey(key); e == nil && n > 0 {
		return s.repositoryConfig.UpdateByKey(key, value)
	}
	return s.repositoryConfig.Create(&model.Config{
		Key:        key,
		Value:      value,
		CreateTime: str.DateTime(time.Now()),
	})
}

// Get 获取数据
func (s *service) Get(group string, key string) string {
	temp := s.Group(group)
	return temp[key]
}
