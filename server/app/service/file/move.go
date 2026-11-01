package file

// MoveTask 创建批量文件移动任务。
func (s *service) MoveTask(paths []string, target string) (string, error) {
	return s.transferTask(paths, target, true)
}
