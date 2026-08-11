package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Disk 封装文件系统操作，限定在指定根目录下进行
type Disk struct {
	root string // 经过标准化的基准根目录
}

// File 描述文件系统对象的元数据信息
type File struct {
	Name       string  `json:"name"`        // 文件/目录名称（不含路径）
	Size       int64   `json:"size"`        // 文件大小（字节），目录为0
	Mode       uint32  `json:"mode"`        // 权限模式（八进制表示，例如 0644）
	IsDir      bool    `json:"is_dir"`      // 是否为目录类型
	UpdateTime int64   `json:"update_time"` // 最后修改时间（Unix时间戳）
	Child      []*File `json:"child"`       // 子文件列表（仅当IsDir为true时有效）
}

const (
	bufferSize   = 32 << 10 // 32KB缓冲区，使用位运算优化
	progressUnit = 1 << 20  // 进度回调触发单位（1MB）
)

// NewDisk 创建新的Disk实例
// root: 基准根目录路径，所有操作将被限制在此目录下
func NewDisk(root string) *Disk {
	if root == "" {
		root = "./"
	}
	return &Disk{root: filepath.Clean(root)}
}

// Path 获取文件或目录路径
// name: 路径
func (d *Disk) Path(name string, hasFile bool, full bool) string {
	var fullPath string
	if full {
		fullPath, _ = filepath.Abs(d.fullPath(name))
	} else {
		fullPath = d.fullPath(name)
	}
	if hasFile {
		return fullPath
	}
	return filepath.Dir(fullPath)
}

// Rename 重命名文件或目录
// oldName: 原始相对路径
// newName: 新的相对路径
func (d *Disk) Rename(oldName, newName string) error {
	srcPath := d.fullPath(oldName)
	destPath := d.fullPath(newName)
	// 检查源路径是否存在
	if !d.Exists(oldName) {
		return fmt.Errorf("源路径不存在: %s", srcPath)
	}
	// 准备目标路径：如果目标已存在则删除，并确保目标父目录存在
	if err := d.prepareDestination(destPath); err != nil {
		return fmt.Errorf("准备目标路径失败: %w", err)
	}
	// 尝试直接重命名
	if err := os.Rename(srcPath, destPath); err == nil {
		return nil
	}
	// 如果 os.Rename 失败，则回退为复制和删除操作
	srcInfo, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		if err := d.copyTree(srcPath, destPath); err != nil {
			return fmt.Errorf("回退复制目录失败: %w", err)
		}
		return os.RemoveAll(srcPath)
	}

	if err := d.copyFile(srcPath, destPath); err != nil {
		return fmt.Errorf("回退复制文件失败: %w", err)
	}
	return os.Remove(srcPath)
}

// AutoCreate 智能创建文件或目录
// name: 相对路径，以路径分隔符结尾时自动创建目录
func (d *Disk) AutoCreate(name string) error {
	fullPath := d.fullPath(name)
	if d.isDirPath(name) {
		return os.MkdirAll(fullPath, 0755)
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return d.createFileIfNotExist(fullPath)
}

// Count 统计文件或目录信息
// name: 文件或目录
// totalSize: 共计大小
// fileCount: 文件数量
// err: 错误信息
func (d *Disk) Count(name string) (totalSize int64, fileCount int64, dirCount int64, err error) {
	if d.IsFile(name) {
		info, err := d.FileInfo(name)
		if err != nil {
			return 0, 0, 0, err
		}
		return info.Size, 1, 0, nil
	}
	totalSize, fileCount, dirCount, err = d.calculateTotal(d.fullPath(name))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("统计信息失败: %w", err)
	}
	return totalSize, fileCount, dirCount, err
}

// OpenFile 获取file对象
// name: 相对路径，以路径分隔符结尾时自动创建目录
func (d *Disk) OpenFile(name string, flag int, perm int) (*os.File, error) {
	return os.OpenFile(d.fullPath(name), flag, os.FileMode(perm))
}

// CreateFile 创建新文件（排他模式）
// fileName: 需要创建的相对路径，父目录必须已存在
func (d *Disk) CreateFile(fileName string) error {
	fullPath := d.fullPath(fileName)
	if !d.dirExists(filepath.Dir(fullPath)) {
		return errors.New("parent directory does not exist")
	}
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	return file.Close()
}

