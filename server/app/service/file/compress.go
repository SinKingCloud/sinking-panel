package file

import (
	"context"
	"server/app/util/archive"
)

// CompressFiles 压缩文件
// ctx: 上下文，用于取消操作
// srcPaths: 要压缩的源文件/目录路径数组
// destPath: 目标压缩文件路径
// format: 压缩格式（zip, tar, gz, tgz）
// callback: 进度回调函数，参数为：
//   - current: 已处理字节数
//   - total: 总字节数
//   - currentFile: 当前正在处理的文件名
//   - totalFiles: 总文件数量
//   - currentIndex: 当前文件索引（从0开始）
//     返回false表示取消操作
func (s *Service) CompressFiles(ctx context.Context, srcPaths []string, destPath string, format string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 创建一个包装回调函数，检查上下文是否已取消
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

	// 使用包装后的回调函数
	err := archive.Compress.CompressFiles(srcPaths, destPath, format, wrappedCallback)

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return err
}

// Extract 解压文件
// ctx: 上下文，用于取消操作
// srcPath: 要解压的源文件路径
// destPath: 解压目标目录
// callback: 进度回调函数，参数为：
//   - current: 已处理字节数
//   - total: 总字节数
//   - currentFile: 当前正在处理的文件名
//   - totalFiles: 总文件数量
//   - currentIndex: 当前文件索引（从0开始）
//     返回false表示取消操作
func (s *Service) Extract(ctx context.Context, srcPath string, destPath string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 创建一个包装回调函数，检查上下文是否已取消
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

	// 使用包装后的回调函数
	err := archive.Compress.Extract(srcPath, destPath, wrappedCallback)

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return err
}

// GetFormatByExt 根据文件扩展名获取压缩格式
func (s *Service) GetFormatByExt(filename string) string {
	return archive.Compress.GetFormatByExt(filename)
}

// IsSupportedFormat 判断是否为支持的压缩格式
func (s *Service) IsSupportedFormat(format string) bool {
	return archive.Compress.IsSupportedFormat(format)
}

// Formats 获取支持的压缩格式常量
func (s *Service) Formats() map[string]string {
	return map[string]string{
		"zip": archive.FormatZip,
		"tar": archive.FormatTar,
		"gz":  archive.FormatGzip,
		"tgz": archive.FormatTgz,
	}
}
