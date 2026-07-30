package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 定义支持的压缩格式
const (
	FormatZip  = "zip"
	FormatTar  = "tar"
	FormatGzip = "gz"
	FormatTgz  = "tgz"
)

// Service 实现压缩与解压缩功能
type Service struct{}

// New 创建压缩服务实例
func New() *Service {
	return &Service{}
}

// Compress 全局单例
var Compress = New()

// GetFormatByExt 根据文件扩展名获取压缩格式
func (s *Service) GetFormatByExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return ""
	}
	ext = ext[1:] // 移除点号
	switch ext {
	case "zip":
		return FormatZip
	case "tar":
		return FormatTar
	case "gz", "gzip":
		if strings.HasSuffix(strings.ToLower(filename), ".tar.gz") {
			return FormatTgz
		}
		return FormatGzip
	case "tgz":
		return FormatTgz
	default:
		return ""
	}
}

// IsSupportedFormat 判断是否为支持的压缩格式
func (s *Service) IsSupportedFormat(format string) bool {
	switch format {
	case FormatZip, FormatTar, FormatGzip, FormatTgz:
		return true
	default:
		return false
	}
}

// CompressFiles 压缩文件或目录
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
func (s *Service) CompressFiles(ctx context.Context, srcPaths []string, destPath, format string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	if !s.IsSupportedFormat(format) {
		return errors.New("不支持的压缩格式")
	}
	if err := s.validateCompressDestination(srcPaths, destPath); err != nil {
		return err
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 再次检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var err error
	defer func() {
		// 如果发生错误，清理可能部分创建的目标文件
		if err != nil {
			_ = os.Remove(destPath)
		}
	}()

	switch format {
	case FormatZip:
		err = s.compressZip(ctx, srcPaths, destPath, callback)
	case FormatTar:
		err = s.compressTar(ctx, srcPaths, destPath, false, callback)
	case FormatGzip:
		if len(srcPaths) != 1 {
			err = errors.New("gzip格式只支持压缩单个文件")
			return err
		}
		err = s.compressGzip(ctx, srcPaths[0], destPath)
	case FormatTgz:
		err = s.compressTar(ctx, srcPaths, destPath, true, callback)
	default:
		err = errors.New("不支持的压缩格式")
	}

	return err
}

// Extract 解压缩文件
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
func (s *Service) Extract(ctx context.Context, srcPath, destPath string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	format := s.GetFormatByExt(srcPath)
	if !s.IsSupportedFormat(format) {
		return errors.New("不支持的压缩格式")
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 确保目标目录存在
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 再次检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var err error

	switch format {
	case FormatZip:
		err = s.extractZip(ctx, srcPath, destPath, callback)
	case FormatTar:
		err = s.extractTar(ctx, srcPath, destPath, false, callback)
	case FormatGzip:
		if strings.HasSuffix(strings.ToLower(srcPath), ".tar.gz") {
			err = s.extractTar(ctx, srcPath, destPath, true, callback)
		} else {
			err = s.extractGzip(ctx, srcPath, destPath)
		}
	case FormatTgz:
		err = s.extractTar(ctx, srcPath, destPath, true, callback)
	default:
		err = errors.New("不支持的压缩格式")
	}

	// 如果解压失败且不是因为上下文取消，清理可能创建的部分文件
	if err != nil && !errors.Is(err, ctx.Err()) {
		// 解压失败但非上下文取消错误，尝试清理部分创建的文件
		// 注意：我们不删除整个目录，因为目标目录可能包含其他文件
		_ = os.Remove(destPath)
	}

	return err
}

// 获取指定目录下所有文件数量（用于进度计算）
func (s *Service) countFiles(paths []string) (int, error) {
	count := 0
	for _, path := range paths {
		err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			count++
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// ---------------------- ZIP 压缩与解压缩 ----------------------

// compressZip 将文件/目录压缩为ZIP文件
func (s *Service) compressZip(ctx context.Context, srcPaths []string, destPath string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 获取文件总数用于进度计算
	totalFiles, err := s.countFiles(srcPaths)
	if err != nil {
		return err
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 计算总字节数
	var totalSize int64
	for _, path := range srcPaths {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			// 每个文件检查一次上下文状态
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 创建ZIP文件
	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var processedSize int64
	fileIndex := int64(0)
	for _, srcPath := range srcPaths {
		srcPath = filepath.Clean(srcPath)
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 遍历源路径中的所有文件和目录
		err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			// 每个文件检查一次上下文状态
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			// 更新进度
			if callback != nil {
				if !callback(processedSize, totalSize, path, int64(totalFiles), fileIndex) {
					return errors.New("操作被取消")
				}
			}

			// 计算在ZIP中的相对路径
			relPath, err := filepath.Rel(filepath.Dir(srcPath), path)
			if err != nil {
				return err
			}

			// 统一使用斜杠作为路径分隔符
			relPath = filepath.ToSlash(relPath)

			fileHeader, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			fileHeader.Name = relPath

			// 如果是目录，添加目录项
			if info.IsDir() {
				fileHeader.Name += "/"
				fileHeader.Method = zip.Store
				_, err = zipWriter.CreateHeader(fileHeader)
				fileIndex++
				return err
			}

			// 创建文件项
			fileHeader.Method = zip.Deflate // 使用压缩
			if info.Mode()&os.ModeSymlink != 0 {
				fileHeader.Method = zip.Store
			}

			writer, err := zipWriter.CreateHeader(fileHeader)
			if err != nil {
				return err
			}

			if info.Mode()&os.ModeSymlink != 0 {
				linkTarget, err := os.Readlink(path)
				if err != nil {
					return err
				}
				written, err := io.WriteString(writer, filepath.ToSlash(linkTarget))
				if err != nil {
					return err
				}
				processedSize += int64(written)
				fileIndex++
				return nil
			}

			// 打开源文件
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// 检查上下文是否已取消
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// 复制文件内容
			written, err := io.Copy(writer, file)
			if err != nil {
				return err
			}

			processedSize += written
			fileIndex++

			// 最终更新进度
			if callback != nil {
				if !callback(processedSize, totalSize, path, int64(totalFiles), fileIndex-1) {
					return errors.New("操作被取消")
				}
			}

			// 再次检查上下文是否已取消
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// extractZip 解压ZIP文件到指定目录
func (s *Service) extractZip(ctx context.Context, srcPath, destPath string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}
	rootPath, err := s.resolvePath(destPath)
	if err != nil {
		return err
	}

	// 打开ZIP文件
	reader, err := zip.OpenReader(srcPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 获取文件总数
	totalFiles := int64(len(reader.File))

	// 计算总字节数
	var totalSize int64
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			totalSize += int64(file.UncompressedSize64)
		}
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 遍历ZIP文件中的所有文件
	var processedSize int64
	for i, file := range reader.File {
		// 每个文件处理前检查上下文状态
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 更新进度
		if callback != nil {
			if !callback(processedSize, totalSize, file.Name, totalFiles, int64(i)) {
				return errors.New("操作被取消")
			}
		}

		// 构建解压后的文件路径
		destFilePath, err := s.getExtractPath(rootPath, file.Name)
		if err != nil {
			return err
		}

		// 如果是目录，创建目录
		if file.FileInfo().IsDir() {
			if destFilePath == rootPath {
				continue
			}
			if info, err := os.Lstat(destFilePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
				if err = os.Remove(destFilePath); err != nil {
					return err
				}
			} else if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err := os.MkdirAll(destFilePath, file.Mode()); err != nil {
				return err
			}
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			srcFile, err := file.Open()
			if err != nil {
				return err
			}
			linkTarget, err := io.ReadAll(io.LimitReader(srcFile, 4097))
			srcFile.Close()
			if err != nil {
				return err
			}
			if len(linkTarget) > 4096 {
				return fmt.Errorf("符号链接目标过长: %s", file.Name)
			}
			if err = s.createExtractSymlink(rootPath, destFilePath, string(linkTarget)); err != nil {
				return err
			}
			processedSize += int64(len(linkTarget))
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
			return err
		}
		if info, err := os.Lstat(destFilePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
			if err = os.Remove(destFilePath); err != nil {
				return err
			}
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}

		// 创建文件
		destFile, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		// 打开ZIP中的文件
		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		// 检查上下文是否已取消
		if ctx.Err() != nil {
			srcFile.Close()
			destFile.Close()
			return ctx.Err()
		}

		// 复制文件内容
		written, err := io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return err
		}

		processedSize += written

		// 最终更新进度
		if callback != nil {
			if !callback(processedSize, totalSize, file.Name, totalFiles, int64(i)) {
				return errors.New("操作被取消")
			}
		}

		// 再次检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return nil
}

// ---------------------- TAR 压缩与解压缩 ----------------------

// compressTar 将文件/目录压缩为TAR文件
// 如果useGzip为true，则使用gzip压缩TAR文件
func (s *Service) compressTar(ctx context.Context, srcPaths []string, destPath string, useGzip bool, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 获取文件总数用于进度计算
	totalFiles, err := s.countFiles(srcPaths)
	if err != nil {
		return err
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 计算总字节数
	var totalSize int64
	for _, path := range srcPaths {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			// 每个文件检查一次上下文状态
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 创建TAR文件
	tarFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	var tarWriter *tar.Writer
	var gzipWriter *gzip.Writer

	if useGzip {
		// 使用gzip压缩
		gzipWriter = gzip.NewWriter(tarFile)
		defer gzipWriter.Close()
		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		// 不使用gzip压缩
		tarWriter = tar.NewWriter(tarFile)
	}
	defer tarWriter.Close()

	var processedSize int64
	fileIndex := int64(0)
	for _, srcPath := range srcPaths {
		srcPath = filepath.Clean(srcPath)
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 遍历源路径中的所有文件和目录
		err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			// 每个文件检查一次上下文状态
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			// 更新进度
			if callback != nil {
				if !callback(processedSize, totalSize, path, int64(totalFiles), fileIndex) {
					return errors.New("操作被取消")
				}
			}

			// 计算在TAR中的相对路径
			relPath, err := filepath.Rel(filepath.Dir(srcPath), path)
			if err != nil {
				return err
			}

			// 创建TAR头部
			linkTarget := ""
			if info.Mode()&os.ModeSymlink != 0 {
				linkTarget, err = os.Readlink(path)
				if err != nil {
					return err
				}
			}
			header, err := tar.FileInfoHeader(info, filepath.ToSlash(linkTarget))
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			// 写入TAR头部
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			// 如果是常规文件，写入文件内容
			if info.Mode().IsRegular() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				// 检查上下文是否已取消
				if ctx.Err() != nil {
					return ctx.Err()
				}

				written, err := io.Copy(tarWriter, file)
				if err != nil {
					return err
				}

				processedSize += written
			}

			fileIndex++

			// 最终更新进度
			if callback != nil && info.Mode().IsRegular() {
				if !callback(processedSize, totalSize, path, int64(totalFiles), fileIndex-1) {
					return errors.New("操作被取消")
				}
			}

			// 再次检查上下文是否已取消
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTar 解压TAR文件到指定目录
// 如果isGzipped为true，则处理.tar.gz文件
func (s *Service) extractTar(ctx context.Context, srcPath, destPath string, isGzipped bool, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}
	rootPath, err := s.resolvePath(destPath)
	if err != nil {
		return err
	}

	// 打开TAR文件
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取源文件大小作为总大小的估计值
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	totalSize := fileInfo.Size()

	var tarReader *tar.Reader
	if isGzipped {
		// 使用gzip解压
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		tarReader = tar.NewReader(gzipReader)
	} else {
		// 不使用gzip解压
		tarReader = tar.NewReader(file)
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 首先计算文件总数
	fileCount := 0
	var headers []*tar.Header
	for {
		// 每次迭代检查上下文状态
		if ctx.Err() != nil {
			return ctx.Err()
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		headers = append(headers, header)
		fileCount++
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 重新打开文件
	file.Close()
	file, err = os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if isGzipped {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		tarReader = tar.NewReader(gzipReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	// 解压文件
	var processedSize int64
	for i := 0; i < fileCount; i++ {
		// 每个文件处理前检查上下文状态
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 读取TAR头部
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 更新进度
		if callback != nil {
			if !callback(processedSize, totalSize, header.Name, int64(fileCount), int64(i)) {
				return errors.New("操作被取消")
			}
		}

		// 构建解压后的文件路径
		destFilePath, err := s.getExtractPath(rootPath, header.Name)
		if err != nil {
			return err
		}

		// 根据头部类型处理
		switch header.Typeflag {
		case tar.TypeDir:
			if destFilePath == rootPath {
				continue
			}
			if info, err := os.Lstat(destFilePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
				if err = os.Remove(destFilePath); err != nil {
					return err
				}
			} else if err != nil && !os.IsNotExist(err) {
				return err
			}
			// 创建目录
			if err := os.MkdirAll(destFilePath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
				return err
			}
			if info, err := os.Lstat(destFilePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
				if err = os.Remove(destFilePath); err != nil {
					return err
				}
			} else if err != nil && !os.IsNotExist(err) {
				return err
			}

			// 创建文件
			file, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// 检查上下文是否已取消
			if ctx.Err() != nil {
				file.Close()
				return ctx.Err()
			}

			// 复制文件内容
			written, err := io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return err
			}
			processedSize += written
		case tar.TypeSymlink:
			if err = s.createExtractSymlink(rootPath, destFilePath, header.Linkname); err != nil {
				return err
			}
		}

		// 最终更新进度
		if callback != nil {
			if !callback(processedSize, totalSize, header.Name, int64(fileCount), int64(i)) {
				return errors.New("操作被取消")
			}
		}
	}

	return nil
}

// ---------------------- GZIP 压缩与解压缩 ----------------------

func (s *Service) getExtractPath(rootPath string, name string) (string, error) {
	name = filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("非法的文件路径: %s", name)
	}
	destPath := filepath.Join(rootPath, name)
	resolvedPath, err := s.resolvePath(destPath)
	if err != nil {
		return "", err
	}
	relPath, err := filepath.Rel(rootPath, resolvedPath)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relPath) {
		return "", fmt.Errorf("非法的文件路径: %s", name)
	}
	return destPath, nil
}

func (s *Service) createExtractSymlink(rootPath string, destPath string, linkTarget string) error {
	if destPath == rootPath || linkTarget == "" {
		return errors.New("符号链接路径不合法")
	}
	linkTarget = filepath.FromSlash(linkTarget)
	if filepath.IsAbs(linkTarget) || filepath.VolumeName(linkTarget) != "" {
		return fmt.Errorf("非法的符号链接目标: %s", linkTarget)
	}
	resolvedTarget, err := s.resolvePath(filepath.Join(filepath.Dir(destPath), linkTarget))
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(rootPath, resolvedTarget)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relPath) {
		return fmt.Errorf("非法的符号链接目标: %s", linkTarget)
	}
	if err = os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	if _, err = os.Lstat(destPath); err == nil {
		if err = os.Remove(destPath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(linkTarget, destPath)
}

func (s *Service) validateCompressDestination(srcPaths []string, dest string) error {
	destPath, err := s.resolvePath(dest)
	if err != nil {
		return fmt.Errorf("解析目标路径失败: %w", err)
	}
	for _, src := range srcPaths {
		srcPath, err := s.resolvePath(src)
		if err != nil {
			return fmt.Errorf("解析源路径失败: %w", err)
		}
		relPath, err := filepath.Rel(srcPath, destPath)
		if err != nil {
			continue
		}
		if relPath == "." {
			return errors.New("压缩目标不能与源路径相同")
		}
		info, err := os.Stat(src)
		if err != nil {
			return err
		}
		if info.IsDir() && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relPath) {
			return errors.New("压缩目标不能位于源目录内")
		}
	}
	return nil
}

func (s *Service) resolvePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolvedPath, err := filepath.EvalSymlinks(absPath); err == nil {
		return filepath.Clean(resolvedPath), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	parent := filepath.Dir(absPath)
	for {
		resolvedParent, err := filepath.EvalSymlinks(parent)
		if err == nil {
			relPath, err := filepath.Rel(parent, absPath)
			if err != nil {
				return "", err
			}
			return filepath.Clean(filepath.Join(resolvedParent, relPath)), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		next := filepath.Dir(parent)
		if next == parent {
			return filepath.Clean(absPath), nil
		}
		parent = next
	}
}

// compressGzip 使用gzip压缩单个文件
// ctx: 上下文，用于取消操作
func (s *Service) compressGzip(ctx context.Context, srcPath string, destPath string) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 检查源文件是否是目录
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return errors.New("不能直接用gzip压缩目录，请使用tgz格式")
	}

	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}

	// 使用defer处理清理逻辑
	success := false
	defer func() {
		_ = destFile.Close()

		// 如果操作未成功完成，删除部分创建的文件
		if !success {
			_ = os.Remove(destPath)
		}
	}()

	// 创建gzip写入器
	gzipWriter := gzip.NewWriter(destFile)
	defer func() {
		_ = gzipWriter.Close()
	}()

	// 设置文件名
	gzipWriter.Name = filepath.Base(srcPath)

	// 使用缓冲区进行复制，并定期检查上下文状态
	buf := make([]byte, 32*1024) // 32KB 缓冲区
	for {
		// 每次读取前检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, readErr := srcFile.Read(buf)
		if n > 0 {
			// 写入前再次检查上下文
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if _, writeErr := gzipWriter.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	// 标记操作成功完成
	success = true
	return nil
}

// extractGzip 解压GZIP文件到指定目录
// ctx: 上下文，用于取消操作
func (s *Service) extractGzip(ctx context.Context, srcPath string, destPath string) error {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	// 创建gzip读取器
	gzipReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = gzipReader.Close()
	}()

	// 确定目标文件名
	var destFilePath string
	if gzipReader.Name != "" {
		destFilePath = filepath.Join(destPath, gzipReader.Name)
	} else {
		// 如果gzip没有保存原始文件名，使用源文件名去掉.gz后缀
		baseName := filepath.Base(srcPath)
		if strings.HasSuffix(strings.ToLower(baseName), ".gz") {
			baseName = baseName[:len(baseName)-3]
		}
		destFilePath = filepath.Join(destPath, baseName)
	}

	// 检查文件路径是否在目标路径内（防止路径穿越攻击）
	if !strings.HasPrefix(destFilePath, filepath.Clean(destPath)+string(os.PathSeparator)) {
		return fmt.Errorf("非法的文件路径: %s", gzipReader.Name)
	}

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
		return err
	}

	// 创建目标文件
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return err
	}

	// 使用defer处理清理逻辑
	success := false
	defer func() {
		_ = destFile.Close()

		// 如果操作未成功完成，删除部分创建的文件
		if !success {
			_ = os.Remove(destFilePath)
		}
	}()

	// 使用缓冲区进行复制，并定期检查上下文状态
	buf := make([]byte, 32*1024) // 32KB 缓冲区
	for {
		// 每次读取前检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, readErr := gzipReader.Read(buf)
		if n > 0 {
			// 写入前再次检查上下文
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if _, writeErr := destFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	// 标记操作成功完成
	success = true
	return nil
}