// CreateDir 递归创建目录结构
// dirName: 需要创建的目录相对路径
func (d *Disk) CreateDir(dirName string) error {
	return os.MkdirAll(d.fullPath(dirName), 0755)
}

// SetFileMode 设置文件/目录权限模式
// name: 存在的相对路径
// mode: Unix风格权限位（例如 0644）
func (d *Disk) SetFileMode(name string, mode uint32) error {
	return os.Chmod(d.fullPath(name), os.FileMode(mode))
}

// GetFileMode 获取当前权限模式
// 返回值：权限位模式，路径不存在时返回0
func (d *Disk) GetFileMode(name string) uint32 {
	info, err := os.Stat(d.fullPath(name))
	if err != nil {
		return 0
	}
	return uint32(info.Mode().Perm())
}

// GetFileContent 获取文件内容
// name: 存在的相对路径
// cursor: 本次读取的起始字节位置
// pageSize: 本次读取的行数
// version: 首次读取返回的文件版本
// 单次最多读取8MB，超长单行会通过游标继续读取
func (d *Disk) GetFileContent(name string, cursor int64, pageSize int, version string) (string, int64, bool, string, error) {
	if cursor < 0 || pageSize <= 0 {
		return "", 0, false, "", errors.New("参数不合法")
	}
	if cursor > 0 && version == "" {
		return "", 0, false, "", errors.New("缺少文件版本，请重新读取")
	}
	fullPath := d.fullPath(name)
	f, err := os.Open(fullPath)
	if err != nil {
		return "", 0, false, "", err
	}
	defer func() {
		_ = f.Close()
	}()
	unlock, err := d.lockFile(f, false)
	if err != nil {
		return "", 0, false, "", err
	}
	defer unlock()
	info, err := f.Stat()
	if err != nil {
		return "", 0, false, "", err
	}
	if !info.Mode().IsRegular() {
		return "", 0, false, "", errors.New("只能读取普通文件")
	}
	if cursor > info.Size() {
		return "", 0, false, "", errors.New("文件游标已失效，请重新读取")
	}
	if _, err = f.Seek(cursor, io.SeekStart); err != nil {
		return "", 0, false, "", err
	}
	currentVersion := fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano())
	if version != "" && version != currentVersion {
		return "", 0, false, "", errors.New("文件内容已发生变化，请重新读取")
	}
	contentSize := info.Size() - cursor
	if contentSize > bufferSize {
		contentSize = bufferSize
	}
	content := make([]byte, 0, int(contentSize))
	buffer := make([]byte, bufferSize)
	lineCount := 0
	stop := false
	for lineCount < pageSize && len(content) < 8<<20 {
		n, readErr := f.Read(buffer)
		if n > 0 {
			take := n
			if remaining := 8<<20 - len(content); take > remaining {
				take = remaining
				stop = true
			}
			for i, value := range buffer[:take] {
				if value == '\n' {
					lineCount++
					if lineCount == pageSize {
						take = i + 1
						stop = true
						break
					}
				}
			}
			content = append(content, buffer[:take]...)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", 0, false, "", readErr
		}
		if stop {
			break
		}
	}
	if !utf8.Valid(content) && len(content) > 0 {
		lastRune := len(content) - 1
		for lastRune > 0 && !utf8.RuneStart(content[lastRune]) {
			lastRune--
		}
		if !utf8.FullRune(content[lastRune:]) {
			content = content[:lastRune]
		}
	}
	if !utf8.Valid(content) {
		return "", 0, false, "", errors.New("文件不是UTF-8文本，无法读取内容")
	}
	nextCursor := cursor + int64(len(content))
	currentInfo, err := os.Stat(fullPath)
	if err != nil {
		return "", 0, false, "", err
	}
	if !os.SameFile(info, currentInfo) || fmt.Sprintf("%d:%d", currentInfo.Size(), currentInfo.ModTime().UnixNano()) != currentVersion {
		return "", 0, false, "", errors.New("文件内容已发生变化，请重新读取")
	}
	return string(content), nextCursor, nextCursor >= info.Size(), currentVersion, nil
}

