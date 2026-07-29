package file

import (
	"context"
	"server/app/enum/file_format"
	"server/app/util/archive"
)

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
