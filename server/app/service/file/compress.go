package file

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"server/app/enum/file_format"
	"server/app/util/archive"
	"server/app/util/file"
)

// CompressTask 校验源文件和目标目录，创建带进度的压缩任务。
func (s *service) CompressTask(paths []string, dir, name, format string) (string, error) {
	if !s.IsSupportedFormat(format) {
		return "", errors.New("不支持的压缩格式")
	}
	if len(paths) == 0 || dir == "" || name == "" {
		return "", errors.New("文件路径、目标目录和文件名不能为空")
	}
	f := file.NewDisk("")
	for _, path := range paths {
		if path == "" || !f.Exists(path) {
			return "", errors.New("源文件或目录不存在: " + path)
		}
	}
	if !f.Exists(dir) {
		if err := f.CreateDir(dir); err != nil {
			return "", fmt.Errorf("创建目标目录失败: %w", err)
		}
	}
	destPath := filepath.Join(dir, name+"."+format)
	if f.Exists(destPath) {
		return "", errors.New("目标文件已存在: " + destPath)
	}
	paths = append([]string(nil), paths...)
	data := map[string]interface{}{"paths": paths, "dest_path": destPath, "format": format}
	return s.createTask("压缩", "压缩文件: "+name+"."+format, data, func(ctx context.Context, update func(float64, string)) error {
		err := s.CompressFiles(ctx, paths, destPath, format, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if ctx.Err() != nil {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			update(progress, "正在压缩: "+currentFile)
			return true
		})
		if err != nil || ctx.Err() != nil {
			_ = f.Delete(destPath)
		}
		return err
	})
}

// ExtractTask 校验压缩文件并创建解压任务。
func (s *service) ExtractTask(path, dir string) (string, error) {
	f := file.NewDisk("")
	if path == "" || !f.Exists(path) {
		return "", errors.New("压缩文件不存在")
	}
	if f.IsDir(path) {
		return "", errors.New("不能解压目录，请选择压缩文件")
	}
	format := s.GetFormatByExt(path)
	if !s.IsSupportedFormat(format) {
		return "", errors.New("不支持的压缩格式或不是有效的压缩文件")
	}
	if dir == "" {
		return "", errors.New("解压目标目录不能为空")
	}
	if !f.Exists(dir) {
		if err := f.CreateDir(dir); err != nil {
			return "", fmt.Errorf("创建目标目录失败: %w", err)
		}
	}
	data := map[string]interface{}{"path": path, "dest_dir": dir, "format": format}
	return s.createTask("解压", "解压文件: "+path, data, func(ctx context.Context, update func(float64, string)) error {
		return s.Extract(ctx, path, dir, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if ctx.Err() != nil {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			update(progress, "正在解压: "+currentFile)
			return true
		})
	})
}

// CompressFiles 压缩文件
// ctx: 上下文，用于取消操作
// srcPaths: 源路径列表
// destPath: 目标路径
// format: 压缩格式
// callback: 进度回调函数，参数为：
//   - current: 已处理字节数
//   - total: 总字节数
//   - currentFile: 当前正在处理的文件名
//   - totalFiles: 总文件数量
//   - currentIndex: 当前文件索引（从0开始）
//     返回false表示取消操作
func (s *service) CompressFiles(ctx context.Context, srcPaths []string, destPath, format string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !s.IsSupportedFormat(format) {
		return errors.New("不支持的压缩格式")
	}

	// 创建包装的回调函数，检查上下文是否已取消
	wrappedCallback := func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			if callback != nil {
				return callback(current, total, currentFile, totalFiles, currentIndex)
			}
			return true
		}
	}

	// 使用带上下文的compress方法
	err := archive.Compress.CompressFiles(ctx, srcPaths, destPath, format, wrappedCallback)

	// 最终检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return err
	}
}

// Extract 解压文件
// ctx: 上下文，用于取消操作
// srcPath: 源路径
// destPath: 目标路径
// callback: 进度回调函数，参数为：
//   - current: 已处理字节数
//   - total: 总字节数
//   - currentFile: 当前正在处理的文件名
//   - totalFiles: 总文件数量
//   - currentIndex: 当前文件索引（从0开始）
//     返回false表示取消操作
func (s *service) Extract(ctx context.Context, srcPath, destPath string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 创建包装的回调函数，检查上下文是否已取消
	wrappedCallback := func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			if callback != nil {
				return callback(current, total, currentFile, totalFiles, currentIndex)
			}
			return true
		}
	}

	// 使用带上下文的解压方法
	err := archive.Compress.Extract(ctx, srcPath, destPath, wrappedCallback)

	// 最终检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return err
	}
}

// GetFormatByExt 根据文件扩展名获取压缩格式
func (s *service) GetFormatByExt(filename string) string {
	return archive.Compress.GetFormatByExt(filename)
}

// IsSupportedFormat 判断是否为支持的压缩格式
func (s *service) IsSupportedFormat(format string) bool {
	if _, ok := file_format.Map()[format]; !ok {
		return false
	}
	return archive.Compress.IsSupportedFormat(format)
}
