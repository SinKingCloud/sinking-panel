package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (s *Manager) packPathArchive(ctx context.Context, path string, directory bool, output io.Writer) (err error) {
	checked, err := s.validateSourcePath(path, directory)
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
	if directory {
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
	compressed := gzip.NewWriter(output)
	archive := tar.NewWriter(compressed)
	defer func() {
		err = errors.Join(err, archive.Close(), compressed.Close(), ctx.Err())
	}()
	return s.writePathEntry(ctx, source, name, filepath.Base(path), archive, make([]byte, 32*1024))
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
