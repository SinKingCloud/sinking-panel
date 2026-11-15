package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// packPathsArchive 流式归档多个来源，保留相对根目录的路径，避免同名文件覆盖。
func (s *Manager) packPathsArchive(ctx context.Context, paths []string, output io.Writer) (err error) {
	compressed := gzip.NewWriter(output)
	archive := tar.NewWriter(compressed)
	defer func() {
		err = errors.Join(err, archive.Close(), compressed.Close(), ctx.Err())
	}()
	metadata, err := json.Marshal(map[string]interface{}{
		"version": 1,
		"time":    time.Now().UTC(),
		"paths":   paths,
	})
	if err != nil {
		return err
	}
	if err = archive.WriteHeader(&tar.Header{
		Name:    "backup.json",
		Mode:    0600,
		Size:    int64(len(metadata)),
		ModTime: time.Now(),
	}); err != nil {
		return err
	}
	if _, err = archive.Write(metadata); err != nil {
		return err
	}
	buffer := make([]byte, 32*1024)
	for _, path := range paths {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = func() error {
			checked, err := s.validateSourcePath(path)
			if err != nil {
				return err
			}
			if checked != path {
				return errors.New("备份来源路径已变化")
			}
			info, err := os.Lstat(path)
			if err != nil {
				return err
			}
			parent, name := filepath.Dir(path), filepath.Base(path)
			if info.IsDir() {
				parent, name = path, "."
			}
			source, err := os.OpenRoot(parent)
			if err != nil {
				return err
			}
			defer source.Close()
			current, err := source.Lstat(name)
			if err != nil || !os.SameFile(info, current) || info.Mode() != current.Mode() {
				return errors.New("备份来源已变化")
			}
			volume := filepath.VolumeName(path)
			archiveName := strings.TrimLeft(filepath.ToSlash(strings.TrimPrefix(path, volume)), "/")
			if volume != "" {
				archiveName = filepath.Join(strings.Trim(filepath.ToSlash(volume), "/:"), archiveName)
			}
			return s.writePathEntry(ctx, source, name, filepath.Join("files", archiveName), archive, buffer)
		}(); err != nil {
			return fmt.Errorf("备份来源 %q 失败: %w", path, err)
		}
	}
	return nil
}

func (s *Manager) writePathEntry(ctx context.Context, source *os.Root, name, archiveName string, archive *tar.Writer, buffer []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := source.Lstat(name)
	if err != nil {
		return err
	}
	link := ""
	if info.Mode()&os.ModeSymlink != 0 {
		link, err = source.Readlink(name)
		if err != nil {
			return err
		}
	} else if !info.IsDir() && !info.Mode().IsRegular() {
		return fmt.Errorf("不支持备份特殊文件: %s", archiveName)
	}
	header, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(archiveName)
	if err = archive.WriteHeader(header); err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	file, err := source.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(info, opened) || info.Mode() != opened.Mode() || info.Size() != opened.Size() || !info.ModTime().Equal(opened.ModTime()) {
		return fmt.Errorf("备份期间文件已变化: %s", archiveName)
	}
	if info.IsDir() {
		for {
			if err = ctx.Err(); err != nil {
				return err
			}
			entries, readErr := file.ReadDir(128)
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return readErr
			}
			for _, entry := range entries {
				if err = s.writePathEntry(ctx, source, filepath.Join(name, entry.Name()), filepath.Join(archiveName, entry.Name()), archive, buffer); err != nil {
					return err
				}
			}
			if errors.Is(readErr, io.EOF) {
				return nil
			}
		}
	}
	n, err := io.CopyBuffer(archive, &pathBackupReader{ctx: ctx, reader: file}, buffer)
	if err != nil {
		return err
	}
	after, err := file.Stat()
	if err != nil {
		return err
	}
	if n != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return fmt.Errorf("备份期间文件已变化: %s", archiveName)
	}
	return ctx.Err()
}

func (r *pathBackupReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}