// UpdateFileContent 更新文件内容
// cursor和nextCursor都为空时替换完整文件，否则替换指定游标区间
func (d *Disk) UpdateFileContent(name string, content string, cursor *int64, nextCursor *int64, version string) (int64, bool, string, error) {
	if (cursor == nil) != (nextCursor == nil) {
		return 0, false, "", errors.New("文件游标参数不完整")
	}
	partial := cursor != nil
	if partial && version == "" {
		return 0, false, "", errors.New("缺少文件版本，请重新读取")
	}
	fullPath := d.fullPath(name)
	if resolvedPath, err := filepath.EvalSymlinks(fullPath); err == nil {
		fullPath = resolvedPath
	} else {
		return 0, false, "", err
	}
	file, err := os.OpenFile(fullPath, os.O_RDWR, 0)
	if err != nil {
		return 0, false, "", err
	}
	defer func() {
		_ = file.Close()
	}()
	unlock, err := d.lockFile(file, true)
	if err != nil {
		return 0, false, "", err
	}
	defer unlock()
	info, err := file.Stat()
	if err != nil {
		return 0, false, "", err
	}
	if !info.Mode().IsRegular() {
		return 0, false, "", errors.New("只能更新普通文件")
	}
	currentVersion := fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano())
	if version != "" && version != currentVersion {
		return 0, false, "", errors.New("文件内容已发生变化，请重新读取后再保存")
	}
	if partial && (*cursor < 0 || *nextCursor < *cursor || *nextCursor > info.Size()) {
		return 0, false, "", errors.New("文件游标已失效，请重新读取")
	}
	backup, err := os.CreateTemp(filepath.Dir(fullPath), "."+filepath.Base(fullPath)+".edit-*")
	if err != nil {
		backup, err = os.CreateTemp("", "sinking-panel-edit-*")
		if err != nil {
			return 0, false, "", err
		}
	}
	backupPath := backup.Name()
	defer func() {
		_ = backup.Close()
		_ = os.Remove(backupPath)
	}()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return 0, false, "", err
	}
	if _, err = io.Copy(backup, file); err != nil {
		return 0, false, "", err
	}
	if err = backup.Sync(); err != nil {
		return 0, false, "", err
	}
	if _, err = backup.Seek(0, io.SeekStart); err != nil {
		return 0, false, "", err
	}
	currentInfo, err := file.Stat()
	if err != nil {
		return 0, false, "", err
	}
	if fmt.Sprintf("%d:%d", currentInfo.Size(), currentInfo.ModTime().UnixNano()) != currentVersion {
		return 0, false, "", errors.New("文件内容已发生变化，请重新读取后再保存")
	}
	pathInfo, err := os.Stat(fullPath)
	if err != nil {
		return 0, false, "", err
	}
	if !os.SameFile(info, pathInfo) || fmt.Sprintf("%d:%d", pathInfo.Size(), pathInfo.ModTime().UnixNano()) != currentVersion {
		return 0, false, "", errors.New("文件内容已发生变化，请重新读取后再保存")
	}
	writeErr := file.Truncate(0)
	modified := writeErr == nil
	if writeErr == nil {
		_, writeErr = file.Seek(0, io.SeekStart)
	}
	if writeErr == nil && partial && *cursor > 0 {
		_, writeErr = io.CopyN(file, backup, *cursor)
	}
	if writeErr == nil {
		_, writeErr = file.WriteString(content)
	}
	if writeErr == nil && partial {
		_, writeErr = backup.Seek(*nextCursor, io.SeekStart)
	}
	if writeErr == nil && partial {
		_, writeErr = io.Copy(file, backup)
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	if writeErr != nil && modified {
		restoreErr := file.Truncate(0)
		if restoreErr == nil {
			_, restoreErr = file.Seek(0, io.SeekStart)
		}
		if restoreErr == nil {
			_, restoreErr = backup.Seek(0, io.SeekStart)
		}
		if restoreErr == nil {
			_, restoreErr = io.Copy(file, backup)
		}
		if restoreErr == nil {
			restoreErr = file.Sync()
		}
		if restoreErr != nil {
			return 0, false, "", errors.Join(writeErr, fmt.Errorf("恢复原文件失败: %w", restoreErr))
		}
		return 0, false, "", writeErr
	}
	if writeErr != nil {
		return 0, false, "", writeErr
	}
	updatedInfo, err := file.Stat()
	if err != nil {
		return 0, false, "", err
	}
	pathInfo, err = os.Stat(fullPath)
	if err != nil {
		return 0, false, "", err
	}
	if !os.SameFile(updatedInfo, pathInfo) {
		return 0, false, "", errors.New("文件内容已发生变化，请重新读取后再保存")
	}
	newCursor := int64(len(content))
	eof := true
	if partial {
		newCursor += *cursor
		eof = *nextCursor == info.Size()
	}
	return newCursor, eof, fmt.Sprintf("%d:%d", pathInfo.Size(), pathInfo.ModTime().UnixNano()), nil
}

