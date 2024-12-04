package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// New 本地文件适配器
func New(root string) *File {
	local := &File{}
	local.SetPathPrefix(root)
	return local
}

type File struct {
	pathPrefix    string // 前缀
	pathSeparator string // 分割符
}

// SetPathPrefix 设置前缀
func (f *File) SetPathPrefix(prefix string) {
	if prefix == "" {
		f.pathPrefix = ""
		return
	}
	// 只在传入的前缀不以'/'结尾时添加'/'，以保持灵活性
	if !strings.HasSuffix(prefix, "/") {
		f.pathSeparator = "/"
		f.pathPrefix = prefix + f.pathSeparator
	} else {
		f.pathSeparator = "/"
		f.pathPrefix = strings.TrimSuffix(prefix, "/") + f.pathSeparator
	}
}

// EnsureDirectory 确认文件夹
func (f *File) EnsureDirectory(root string, access uint32) error {
	// 先判断目录是否已存在
	_, err := os.Stat(root)
	if err == nil {
		return nil
	}

	err = os.MkdirAll(root, f.FormatPerm(access))
	if err != nil {
		return errors.New("创建文件夹失败," + err.Error())
	}

	// 确认创建后是目录
	if !f.IsDir(root) {
		return errors.New("创建根目录文件夹失败")
	}
	return nil
}

// GetPathPrefix 获取前缀
func (f *File) GetPathPrefix() string {
	return f.pathPrefix
}

// ApplyPathPrefix 添加前缀
func (f *File) ApplyPathPrefix(path string) string {
	return f.GetPathPrefix() + strings.TrimPrefix(path, "/")
}

// RemovePathPrefix 移除前缀
func (f *File) RemovePathPrefix(path string) string {
	prefix := f.GetPathPrefix()
	return strings.TrimPrefix(path, prefix)
}

// Has 是否存在
func (f *File) Has(path string) bool {
	location := f.ApplyPathPrefix(path)
	_, err := os.Stat(location)
	return err == nil || os.IsExist(err)
}

// Write 上传
func (f *File) Write(path string, contents string, access uint32) (map[string]any, error) {
	location := f.ApplyPathPrefix(path)
	out, createErr := os.Create(location)
	if createErr != nil {
		return nil, errors.New("创建文件失败," + createErr.Error())
	}
	defer func() {
		_ = out.Close()
	}()
	_, writeErr := out.WriteString(contents)
	if writeErr != nil {
		return nil, errors.New("修改文件失败," + writeErr.Error())
	}
	size, sizeErr := f.FileSize(location)
	if sizeErr != nil {
		return nil, errors.New("获取文件信息失败," + sizeErr.Error())
	}
	result := map[string]any{
		"type":     "file",
		"size":     size,
		"path":     path,
		"contents": contents,
	}
	if access > 0 {
		_, _ = f.SetVisibility(location, access)
	}
	return result, nil
}

// WriteStream 上传 Stream 文件类型
func (f *File) WriteStream(path string, stream io.Reader, access uint32) (map[string]any, error) {
	// 对传入的流进行有效性校验
	if stream == nil {
		return nil, errors.New("传入的文件流不能为空")
	}
	location := f.ApplyPathPrefix(path)
	newFile, createErr := os.Create(location)
	if createErr != nil {
		return nil, errors.New("创建文件失败," + createErr.Error())
	}
	defer func() {
		_ = newFile.Close()
	}()
	_, copyErr := io.Copy(newFile, stream)
	if copyErr != nil {
		return nil, errors.New("写入文件流失败, " + copyErr.Error())
	}
	result := map[string]any{
		"type": "file",
		"path": path,
	}
	_, _ = f.SetVisibility(location, access)
	return result, nil
}

// Update 更新
func (f *File) Update(path string, contents string) (map[string]any, error) {
	location := f.ApplyPathPrefix(path)
	out, createErr := os.Create(location)
	if createErr != nil {
		return nil, errors.New("创建文件失败," + createErr.Error())
	}
	defer func() {
		_ = out.Close()
	}()
	_, writeErr := out.WriteString(contents)
	if writeErr != nil {
		return nil, errors.New("写入文件失败," + writeErr.Error())
	}
	size, sizeErr := f.FileSize(location)
	if sizeErr != nil {
		return nil, errors.New("获取文件信息失败," + sizeErr.Error())
	}
	result := map[string]any{
		"type":     "file",
		"size":     size,
		"path":     path,
		"contents": contents,
	}
	return result, nil
}

