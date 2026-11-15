package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (s *Manager) parsePathOptions(content json.RawMessage) (PathOptions, error) {
	var options *PathOptions
	if err := json.Unmarshal(content, &options); err != nil || options == nil {
		return PathOptions{}, errors.New("路径备份参数必须是 JSON 对象")
	}
	return *options, nil
}

// ValidatePaths 校验并规范化来源，不创建备份目录。
func (s *Manager) ValidatePaths(paths []string) ([]string, error) {
	if len(paths) < 1 || len(paths) > 1000 {
		return nil, errors.New("备份路径数量须为 1 到 1000")
	}
	result := make([]string, 0, len(paths))
	seen := make(map[string]bool, len(paths))
	for _, path := range paths {
		path, err := s.validateSourcePath(path)
		if err != nil {
			return nil, err
		}
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	// 已选择父目录时不重复打包其中的文件或子目录。
	filtered := result[:0]
	for _, path := range result {
		covered := false
		for parent := filepath.Dir(path); parent != path; parent = filepath.Dir(parent) {
			if seen[parent] {
				covered = true
				break
			}
			if filepath.Dir(parent) == parent {
				break
			}
		}
		if !covered {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
}

func (s *Manager) validateSourcePath(path string) (string, error) {
	if !filepath.IsAbs(path) || strings.ContainsRune(path, '\x00') {
		return "", fmt.Errorf("备份来源须为绝对路径: %q", path)
	}
	path = filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("解析备份来源 %q 失败: %w", path, err)
	}
	path = resolved
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("读取备份来源 %q 失败: %w", path, err)
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return "", fmt.Errorf("备份来源须为普通文件或目录: %q", path)
	}
	output, err := s.resolveBackupRoot()
	if err != nil {
		return "", err
	}
	for _, pair := range [][2]string{{output, path}, {path, output}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
			return "", errors.New("备份来源与备份目录不能相同或互相包含")
		}
	}
	return filepath.Clean(path), nil
}

// 输出目录可能尚未创建，解析已有父目录以识别符号链接别名。
func (s *Manager) resolveBackupRoot() (string, error) {
	for parent := s.root; ; parent = filepath.Dir(parent) {
		resolved, err := filepath.EvalSymlinks(parent)
		if err == nil {
			relative, err := filepath.Rel(parent, s.root)
			if err != nil {
				return "", err
			}
			return filepath.Join(resolved, relative), nil
		}
		if !errors.Is(err, os.ErrNotExist) || filepath.Dir(parent) == parent {
			return "", fmt.Errorf("解析备份目录失败: %w", err)
		}
	}
}

// RunPaths 将自定义的文件和目录合并归档，每次执行生成一个副本。
func (s *Manager) RunPaths(ctx context.Context, id int64, paths []string, config Config, writeLog func(string)) (err error) {
	if writeLog == nil {
		writeLog = func(string) {}
	}
	defer func() {
		if err != nil {
			writeLog("自定义备份失败: " + err.Error())
		}
	}()
	if id <= 0 {
		return errors.New("备份计划 ID 不合法")
	}
	if _, running := s.pathRuns.LoadOrStore(id, struct{}{}); running {
		return errors.New("该备份计划正在执行")
	}
	defer s.pathRuns.Delete(id)
	if err = ctx.Err(); err != nil {
		return err
	}
	paths, err = s.ValidatePaths(paths)
	if err != nil {
		return err
	}
	writeLog(fmt.Sprintf("开始自定义备份，共 %d 个来源", len(paths)))
	file, err := s.Save(ctx, Target{
		Key:         "file",
		PlanId:      id,
		Extension:   ".tar.gz",
		ContentType: "application/gzip",
	}, config, func(ctx context.Context, output io.Writer) error {
		return s.packPathsArchive(ctx, paths, output)
	})
	if err != nil {
		return err
	}
	writeLog(fmt.Sprintf("备份完成，文件: %s，大小: %d 字节", file.Name, file.Size))
	if file.Warning != "" {
		writeLog(file.Warning)
	}
	return nil
}
