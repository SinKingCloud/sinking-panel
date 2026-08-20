package container

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type dockerManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type dockerConfig struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Created      string `json:"created"`
	Config       struct {
		Env        []string `json:"Env"`
		Entrypoint []string `json:"Entrypoint"`
		Cmd        []string `json:"Cmd"`
		WorkingDir string   `json:"WorkingDir"`
		User       string   `json:"User"`
	} `json:"config"`
}

type layerDirectoryMetadata struct {
	path   string
	header tar.Header
}

func (m *Manager) importDockerSave(source string) (*Image, error) {
	info, err := os.Lstat(source)
	if err != nil {
		return nil, fmt.Errorf("镜像文件不存在: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("镜像路径必须是普通文件")
	}
	digest, err := m.fileDigest(source)
	if err != nil {
		return nil, fmt.Errorf("计算镜像摘要失败: %w", err)
	}
	id := "sha256:" + digest
	metadataPath := filepath.Join(m.imagesRoot, digest+".json")
	expectedRootfs := filepath.Join(m.imagesRoot, digest)
	if data, readErr := os.ReadFile(metadataPath); readErr == nil {
		var image Image
		if json.Unmarshal(data, &image) == nil && image.ID == id {
			image.Rootfs = expectedRootfs
			if rootfsInfo, statErr := os.Lstat(expectedRootfs); statErr == nil && rootfsInfo.IsDir() && rootfsInfo.Mode()&os.ModeSymlink == 0 {
				return &image, nil
			}
		}
	}
	workDir, err := os.MkdirTemp(m.imagesRoot, ".import-")
	if err != nil {
		return nil, fmt.Errorf("创建镜像临时目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)
	archiveDir := filepath.Join(workDir, "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return nil, err
	}
	if err := m.extractOuterArchive(source, archiveDir); err != nil {
		return nil, fmt.Errorf("解包 Docker 镜像失败: %w", err)
	}
	manifestData, err := os.ReadFile(filepath.Join(archiveDir, "manifest.json"))
	if err != nil {
		return nil, errors.New("镜像归档缺少 manifest.json，不是 docker save 格式")
	}
	var manifests []dockerManifest
	if err := json.Unmarshal(manifestData, &manifests); err != nil || len(manifests) == 0 {
		return nil, errors.New("镜像 manifest.json 无效")
	}
	manifest := manifests[0]
	if manifest.Config == "" || len(manifest.Layers) == 0 {
		return nil, errors.New("镜像没有可运行的配置或文件层")
	}
	configPath, err := m.safeJoin(archiveDir, filepath.FromSlash(manifest.Config))
	if err != nil {
		return nil, fmt.Errorf("镜像配置路径不安全: %w", err)
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取镜像配置失败: %w", err)
	}
	var config dockerConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("镜像配置无效: %w", err)
	}
	rootfsStage := filepath.Join(workDir, "rootfs")
	if err := os.MkdirAll(rootfsStage, 0755); err != nil {
		return nil, err
	}
	// 宿主机 umask 可能移除执行权限，镜像根目录必须允许非 root 用户进入。
	if err := os.Chmod(rootfsStage, 0755); err != nil {
		return nil, fmt.Errorf("设置镜像根目录权限失败: %w", err)
	}
	for _, layer := range manifest.Layers {
		layerPath, err := m.safeJoin(archiveDir, filepath.FromSlash(layer))
		if err != nil {
			return nil, fmt.Errorf("镜像层路径不安全: %w", err)
		}
		if err := m.applyLayer(layerPath, rootfsStage); err != nil {
			return nil, fmt.Errorf("应用镜像层 %s 失败: %w", layer, err)
		}
	}
	finalDir := expectedRootfs
	if err := os.RemoveAll(finalDir); err != nil {
		return nil, err
	}
	if err := os.Rename(rootfsStage, finalDir); err != nil {
		return nil, fmt.Errorf("保存镜像 rootfs 失败: %w", err)
	}
	createdAt := time.Now().Unix()
	if config.Created != "" {
		if created, parseErr := time.Parse(time.RFC3339Nano, config.Created); parseErr == nil {
			createdAt = created.Unix()
		}
	}
	name := ""
	if len(manifest.RepoTags) > 0 {
		name = manifest.RepoTags[0]
	}
	image := &Image{
		ID:           id,
		Name:         name,
		Tags:         append([]string(nil), manifest.RepoTags...),
		Rootfs:       finalDir,
		OS:           config.OS,
		Architecture: config.Architecture,
		User:         config.Config.User,
		Entrypoint:   append([]string(nil), config.Config.Entrypoint...),
		Command:      append([]string(nil), config.Config.Cmd...),
		Env:          append([]string(nil), config.Config.Env...),
		WorkingDir:   config.Config.WorkingDir,
		CreatedAt:    createdAt,
	}
	if image.OS == "" {
		image.OS = "linux"
	}
	if err := m.writeJSON(metadataPath, image); err != nil {
		_ = os.RemoveAll(finalDir)
		return nil, fmt.Errorf("保存镜像元数据失败: %w", err)
	}
	return image, nil
}

func (m *Manager) fileDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (m *Manager) extractOuterArchive(source, destination string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	reader, closer, err := m.openTarReader(file)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		path, err := m.safeJoin(destination, filepath.FromSlash(header.Name))
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := m.writeTarFile(path, reader, header.Mode, destination); err != nil {
				return err
			}
		default:
			// Docker save 外层归档只需要普通文件和目录。
			if header.Size > 0 {
				if _, err := io.Copy(io.Discard, reader); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *Manager) applyLayer(layerPath, rootfs string) error {
	file, err := os.Open(layerPath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader, closer, err := m.openTarReader(file)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}
	directories := make(map[string]layerDirectoryMetadata)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(filepath.ToSlash(header.Name), "./")
		if name == "" || name == "." {
			if header.Typeflag == tar.TypeDir {
				copyHeader := *header
				directories[rootfs] = layerDirectoryMetadata{path: rootfs, header: copyHeader}
			}
			continue
		}
		base := filepath.Base(name)
		parent := filepath.Dir(name)
		if strings.HasPrefix(base, ".wh.") {
			if base == ".wh..wh..opq" {
				dir, err := m.safeJoin(rootfs, filepath.FromSlash(parent))
				if err != nil {
					return err
				}
				info, err := m.archivePathInfo(dir, rootfs)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return err
				}
				if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return fmt.Errorf("镜像白化目录不安全: %s", parent)
				}
				entries, readErr := os.ReadDir(dir)
				if readErr != nil {
					return readErr
				}
				for _, entry := range entries {
					if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
						return err
					}
				}
			} else {
				whiteoutName := strings.TrimPrefix(base, ".wh.")
				if whiteoutName == "" {
					return errors.New("镜像白化路径无效")
				}
				target, err := m.safeJoin(rootfs, filepath.FromSlash(filepath.Join(parent, whiteoutName)))
				if err != nil {
					return err
				}
				if _, err := m.archivePathInfo(target, rootfs); errors.Is(err, os.ErrNotExist) {
					continue
				} else if err != nil {
					return err
				}
				if err := os.RemoveAll(target); err != nil {
					return err
				}
			}
			continue
		}
		destination, err := m.safeJoin(rootfs, filepath.FromSlash(name))
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := m.prepareDirectory(destination, rootfs); err != nil {
				return err
			}
			copyHeader := *header
			directories[destination] = layerDirectoryMetadata{path: destination, header: copyHeader}
		case tar.TypeReg, tar.TypeRegA:
			if err := m.writeTarFile(destination, reader, header.Mode, rootfs); err != nil {
				return err
			}
			if err := m.restoreLayerMetadata(destination, rootfs, header); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := m.ensureParent(destination, rootfs); err != nil {
				return err
			}
			_ = os.RemoveAll(destination)
			if err := os.Symlink(header.Linkname, destination); err != nil {
				return err
			}
			if err := m.restoreLayerMetadata(destination, rootfs, header); err != nil {
				return err
			}
		case tar.TypeLink:
			if err := m.ensureParent(destination, rootfs); err != nil {
				return err
			}
			linkPath, err := m.safeJoin(rootfs, filepath.FromSlash(header.Linkname))
			if err != nil {
				return err
			}
			linkInfo, err := m.archivePathInfo(linkPath, rootfs)
			if err != nil {
				return err
			}
			if !linkInfo.Mode().IsRegular() {
				return fmt.Errorf("镜像硬链接源不是普通文件: %s", header.Linkname)
			}
			_ = os.RemoveAll(destination)
			if err := os.Link(linkPath, destination); err != nil {
				return err
			}
			if err := m.restoreLayerMetadata(destination, rootfs, header); err != nil {
				return err
			}
		case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			return fmt.Errorf("镜像层包含暂不支持的特殊文件: %s", name)
		default:
			if header.Size > 0 {
				if _, err := io.Copy(io.Discard, reader); err != nil {
					return err
				}
			}
		}
	}
	// 子项全部创建完成后再恢复目录元数据，避免目录时间和权限被后续写入覆盖。
	orderedDirectories := make([]layerDirectoryMetadata, 0, len(directories))
	for _, directory := range directories {
		orderedDirectories = append(orderedDirectories, directory)
	}
	sort.Slice(orderedDirectories, func(i, j int) bool {
		return strings.Count(orderedDirectories[i].path, string(filepath.Separator)) > strings.Count(orderedDirectories[j].path, string(filepath.Separator))
	})
	for _, directory := range orderedDirectories {
		info, err := m.archivePathInfo(directory.path, rootfs)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			continue
		}
		if err := m.restoreLayerMetadata(directory.path, rootfs, &directory.header); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) openTarReader(file *os.File) (*tar.Reader, io.Closer, error) {
	var header [2]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return nil, nil, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, nil, err
	}
	if header[0] == 0x1f && header[1] == 0x8b {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, nil, err
		}
		return tar.NewReader(gzipReader), gzipReader, nil
	}
	return tar.NewReader(file), nil, nil
}