// Delete 递归删除文件或目录
// name: 需要删除的相对路径
func (d *Disk) Delete(name string) error {
	return os.RemoveAll(d.fullPath(name))
}

// Move 移动文件/目录到目标目录
// src: 需要移动的源路径（文件或目录）
// destDir: 目标目录的相对路径
func (d *Disk) Move(src, destDir string) error {
	srcPath := d.fullPath(src)
	destPath := d.fullPath(destDir)
	baseName := filepath.Base(srcPath)
	finalDest := filepath.Join(destPath, baseName)
	if err := d.validateTransferDestination(srcPath, finalDest); err != nil {
		return err
	}

	if err := d.prepareDestination(finalDest); err != nil {
		return fmt.Errorf("prepare destination failed: %w", err)
	}

	if err := os.Rename(srcPath, finalDest); err == nil {
		return nil
	}

	if err := d.copyTree(srcPath, finalDest); err != nil {
		return fmt.Errorf("cross-device copy failed: %w", err)
	}
	return os.RemoveAll(srcPath)
}

// Copy 复制文件/目录到目标目录
// src: 需要复制的源路径（文件或目录）
// destDir: 目标目录的相对路径
func (d *Disk) Copy(src, destDir string) error {
	srcPath := d.fullPath(src)
	destPath := d.fullPath(destDir)
	baseName := filepath.Base(srcPath)
	finalDest := filepath.Join(destPath, baseName)
	if err := d.validateTransferDestination(srcPath, finalDest); err != nil {
		return err
	}

	if err := d.prepareDestination(finalDest); err != nil {
		return fmt.Errorf("prepare destination failed: %w", err)
	}
	return d.copyTree(srcPath, finalDest)
}

// CopyWithProcess 带进度回调的文件/目录复制
// src: 源路径（文件或目录）
// destDir: 目标目录路径
// callback: 进度回调函数，参数分别为：
//
//	current - 已复制字节数
//	total - 总字节数
//	currentFile - 当前正在处理的文件名
//	totalFiles - 总文件数量
//	currentIndex - 当前文件序号（从0开始）
func (d *Disk) CopyWithProcess(ctx context.Context, src, destDir string, callback func(int64, int64, string, int64, int64) bool) error {
	srcPath := d.fullPath(src)
	destPath := d.fullPath(destDir)
	baseName := filepath.Base(srcPath)
	finalDest := filepath.Join(destPath, baseName)
	if err := d.validateTransferDestination(srcPath, finalDest); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	totalSize, fileCount, _, err := d.calculateTotalWithContext(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("calculate total failed: %w", err)
	}

	if err = ctx.Err(); err != nil {
		return err
	}

	if err := d.prepareDestination(finalDest); err != nil {
		return err
	}

	var copied atomic.Int64
	if err = d.copyTreeWithProgress(ctx, srcPath, finalDest, totalSize, fileCount, &copied, callback); err != nil {
		return cleanPartialDestination(finalDest, err)
	}
	return nil
}