// UpdateStream 更新
func (f *File) UpdateStream(path string, stream io.Reader, access uint32) (map[string]any, error) {
	return f.WriteStream(path, stream, access)
}

// Read 读取
func (f *File) Read(path string) (map[string]any, error) {
	location := f.ApplyPathPrefix(path)
	file, openErr := os.Open(location)
	if openErr != nil {
		return nil, errors.New("打开文件失败," + openErr.Error())
	}
	defer func() {
		_ = file.Close()
	}()
	data, readAllErr := io.ReadAll(file)
	if readAllErr != nil {
		// 对读取错误进行更细致的处理，这里只是简单示例
		switch readAllErr {
		case io.EOF:
			return nil, errors.New("读取到文件末尾")
		default:
			return nil, errors.New("读取文件失败," + readAllErr.Error())
		}
	}
	contents := fmt.Sprintf("%s", data)
	return map[string]any{
		"type":     "file",
		"path":     path,
		"contents": contents,
	}, nil
}

// ReadStream 读取成文件流
// 打开文件需要手动关闭
func (f *File) ReadStream(path string) (map[string]any, error) {
	location := f.ApplyPathPrefix(path)
	stream, err := os.Open(location)
	if err != nil {
		return nil, errors.New("打开文件失败," + err.Error())
	}
	return map[string]any{
		"type":   "file",
		"path":   path,
		"stream": stream,
	}, nil
}

// Rename 重命名
func (f *File) Rename(path string, newpath string) error {
	// 对源文件和目标文件路径进行有效性校验
	if !f.Has(path) {
		return errors.New("源文件不存在")
	}
	if f.Has(newpath) {
		return errors.New("目标文件已存在")
	}
	location := f.ApplyPathPrefix(path)
	destination := f.ApplyPathPrefix(newpath)
	err := os.Rename(location, destination)
	if err != nil {
		return errors.New("重命名文件失败," + err.Error())
	}
	return nil
}

// Copy 复制
func (f *File) Copy(path string, newpath string) error {
	// 对源文件和目标文件路径进行有效性校验
	if !f.Has(path) {
		return errors.New("源文件不存在")
	}
	location := f.ApplyPathPrefix(path)
	destination := f.ApplyPathPrefix(newpath)
	locationStat, e := os.Stat(location)
	if e != nil {
		return e
	}
	if !locationStat.Mode().IsRegular() {
		return fmt.Errorf("%s 不是一个正常的文件", path)
	}
	src, openErr := os.Open(location)
	if openErr != nil {
		return openErr
	}
	defer func() {
		_ = src.Close()
	}()
	// 确保目标文件所在目录存在
	dir := filepath.Dir(destination)
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.New("创建目标文件所在目录失败")
		}
	}
	dsc, createErr := os.Create(destination)
	if createErr != nil {
		return createErr
	}
	defer func() {
		_ = dsc.Close()
	}()
	_, copyErr := io.Copy(dsc, src)
	if copyErr != nil {
		return errors.New("复制失败," + copyErr.Error())
	}
	return nil
}

// Delete 删除
func (f *File) Delete(path string) error {
	location := f.ApplyPathPrefix(path)
	// 先判断文件是否存在
	_, err := os.Stat(location)
	if err != nil {
		return errors.New("文件不存在")
	}
	// 再判断是否为文件类型
	if !f.IsFile(location) {
		return errors.New("文件删除失败")
	}
	if err := os.Remove(location); err != nil {
		return errors.New("文件删除失败," + err.Error())
	}
	return nil
}

