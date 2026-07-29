package config

// selectByGroup 查看配置数据
func (s *service) selectByGroup(group string) map[string]string {
	temp := make(map[string]string)
	configs, err := s.repositoryConfig.FindByGroup(group)
	if err != nil {
		return temp
	}
	for _, v := range configs {
		temp[v.Key] = v.Value
	}
	return temp
}