// MoveWithProcess 带进度回调的文件/目录移动
// 参数说明同CopyWithProcess
func (d *Disk) MoveWithProcess(ctx context.Context, src, destDir string, callback func(int64, int64, string, int64, int64) bool) error {
	srcPath := d.fullPath(src)
	destPath := d.fullPath(destDir)
	baseName := filepath.Base(srcPath)
	finalDest := filepath.Join(destPath, baseName)
	if err := d.validateTransferDestination(srcPath, finalDest); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if callback != nil && !callback(0, 0, baseName, 1, 0) {
		return context.Canceled
	}
	complete := func() {
		if callback == nil {
			return
		}
		var size int64
		if info, err := os.Lstat(finalDest); err == nil {
			size = info.Size()
		}
		callback(size, size, baseName, 1, 0)
	}
	if err := d.prepareDestination(finalDest); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.Rename(srcPath, finalDest); err == nil {
		complete()
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	totalSize, fileCount, _, err := d.calculateTotalWithContext(ctx, srcPath)
	if err != nil {
		return cleanPartialDestination(finalDest, fmt.Errorf("calculate total failed: %w", err))
	}

	var copied atomic.Int64
	if err = d.copyTreeWithProgress(ctx, srcPath, finalDest, totalSize, fileCount, &copied, callback); err != nil {
		return cleanPartialDestination(finalDest, fmt.Errorf("copy failed: %w", err))
	}
	if err = ctx.Err(); err != nil {
		return cleanPartialDestination(finalDest, err)
	}
	if err = os.RemoveAll(srcPath); err != nil {
		return cleanPartialDestination(finalDest, fmt.Errorf("remove source failed: %w", err))
	}
	return nil
}

// IsFile 检查路径是否为文件
// 返回值：true表示是文件，false表示不存在或为目录
func (d *Disk) IsFile(name string) bool {
	info, err := os.Stat(d.fullPath(name))
	return err == nil && !info.IsDir()
}

// IsDir 检查路径是否为目录
// 返回值：true表示是目录，false表示不存在或为文件
func (d *Disk) IsDir(name string) bool {
	info, err := os.Stat(d.fullPath(name))
	return err == nil && info.IsDir()
}

// Exists 检查路径是否存在
// 返回值：true表示存在，false表示不存在
func (d *Disk) Exists(name string) bool {
	_, err := os.Lstat(d.fullPath(name))
	return !os.IsNotExist(err)
}

// FileInfo 获取文件/目录元信息
// 返回值：File结构指针和可能的错误信息
func (d *Disk) FileInfo(name string) (*File, error) {
	fullPath := d.fullPath(name)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	return &File{
		Name:       filepath.Base(fullPath),
		Size:       info.Size(),
		Mode:       uint32(info.Mode().Perm()),
		IsDir:      info.IsDir(),
		UpdateTime: info.ModTime().Unix(),
	}, nil
}

// FileList 递归获取目录结构信息
// dir: 需要查看的目录相对路径
// 返回值：目录结构的File切片
func (d *Disk) FileList(dir string) ([]*File, error) {
	return d.buildFileTree(d.fullPath(dir), false)
}

// FileListWithPage 分页获取指定目录下的文件或目录列表（非递归），
// 同时返回该目录下的总项数，可选排序：
// orderByField：排序字段，支持 "name", "size", "update_time"
// orderByType：排序类型，支持 "asc"（升序）和 "desc"（降序）
func (d *Disk) FileListWithPage(dir string, page, pageSize int, orderByField, orderByType string, keywords ...string) ([]*File, int64, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, 0, errors.New("无效的分页参数")
	}

	fullPath := d.fullPath(dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, 0, err
	}

	keyword := ""
	if len(keywords) > 0 {
		keyword = strings.ToLower(strings.TrimSpace(keywords[0]))
	}
	var files []*File
	for _, entry := range entries {
		if keyword != "" && !strings.Contains(strings.ToLower(entry.Name()), keyword) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue // 跳过获取信息失败的项
		}
		file := &File{
			Name:       entry.Name(),
			Size:       info.Size(),
			Mode:       uint32(info.Mode().Perm()),
			IsDir:      entry.IsDir(),
			UpdateTime: info.ModTime().Unix(),
		}
		files = append(files, file)
	}
	total := int64(len(files))

	// 排序：先将排序类型转换为小写，方便比较
	orderByType = strings.ToLower(orderByType)
	switch orderByField {
	case "name":
		sort.Slice(files, func(i, j int) bool {
			if orderByType == "desc" {
				return files[i].Name > files[j].Name
			}
			return files[i].Name < files[j].Name
		})
		break
	case "size":
		sort.Slice(files, func(i, j int) bool {
			if orderByType == "desc" {
				return files[i].Size > files[j].Size
			}
			return files[i].Size < files[j].Size
		})
		break
	case "update_time":
		sort.Slice(files, func(i, j int) bool {
			if orderByType == "desc" {
				return files[i].UpdateTime > files[j].UpdateTime
			}
			return files[i].UpdateTime < files[j].UpdateTime
		})
		break
	}
	// 分页处理
	start := int64((page - 1) * pageSize)
	if start >= total {
		// 分页起始位置超出总数，返回空列表
		return []*File{}, total, nil
	}
	end := start + int64(pageSize)
	if end > total {
		end = total
	}
	return files[start:end], total, nil
}