// DeleteDir 删除文件夹
func (f *File) DeleteDir(dirname string) error {
	location := f.ApplyPathPrefix(dirname)
	// 先判断文件夹是否存在
	_, err := os.Stat(location)
	if err != nil {
		return errors.New("文件夹不存在")
	}
	// 再判断是否为文件夹类型
	if !f.IsDir(location) {
		return errors.New("文件夹删除失败")
	}
	if err := os.RemoveAll(location); err != nil {
		return errors.New("文件夹删除失败," + err.Error())
	}
	return nil
}

// CreateDir 创建文件夹
func (f *File) CreateDir(dirname string, access uint32) (map[string]string, error) {
	location := f.ApplyPathPrefix(dirname)
	err := os.MkdirAll(location, f.FormatPerm(access))
	if err != nil {
		return nil, errors.New("创建文件夹失败," + err.Error())
	}
	// 确认创建后是目录
	if !f.IsDir(location) {
		return nil, errors.New("文件夹创建失败")
	}
	data := map[string]string{
		"path": dirname,
		"type": "dir",
	}
	return data, nil
}

// ListContents 列出内容
func (f *File) ListContents(directory string, recursive ...bool) ([]map[string]any, error) {
	location := f.ApplyPathPrefix(directory)
	// 先判断目录是否存在
	_, err := os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, errors.New("指定目录不存在")
		}
		return []map[string]any{}, errors.New("获取目录信息失败")
	}
	var iterator []map[string]any
	if len(recursive) > 0 && recursive[0] {
		iterator, _ = f.GetRecursiveDirectoryIterator(location)
	} else {
		iterator, _ = f.GetDirectoryIterator(location)
	}
	var result []map[string]any
	for _, file := range iterator {
		path, _ := f.NormalizeFileInfo(file)

		result = append(result, path)
	}
	return result, nil
}

func (f *File) GetMetadata(path string) (map[string]any, error) {
	location := f.ApplyPathPrefix(path)
	info := f.FileInfo(location)
	return f.NormalizeFileInfo(info)
}

func (f *File) GetSize(path string) (map[string]any, error) {
	return f.GetMetadata(path)
}

func (f *File) GetMimetype(path string) (map[string]any, error) {
	location := f.ApplyPathPrefix(path)
	f2, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f2.Close()
	}()
	// 头部字节
	buffer := make([]byte, 32)
	if _, err := f2.Read(buffer); err != nil {
		return nil, err
	}
	mimetype := http.DetectContentType(buffer)
	return map[string]any{
		"path":     path,
		"type":     "file",
		"mimetype": mimetype,
	}, nil
}

func (f *File) GetTimestamp(path string) (map[string]any, error) {
	return f.GetMetadata(path)
}

// GetVisibility 设置文件的权限
func (f *File) GetVisibility(path string) (map[string]string, error) {
	location := f.ApplyPathPrefix(path)
	permissions, _ := f.FileMode(location)
	permission := fmt.Sprintf("%o", permissions)
	data := map[string]string{
		"path":       path,
		"visibility": permission,
	}
	return data, nil
}

// SetVisibility 设置文件的权限
func (f *File) SetVisibility(path string, access uint32) (map[string]string, error) {
	location := f.ApplyPathPrefix(path)
	// 对传入的权限值进行有效性校验
	if access < 0 {
		return nil, errors.New("权限值不能为负数")
	}
	e := os.Chmod(location, f.FormatPerm(access))
	if e != nil {
		return nil, errors.New("设置文件权限失败")
	}
	data := map[string]string{
		"path":       path,
		"visibility": strconv.Itoa(int(access)),
	}
	return data, nil
}

// NormalizeFileInfo NormalizeFileInfo
func (f *File) NormalizeFileInfo(file map[string]any) (map[string]any, error) {
	return f.MapFileInfo(file)
}

// GuardAgainstUnreadableFileInfo 是否可读
func (f *File) GuardAgainstUnreadableFileInfo(fp string) error {
	_, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	return nil
}

