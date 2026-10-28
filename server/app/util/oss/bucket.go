package oss

import "context"

// ListBuckets 获取账户下的全部存储桶。
func (m *Manager) ListBuckets(ctx context.Context) ([]Bucket, error) {
	if err := m.ready(ctx, false); err != nil {
		return nil, err
	}
	return m.driver.ListBuckets(ctx)
}

// HeadBucket 检查绑定的存储桶是否可访问。
func (m *Manager) HeadBucket(ctx context.Context) error {
	if err := m.ready(ctx, true); err != nil {
		return err
	}
	return m.driver.HeadBucket(ctx)
}

// CreateBucket 在配置区域创建绑定的存储桶。
func (m *Manager) CreateBucket(ctx context.Context) error {
	if err := m.ready(ctx, true); err != nil {
		return err
	}
	return m.driver.CreateBucket(ctx)
}

// DeleteBucket 仅删除空桶，非空桶由厂商拒绝；不会递归删除对象。
func (m *Manager) DeleteBucket(ctx context.Context) error {
	if err := m.ready(ctx, true); err != nil {
		return err
	}
	return m.driver.DeleteBucket(ctx)
}