// FileTree 递归获取目录结构信息
// dir: 需要遍历的目录相对路径
// 返回值：包含完整目录结构的File切片
func (d *Disk) FileTree(dir string) ([]*File, error) {
	return d.buildFileTree(d.fullPath(dir), true)
}

/******************** 内部辅助方法 ********************/

func (d *Disk) fullPath(name string) string {
	cleanPath := filepath.Clean(name)
	if filepath.IsAbs(cleanPath) {
		return cleanPath
	}
	return filepath.Join(d.root, cleanPath)
}

func (d *Disk) isDirPath(name string) bool {
	return strings.HasSuffix(name, "/") || strings.HasSuffix(name, "\\") || filepath.Base(name) == ""
}

func (d *Disk) createFileIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		return file.Close()
	}
	return nil
}

func (d *Disk) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

/******************** 核心复制逻辑 ********************/

func (d *Disk) copyTree(src, dest string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return d.copySymlink(src, dest)
	}

	if srcInfo.IsDir() {
		if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
			return err
		}

		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, _ := filepath.Rel(src, path)
			targetPath := filepath.Join(dest, relPath)

			if path == src {
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				return d.copySymlink(path, targetPath)
			}
			if info.IsDir() {
				return os.Mkdir(targetPath, info.Mode())
			}

			return d.copyFile(path, targetPath)
		})
	}
	return d.copyFile(src, dest)
}

func (d *Disk) copyFile(src, dest string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return d.copySymlink(src, dest)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	if _, err = io.CopyBuffer(destFile, srcFile, make([]byte, bufferSize)); err != nil {
		return err
	}

	return os.Chtimes(dest, time.Now(), srcInfo.ModTime())
}

func (d *Disk) copySymlink(src, dest string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.Symlink(target, dest)
}

/******************** 进度跟踪逻辑 ********************/

func (d *Disk) calculateTotal(path string) (int64, int64, int64, error) {
	return d.calculateTotalWithContext(context.Background(), path)
}

func (d *Disk) calculateTotalWithContext(ctx context.Context, path string) (int64, int64, int64, error) {
	var totalSize, fileCount, dirCount int64
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			atomic.AddInt64(&totalSize, info.Size())
			atomic.AddInt64(&fileCount, 1)
		} else {
			atomic.AddInt64(&dirCount, 1)
		}
		return nil
	})
	return totalSize, fileCount, dirCount, err
}

func (d *Disk) copyTreeWithProgress(ctx context.Context, src, dest string, totalSize, totalFiles int64, copied *atomic.Int64, callback func(int64, int64, string, int64, int64) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	var fileIndex atomic.Int64
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return d.copySymlinkWithProgress(ctx, src, dest, srcInfo.Size(), totalSize, copied, callback, totalFiles, 0)
	}
	if srcInfo.IsDir() {
		if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
			return err
		}

		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(dest, relPath)

			if path == src {
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				currentIndex := fileIndex.Add(1) - 1
				return d.copySymlinkWithProgress(ctx, path, targetPath, info.Size(), totalSize, copied, callback, totalFiles, currentIndex)
			}
			if info.IsDir() {
				return os.Mkdir(targetPath, info.Mode())
			}

			currentIndex := fileIndex.Add(1) - 1
			return d.copyFileWithProgress(ctx, path, targetPath, totalSize, copied, callback, totalFiles, currentIndex)
		})
	}
	return d.copyFileWithProgress(ctx, src, dest, totalSize, copied, callback, totalFiles, 0)
}