// GetRecursiveDirectoryIterator 获取全部文件
func (f *File) GetRecursiveDirectoryIterator(path string) ([]map[string]any, error) {
	var files []map[string]any
	err := filepath.Walk(path, func(wPath string, info os.FileInfo, err error) error {
		var fileType string
		if info.IsDir() {
			fileType = "dir"
		} else {
			fileType = "file"
		}
		files = append(files, map[string]any{
			"type":      fileType,
			"path":      path,
			"filename":  info.Name(),
			"pathname":  path + "/" + info.Name(),
			"timestamp": info.ModTime().Unix(),
			"info":      info,
		})
		return nil
	})
	if err != nil {
		return nil, errors.New("获取文件夹列表失败")
	}
	return files, nil
}

// GetDirectoryIterator 一级目录索引
func (f *File) GetDirectoryIterator(path string) ([]map[string]any, error) {
	fs, err := os.ReadDir(path)
	if err != nil {
		return []map[string]any{}, err
	}
	sz := len(fs)
	if sz == 0 {
		return []map[string]any{}, nil
	}
	ret := make([]map[string]any, 0, sz)
	for i := 0; i < sz; i++ {
		info := fs[i]
		name := info.Name()
		stat, _ := info.Info()
		if name != "." && name != ".." {
			var fileType string
			if info.IsDir() {
				fileType = "dir"
			} else {
				fileType = "file"
			}
			ret = append(ret, map[string]any{
				"type":      fileType,
				"path":      path,
				"filename":  name,
				"pathname":  path + "/" + name,
				"timestamp": stat.ModTime().Unix(),
				"info":      info,
			})
		}
	}
	return ret, nil
}

func (f *File) FileInfo(path string) map[string]any {
	info, e := os.Stat(path)
	if e != nil {
		return nil
	}
	var fileType string
	if info.IsDir() {
		fileType = "dir"
	} else {
		fileType = "file"
	}
	return map[string]any{
		"type":      fileType,
		"path":      filepath.Dir(path),
		"filename":  info.Name(),
		"pathname":  path,
		"timestamp": info.ModTime().Unix(),
		"info":      info,
	}
}

func (f *File) GetFilePath(file map[string]any) string {
	location := file["pathname"].(string)
	path := f.RemovePathPrefix(location)
	return strings.Trim(strings.Replace(path, "\\", "/", -1), "/")
}

// MapFileInfo 获取全部文件
func (f *File) MapFileInfo(data map[string]any) (map[string]any, error) {
	// 对传入的data中的type字段进行有效性校验
	if _, ok := data["type"]; !ok {
		return nil, errors.New("传入的文件信息缺少type字段")
	}
	normalized := map[string]any{
		"type":      data["type"],
		"path":      f.GetFilePath(data),
		"timestamp": data["timestamp"],
	}
	if data["type"] == "file" {
		normalized["size"] = data["info"].(os.FileInfo).Size()
	}
	return normalized, nil
}

// IsFile 是否为文件
func (f *File) IsFile(fp string) bool {
	return !f.IsDir(fp)
}

// IsDir 是否为目录
func (f *File) IsDir(fp string) bool {
	f2, e := os.Stat(fp)
	if e != nil {
		return false
	}
	return f2.IsDir()
}

// FileSize 文件大小
func (f *File) FileSize(fp string) (int64, error) {
	f2, e := os.Stat(fp)
	if e != nil {
		return 0, e
	}
	return f2.Size(), nil
}

// FileMode 文件权限
func (f *File) FileMode(fp string) (uint32, error) {
	f2, e := os.Stat(fp)
	if e != nil {
		return 0, e
	}
	perm := f2.Mode().Perm()
	return uint32(perm), nil
}

// FormatPerm 转换
func (f *File) FormatPerm(i uint32) os.FileMode {
	return os.FileMode(i)
}

// Symlink 软链接
func (f *File) Symlink(target, link string) error {
	// 对传入的目标路径和链接路径进行有效性校验
	if target == "" {
		return errors.New("目标路径不能为空")
	}
	if link == "" {
		return errors.New("链接路径不能为空")
	}
	return os.Symlink(target, link)
}

// Readlink 读取链接
func (f *File) Readlink(link string) (string, error) {
	return os.Readlink(link)
}

// IsSymlink 是否为软链接
func (f *File) IsSymlink(m os.FileMode) bool {
	return m&os.ModeSymlink != 0
}
