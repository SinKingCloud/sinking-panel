package software

import (
	"server/app/model"
	"server/app/util/plugin"
)

// record 缓存安装记录，返回副本；失效和首次加载共用锁，避免旧查询覆盖新状态。
func (s *service) record(code string) (*model.Software, error) {
	s.cacheMu.RLock()
	record := s.records[code]
	if record != nil {
		snapshot := *record
		s.cacheMu.RUnlock()
		return &snapshot, nil
	}
	s.cacheMu.RUnlock()
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	record = s.records[code]
	if record == nil {
		var err error
		record, err = s.repository.FindByCode(code)
		if err != nil || record == nil {
			return nil, err
		}
		s.records[code] = record
	}
	snapshot := *record
	return &snapshot, nil
}

func (s *service) invalidate(code string) {
	s.cacheMu.Lock()
	delete(s.records, code)
	s.cacheMu.Unlock()
}

// manifest 保留提交安装时的依赖清单，排队阶段无需等待代码目录创建。
func (s *service) manifest(code string) (*plugin.Manifest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if manifest := s.manifests[code]; manifest != nil {
		return manifest, nil
	}
	manifest, err := s.plugins.Manifest(code)
	if err == nil {
		s.manifests[code] = manifest
	}
	return manifest, err
}
func (s *service) invalidateManifest(code string) {
	s.mu.Lock()
	delete(s.manifests, code)
	delete(s.icons, code)
	s.mu.Unlock()
}