func (d *Disk) copyFileWithProgress(ctx context.Context, src, dest string, totalSize int64, copied *atomic.Int64, callback func(int64, int64, string, int64, int64) bool, totalFiles, currentIndex int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return d.copySymlinkWithProgress(ctx, src, dest, info.Size(), totalSize, copied, callback, totalFiles, currentIndex)
	}
	if callback != nil && !callback(copied.Load(), totalSize, filepath.Base(src), totalFiles, currentIndex) {
		return context.Canceled
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()
	var (
		buf       = make([]byte, bufferSize)
		lastFlush int64
	)
	for {
		if err = ctx.Err(); err != nil {
			return err
		}

		n, err := srcFile.Read(buf)
		if n > 0 {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if _, wErr := destFile.Write(buf[:n]); wErr != nil {
				return wErr
			}

			newCopied := copied.Add(int64(n))
			if newCopied-lastFlush >= progressUnit || err == io.EOF {
				if callback != nil && !callback(newCopied, totalSize, filepath.Base(src), totalFiles, currentIndex) {
					return context.Canceled
				}
				lastFlush = newCopied / progressUnit * progressUnit
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
		}

		if err == io.EOF {
			if callback != nil && !callback(copied.Load(), totalSize, filepath.Base(src), totalFiles, currentIndex) {
				return context.Canceled
			}
			break
		}
		if err != nil {
			return err
		}
	}

	if err = ctx.Err(); err != nil {
		return err
	}
	_ = os.Chtimes(dest, time.Now(), info.ModTime())
	_ = os.Chmod(dest, info.Mode())

	return nil
}

func (d *Disk) copySymlinkWithProgress(ctx context.Context, src, dest string, size, totalSize int64, copied *atomic.Int64, callback func(int64, int64, string, int64, int64) bool, totalFiles, currentIndex int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if callback != nil && !callback(copied.Load(), totalSize, filepath.Base(src), totalFiles, currentIndex) {
		return context.Canceled
	}
	if err := d.copySymlink(src, dest); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	current := copied.Add(size)
	if callback != nil && !callback(current, totalSize, filepath.Base(src), totalFiles, currentIndex) {
		return context.Canceled
	}
	return nil
}

/******************** 其他辅助方法 ********************/

func (d *Disk) validateTransferDestination(src, dest string) error {
	srcPath, err := d.resolvePath(src)
	if err != nil {
		return fmt.Errorf("解析源路径失败: %w", err)
	}
	destPath, err := d.resolvePath(dest)
	if err != nil {
		return fmt.Errorf("解析目标路径失败: %w", err)
	}

	relPath, err := filepath.Rel(srcPath, destPath)
	if err != nil {
		return nil
	}
	if relPath == "." {
		return errors.New("目标路径不能与源路径相同")
	}

	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relPath) {
		return errors.New("目标路径不能位于源目录内")
	}
	return nil
}

func (d *Disk) resolvePath(path string) (string, error) {
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

func (d *Disk) prepareDestination(dest string) error {
	if err := os.RemoveAll(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clean destination failed: %w", err)
	}
	return os.MkdirAll(filepath.Dir(dest), 0755)
}

func cleanPartialDestination(dest string, operationErr error) error {
	if err := os.RemoveAll(dest); err != nil {
		return errors.Join(operationErr, fmt.Errorf("clean partial destination failed: %w", err))
	}
	return operationErr
}

func (d *Disk) buildFileTree(root string, recursive bool) ([]*File, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var files []*File
	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := &File{
			Name:       entry.Name(),
			Size:       info.Size(),
			Mode:       uint32(info.Mode().Perm()),
			IsDir:      entry.IsDir(),
			UpdateTime: info.ModTime().Unix(),
		}

		if recursive && entry.IsDir() {
			children, _ := d.buildFileTree(fullPath, true)
			file.Child = children
		}

		files = append(files, file)
	}
	return files, nil
}
