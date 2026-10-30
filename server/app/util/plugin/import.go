package plugin

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Import 解包并检查插件清单，返回临时目录，由调用方提交到正式目录或清理。
func (m *Manager) Import(ctx context.Context, path string) (*Package, error) {
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(m.ctx, cancel)
	defer func() { stop(); cancel() }()
	m.mu.Lock()
	if m.ctx.Err() != nil || ctx.Err() != nil {
		m.mu.Unlock()
		return nil, errors.New("插件管理器已关闭")
	}
	m.imports.Add(1)
	m.mu.Unlock()
	defer m.imports.Done()
	maxArchiveSize, maxExpandedSize := m.options.MaxArchiveSize, m.options.MaxExpandedSize
	if maxArchiveSize <= 0 {
		maxArchiveSize = 32 << 20
	}
	if maxExpandedSize <= 0 {
		maxExpandedSize = 64 << 20
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开插件包失败: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxArchiveSize {
		return nil, fmt.Errorf("插件包必须是普通 ZIP 文件，且不超过 %d MiB", maxArchiveSize>>20)
	}
	archive, err := zip.NewReader(file, info.Size())
	if err != nil {
		return nil, fmt.Errorf("读取插件ZIP包失败: %w", err)
	}
	if len(archive.File) == 0 || len(archive.File) > MaxPackageFiles {
		return nil, errors.New("插件包文件数量必须在1到1000之间")
	}
	if err = os.MkdirAll(m.root, 0700); err != nil {
		return nil, err
	}
	temporary, err := os.MkdirTemp(m.root, ".install-")
	if err != nil {
		return nil, err
	}
	retained := false
	defer func() {
		if !retained {
			_ = os.RemoveAll(temporary)
		}
	}()
	root, err := os.OpenRoot(temporary)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	seen := make(map[string]bool, len(archive.File))
	var expanded, codeSize, imageSize int64
	for _, item := range archive.File {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(item.Name, "/")
		if name == "" || !filepath.IsLocal(name) || strings.ContainsAny(name, `\:`) {
			return nil, errors.New("插件包包含非法文件路径")
		}
		for _, part := range strings.Split(name, "/") {
			if part == "" || part == "." || part == ".." {
				return nil, errors.New("插件包包含非法文件路径")
			}
		}
		if seen[name] {
			return nil, fmt.Errorf("插件包包含重复文件: %s", name)
		}
		seen[name] = true
		mode := item.Mode()
		if mode.Type() != 0 && mode.Type() != os.ModeDir {
			return nil, errors.New("插件包不支持符号链接或特殊文件")
		}
		imageArchive, err := PackageImageEntry(name, mode.IsDir())
		if err != nil {
			return nil, err
		}
		if mode.IsDir() {
			if err = root.MkdirAll(name, 0700); err != nil {
				return nil, err
			}
			continue
		}
		remaining := maxExpandedSize - expanded
		contentRemaining := MaxSourceSize - codeSize
		if imageArchive {
			contentRemaining = MaxImageSize - imageSize
		}
		if contentRemaining < remaining {
			remaining = contentRemaining
		}
		if remaining < 0 || item.UncompressedSize64 > uint64(remaining) {
			return nil, errors.New("插件包超过大小限制：源码最多 64 MiB，镜像最多 16 GiB，展开总量不得超过宿主配置")
		}
		if err = root.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return nil, err
		}
		input, err := item.Open()
		if err != nil {
			return nil, fmt.Errorf("解压插件文件失败: %w", err)
		}
		output, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			_ = input.Close()
			return nil, err
		}
		size, copyErr := io.Copy(output, io.LimitReader(&importReader{ctx: ctx, reader: input}, remaining+1))
		closeErr := errors.Join(output.Close(), input.Close())
		if copyErr != nil || closeErr != nil {
			return nil, fmt.Errorf("解压插件文件失败: %w", errors.Join(copyErr, closeErr))
		}
		expanded += size
		if imageArchive {
			imageSize += size
		} else {
			codeSize += size
		}
		if size > remaining || uint64(size) != item.UncompressedSize64 {
			return nil, errors.New("插件包解压大小不正确或超过允许上限")
		}
	}
	if !seen[m.options.ManifestName] {
		return nil, fmt.Errorf("插件包根目录缺少 %s", m.options.ManifestName)
	}
	result, err := m.info(filepath.Base(temporary))
	if err != nil {
		return nil, err
	}
	// 返回前释放目录句柄，调用方可以在 Windows 上直接移动整个目录。
	if err = root.Close(); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	retained = true
	return result, nil
}

// importReader 在大镜像复制期间也响应管理器关闭，不必等待整个文件结束。
type importReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *importReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