func (m *Manager) writeTarFile(path string, reader io.Reader, mode int64, root string) error {
	if err := m.ensureParent(path, root); err != nil {
		return err
	}
	_ = os.RemoveAll(path)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(mode)&07777)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, reader)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	// OpenFile 会应用宿主机 umask，写入后恢复镜像记录的权限。
	return os.Chmod(path, os.FileMode(mode)&07777)
}

func (m *Manager) restoreLayerMetadata(path, root string, header *tar.Header) error {
	info, err := m.archivePathInfo(path, root)
	if err != nil {
		return err
	}
	// Linux 容器依赖镜像内的 UID/GID；符号链接使用 Lchown，不能跟随到 rootfs 外。
	if runtime.GOOS == "linux" {
		if err := os.Lchown(path, header.Uid, header.Gid); err != nil {
			return fmt.Errorf("恢复镜像路径所有者失败 %s: %w", path, err)
		}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	if err := os.Chmod(path, os.FileMode(header.Mode)&07777); err != nil {
		return fmt.Errorf("恢复镜像路径权限失败 %s: %w", path, err)
	}
	if header.ModTime.IsZero() {
		return nil
	}
	accessTime := header.AccessTime
	if accessTime.IsZero() {
		accessTime = header.ModTime
	}
	if err := os.Chtimes(path, accessTime, header.ModTime); err != nil {
		return fmt.Errorf("恢复镜像路径时间失败 %s: %w", path, err)
	}
	return nil
}

func (m *Manager) prepareDirectory(path, root string) error {
	if err := m.ensureParent(path, root); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err == nil && !info.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return os.MkdirAll(path, 0755)
}

func (m *Manager) ensureParent(path, root string) error {
	return m.checkArchiveParent(path, root, true)
}

func (m *Manager) archivePathInfo(path, root string) (os.FileInfo, error) {
	if err := m.checkArchiveParent(path, root, false); err != nil {
		return nil, err
	}
	return os.Lstat(path)
}

func (m *Manager) checkArchiveParent(path, root string, create bool) error {
	parent := filepath.Dir(path)
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	parent, err = filepath.Abs(parent)
	if err != nil {
		return err
	}
	if err := m.ensureInside(root, path); err != nil {
		return err
	}
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return errors.New("归档根目录不安全")
	}
	if path == root {
		return nil
	}
	if err := m.ensureInside(root, parent); err != nil {
		return err
	}
	rel, err := filepath.Rel(root, parent)
	if err != nil {
		return err
	}
	current := root
	if rel == "." {
		return nil
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			if !create {
				return statErr
			}
			if err := os.Mkdir(current, 0755); err != nil {
				return err
			}
			continue
		}
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("归档父目录不能是符号链接")
		}
		if !info.IsDir() {
			return fmt.Errorf("归档父路径不是目录: %s", current)
		}
	}
	return nil
}

func (m *Manager) safeJoin(root, name string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	name = filepath.Clean(name)
	if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
		return "", errors.New("归档路径越界")
	}
	path := filepath.Join(root, name)
	if err := m.ensureInside(root, path); err != nil {
		return "", err
	}
	return path, nil
}

func (m *Manager) ensureInside(root, path string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("路径超出容器目录")
	}
	return nil
}
