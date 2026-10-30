package software

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"server/app/util/plugin"
)

const softwareArchiveLimit int64 = plugin.MaxImageSize + plugin.MaxSourceSize + (16 << 20)

// packDirectory 只携带运行源码与可选镜像；大归档以 ZIP Store 流式复制。
func (s *service) packDirectory(ctx context.Context, directory string, output io.Writer) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	writer := zip.NewWriter(&archiveWriter{ctx: ctx, writer: output, remaining: softwareArchiveLimit})
	var count int
	var codeSize, imageSize int64
	for _, name := range []string{"software.json", "server", "web", "images"} {
		if name == "web" || name == "images" {
			if _, err := root.Lstat(name); errors.Is(err, fs.ErrNotExist) {
				continue
			}
		}
		err = fs.WalkDir(root.FS(), name, func(relative string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("软件目录不支持符号链接或特殊文件: %s", relative)
			}
			imageArchive, err := plugin.PackageImageEntry(relative, false)
			if err != nil {
				return err
			}
			count++
			if imageArchive {
				imageSize += info.Size()
			} else {
				codeSize += info.Size()
			}
			if count > plugin.MaxPackageFiles || codeSize > plugin.MaxSourceSize || imageSize > plugin.MaxImageSize {
				return errors.New("软件包最多 1000 个文件、64 MiB 源码和 16 GiB 镜像归档")
			}
			header := &zip.FileHeader{Name: relative, Method: zip.Deflate}
			if strings.HasPrefix(relative, "images/") {
				header.Method = zip.Store
			}
			target, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}
			input, err := root.Open(relative)
			if err != nil {
				return err
			}
			current, statErr := input.Stat()
			if statErr != nil || !current.Mode().IsRegular() || !os.SameFile(info, current) || current.Size() != info.Size() {
				return errors.Join(fmt.Errorf("打包期间文件发生变动: %s", relative), statErr, input.Close())
			}
			written, copyErr := io.Copy(target, io.LimitReader(input, info.Size()+1))
			closeErr := input.Close()
			if written != info.Size() {
				return errors.Join(fmt.Errorf("打包期间文件发生变动: %s", relative), copyErr, closeErr)
			}
			return errors.Join(copyErr, closeErr)
		})
		if err != nil {
			break
		}
	}
	return errors.Join(err, writer.Close())
}

// archiveWriter 同时限制 ZIP 实际体积，避免镜像加上源码和目录元数据越过上限。
type archiveWriter struct {
	ctx       context.Context
	writer    io.Writer
	remaining int64
}

func (w *archiveWriter) Write(data []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	if int64(len(data)) > w.remaining {
		return 0, errors.New("软件 ZIP 包超过允许大小")
	}
	n, err := w.writer.Write(data)
	w.remaining -= int64(n)
	return n, err
}
