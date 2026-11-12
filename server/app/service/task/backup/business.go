package backup

import (
	"context"
	"errors"
	"io"
	"server/app/util/page"
)

// Select 分页返回对象在指定存储中的备份，包含该对象所有计划的副本。
func (s *Manager) Select(ctx context.Context, target Target, config Config, pageNumber, pageSize int) (*page.Result[*File], error) {
	if pageNumber < 1 || pageNumber > 1000000 || pageSize < 1 || pageSize > 100 {
		return nil, errors.New("备份分页参数不合法")
	}
	target, err := s.validateTarget(target)
	if err != nil {
		return nil, err
	}
	config, err = s.validateStorage(config)
	if err != nil {
		return nil, err
	}
	files, err := s.selectBackupFiles(ctx, target, config)
	if err != nil {
		return nil, err
	}
	total := int64(len(files))
	start := min(int64(pageNumber-1)*int64(pageSize), total)
	end := min(start+int64(pageSize), total)
	return &page.Result[*File]{
		List:     files[start:end],
		Total:    &total,
		Page:     &pageNumber,
		PageSize: pageSize,
	}, nil
}

// Open 返回备份文件流，由调用者关闭及解释文件内容。
func (s *Manager) Open(ctx context.Context, target Target, config Config, name string) (io.ReadCloser, error) {
	target, err := s.validateTarget(target)
	if err != nil {
		return nil, err
	}
	config, err = s.validateStorage(config)
	if err != nil {
		return nil, err
	}
	return s.openBackupFile(ctx, target, config, name)
}

// Delete 删除指定备份，与同一执行器的文件发布及副本清理串行。
func (s *Manager) Delete(ctx context.Context, target Target, config Config, name string) error {
	target, err := s.validateTarget(target)
	if err != nil {
		return err
	}
	config, err = s.validateStorage(config)
	if err != nil {
		return err
	}
	select {
	case s.gate <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-s.gate }()
	return s.deleteBackupFile(ctx, target, config, name)
}
