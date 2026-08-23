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
	"reflect"
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

type dockerLayerMetadata struct {
	ID      string `json:"id"`
	Created string `json:"created"`
	Config  struct {
		Env        []string `json:"Env,omitempty"`
		Entrypoint []string `json:"Entrypoint,omitempty"`
		Cmd        []string `json:"Cmd,omitempty"`
		WorkingDir string   `json:"WorkingDir,omitempty"`
		User       string   `json:"User,omitempty"`
	} `json:"config"`
}

type layerDirectoryMetadata struct {
	path   string
	header tar.Header
}

func (m *Manager) importSave(source string) (*Image, error) {
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
				if err := m.validateImageSecurity(&image); err != nil {
					return nil, err
				}
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
	if err := m.normalizePlatformRootfsOwnership(rootfsStage); err != nil {
		return nil, fmt.Errorf("映射镜像 rootfs 所有者失败: %w", err)
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
		ID:              id,
		Name:            name,
		Tags:            append([]string(nil), manifest.RepoTags...),
		Rootfs:          finalDir,
		OS:              config.OS,
		Architecture:    config.Architecture,
		User:            config.Config.User,
		Entrypoint:      append([]string(nil), config.Config.Entrypoint...),
		Command:         append([]string(nil), config.Config.Cmd...),
		Env:             append([]string(nil), config.Config.Env...),
		WorkingDir:      config.Config.WorkingDir,
		CreatedAt:       createdAt,
		SecurityVersion: m.security.Version,
		UIDMapStart:     m.security.UIDMapStart,
		GIDMapStart:     m.security.GIDMapStart,
		IDMapSize:       m.security.IDMapSize,
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
	// Linux 容器依赖镜像内的 UID/GID；写入宿主前先转换到 User Namespace 映射。
	if runtime.GOOS == "linux" {
		uid, gid, err := m.mapOwnership(header.Uid, header.Gid)
		if err != nil {
			return fmt.Errorf("镜像路径所有者无效 %s: %w", path, err)
		}
		if err := os.Lchown(path, uid, gid); err != nil {
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

// exportImage 生成 Docker save 兼容的单层镜像归档。
func (m *Manager) exportImage(image *Image, destination string) error {
	if image == nil {
		return errors.New("镜像不能为空")
	}
	if strings.TrimSpace(destination) == "" {
		return errors.New("导出路径不能为空")
	}
	rootfsInfo, err := os.Lstat(image.Rootfs)
	if err != nil {
		return fmt.Errorf("读取镜像 rootfs 失败: %w", err)
	}
	if !rootfsInfo.IsDir() || rootfsInfo.Mode()&os.ModeSymlink != 0 {
		return errors.New("镜像 rootfs 不是安全目录")
	}
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("解析导出路径失败: %w", err)
	}
	if err := m.ensureInside(image.Rootfs, absDestination); err == nil {
		return errors.New("导出文件不能放在镜像 rootfs 内")
	}
	parent := filepath.Dir(absDestination)
	rootfsReal, err := filepath.EvalSymlinks(image.Rootfs)
	if err != nil {
		return fmt.Errorf("解析镜像 rootfs 失败: %w", err)
	}
	// 词法路径检查无法识别父目录符号链接。解析最近的已存在父目录，
	// 再拼回缺失部分，避免通过链接把导出文件写回镜像 rootfs。
	parentReal := parent
	missing := make([]string, 0, 4)
	for {
		if _, statErr := os.Lstat(parentReal); statErr == nil {
			parentReal, err = filepath.EvalSymlinks(parentReal)
			if err != nil {
				return fmt.Errorf("解析导出目录失败: %w", err)
			}
			for index := len(missing) - 1; index >= 0; index-- {
				parentReal = filepath.Join(parentReal, missing[index])
			}
			break
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("检查导出目录失败: %w", statErr)
		}
		next := filepath.Dir(parentReal)
		if next == parentReal {
			return errors.New("找不到导出目录的有效父路径")
		}
		missing = append(missing, filepath.Base(parentReal))
		parentReal = next
	}
	canonicalDestination := filepath.Join(parentReal, filepath.Base(absDestination))
	if rel, relErr := filepath.Rel(rootfsReal, canonicalDestination); relErr == nil &&
		(rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
		return errors.New("导出文件不能放在镜像 rootfs 内")
	}
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("创建导出目录失败: %w", err)
	}
	if info, statErr := os.Lstat(absDestination); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return errors.New("导出路径不是普通文件")
		}
		return errors.New("导出文件已存在")
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("检查导出文件失败: %w", statErr)
	}

	layerFile, err := os.CreateTemp(parent, ".container-layer-*.tar")
	if err != nil {
		return fmt.Errorf("创建镜像层临时文件失败: %w", err)
	}
	layerPath := layerFile.Name()
	defer os.Remove(layerPath)
	if err := m.writeRootfsTar(layerFile, image.Rootfs); err != nil {
		_ = layerFile.Close()
		return fmt.Errorf("生成镜像层失败: %w", err)
	}
	if err := layerFile.Sync(); err != nil {
		_ = layerFile.Close()
		return fmt.Errorf("同步镜像层失败: %w", err)
	}
	if err := layerFile.Close(); err != nil {
		return fmt.Errorf("关闭镜像层失败: %w", err)
	}
	layerInfo, err := os.Stat(layerPath)
	if err != nil {
		return fmt.Errorf("读取镜像层大小失败: %w", err)
	}
	layerDigest, err := m.fileDigest(layerPath)
	if err != nil {
		return fmt.Errorf("计算镜像层摘要失败: %w", err)
	}

	configData, err := json.Marshal(struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
		Created      string `json:"created"`
		Config       struct {
			Env        []string `json:"Env,omitempty"`
			Entrypoint []string `json:"Entrypoint,omitempty"`
			Cmd        []string `json:"Cmd,omitempty"`
			WorkingDir string   `json:"WorkingDir,omitempty"`
			User       string   `json:"User,omitempty"`
		} `json:"config"`
		Rootfs struct {
			Type    string   `json:"type"`
			DiffIDs []string `json:"diff_ids"`
		} `json:"rootfs"`
	}{
		Architecture: image.Architecture,
		OS:           image.OS,
		Created:      time.Unix(image.CreatedAt, 0).UTC().Format(time.RFC3339Nano),
		Config: struct {
			Env        []string `json:"Env,omitempty"`
			Entrypoint []string `json:"Entrypoint,omitempty"`
			Cmd        []string `json:"Cmd,omitempty"`
			WorkingDir string   `json:"WorkingDir,omitempty"`
			User       string   `json:"User,omitempty"`
		}{
			Env: image.Env, Entrypoint: image.Entrypoint, Cmd: image.Command,
			WorkingDir: image.WorkingDir, User: image.User,
		},
		Rootfs: struct {
			Type    string   `json:"type"`
			DiffIDs []string `json:"diff_ids"`
		}{Type: "layers", DiffIDs: []string{"sha256:" + layerDigest}},
	})
	if err != nil {
		return fmt.Errorf("生成镜像配置失败: %w", err)
	}
	configDigest := sha256.Sum256(configData)
	configName := hex.EncodeToString(configDigest[:]) + ".json"
	repoTags := append([]string(nil), image.Tags...)
	if len(repoTags) == 0 && image.Name != "" {
		repoTags = []string{image.Name}
	}
	repositories := make(map[string]map[string]string)
	for _, repoTag := range repoTags {
		if strings.Contains(repoTag, "@") {
			continue
		}
		slash := strings.LastIndex(repoTag, "/")
		colon := strings.LastIndex(repoTag, ":")
		repository, tag := repoTag, "latest"
		if colon > slash {
			repository, tag = repoTag[:colon], repoTag[colon+1:]
		}
		if repository == "" || tag == "" {
			continue
		}
		if repositories[repository] == nil {
			repositories[repository] = make(map[string]string)
		}
		repositories[repository][tag] = layerDigest
	}
	manifestData, err := json.Marshal([]dockerManifest{{
		Config:   configName,
		RepoTags: repoTags,
		Layers:   []string{"layer/layer.tar"},
	}})
	if err != nil {
		return fmt.Errorf("生成镜像清单失败: %w", err)
	}
	repositoriesData, err := json.Marshal(repositories)
	if err != nil {
		return fmt.Errorf("生成镜像仓库索引失败: %w", err)
	}
	layerMetadata, err := json.Marshal(dockerLayerMetadata{
		ID:      layerDigest,
		Created: time.Unix(image.CreatedAt, 0).UTC().Format(time.RFC3339Nano),
		Config: struct {
			Env        []string `json:"Env,omitempty"`
			Entrypoint []string `json:"Entrypoint,omitempty"`
			Cmd        []string `json:"Cmd,omitempty"`
			WorkingDir string   `json:"WorkingDir,omitempty"`
			User       string   `json:"User,omitempty"`
		}{
			Env: image.Env, Entrypoint: image.Entrypoint, Cmd: image.Command,
			WorkingDir: image.WorkingDir, User: image.User,
		},
	})
	if err != nil {
		return fmt.Errorf("生成镜像层元数据失败: %w", err)
	}

	output, err := os.CreateTemp(parent, ".container-export-*.tmp")
	if err != nil {
		return fmt.Errorf("创建导出临时文件失败: %w", err)
	}
	outputPath := output.Name()
	defer os.Remove(outputPath)
	if err := m.writeSave(output, configName, configData, manifestData, repositoriesData, layerMetadata, layerPath, layerInfo.Size()); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		return fmt.Errorf("同步导出文件失败: %w", err)
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("关闭导出文件失败: %w", err)
	}
	if err := os.Rename(outputPath, absDestination); err != nil {
		return fmt.Errorf("保存导出文件失败: %w", err)
	}
	return nil
}

// commitInstance 将实例当前 rootfs 复制到镜像存储，再写入新的镜像元数据。
func (m *Manager) commitInstance(instance *Instance, name string, tags []string) (*Image, error) {
	if instance == nil {
		return nil, errors.New("实例不能为空")
	}
	if instance.Status != StatusRunning && instance.Status != StatusStarting {
		if instance.WritableLayer {
			return nil, errors.New("可写实例未运行，rootfs 尚未挂载，无法创建快照")
		}
	}
	m.mu.RLock()
	base := m.images[instance.ImageID]
	if base != nil {
		base = m.cloneImage(base)
	}
	m.mu.RUnlock()
	if base == nil {
		return nil, fmt.Errorf("实例基础镜像不存在: %s", instance.ImageID)
	}
	sourceRootfs := instance.Rootfs
	if sourceRootfs == "" {
		sourceRootfs = base.Rootfs
	}
	if info, err := os.Lstat(sourceRootfs); err != nil {
		return nil, fmt.Errorf("读取实例 rootfs 失败: %w", err)
	} else if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("实例 rootfs 不是安全目录")
	}

	m.imageMu.Lock()
	defer m.imageMu.Unlock()
	stage, err := os.MkdirTemp(m.imagesRoot, ".commit-")
	if err != nil {
		return nil, fmt.Errorf("创建镜像快照临时目录失败: %w", err)
	}
	defer os.RemoveAll(stage)
	stageRootfs := filepath.Join(stage, "rootfs")
	if err := m.ensureManagedDirectory(stageRootfs); err != nil {
		return nil, err
	}
	if err := m.copyRootfsSnapshot(sourceRootfs, base.Rootfs, stageRootfs, instance.Mounts); err != nil {
		return nil, fmt.Errorf("复制实例 rootfs 失败: %w", err)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = strings.TrimSpace(instance.Name)
	}
	if name == "" {
		name = base.Name
	}
	if name == "" {
		name = "committed-" + instance.ID
	}
	normalizedTags := make([]string, 0, len(tags))
	seenTags := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, exists := seenTags[tag]; exists {
			continue
		}
		seenTags[tag] = struct{}{}
		normalizedTags = append(normalizedTags, tag)
	}
	image := &Image{
		Name:            name,
		Tags:            normalizedTags,
		OS:              base.OS,
		Architecture:    base.Architecture,
		User:            base.User,
		Entrypoint:      append([]string(nil), base.Entrypoint...),
		Command:         append([]string(nil), base.Command...),
		Env:             append([]string(nil), base.Env...),
		WorkingDir:      base.WorkingDir,
		CreatedAt:       time.Now().Unix(),
		SecurityVersion: m.security.Version,
		UIDMapStart:     m.security.UIDMapStart,
		GIDMapStart:     m.security.GIDMapStart,
		IDMapSize:       m.security.IDMapSize,
	}
	if len(instance.Command) > 0 {
		image.Command = append([]string(nil), instance.Command...)
	}
	if len(instance.Env) > 0 {
		image.Env = m.mergeEnv(image.Env, instance.Env)
	}
	if instance.WorkingDir != "" {
		image.WorkingDir = instance.WorkingDir
	}
	if len(image.Tags) == 0 && strings.Contains(name, ":") {
		image.Tags = []string{name}
	}
	configData, err := json.Marshal(struct {
		Name         string   `json:"name"`
		Tags         []string `json:"tags"`
		OS           string   `json:"os"`
		Architecture string   `json:"architecture"`
		User         string   `json:"user"`
		Entrypoint   []string `json:"entrypoint"`
		Command      []string `json:"command"`
		Env          []string `json:"env"`
		WorkingDir   string   `json:"working_dir"`
		CreatedAt    int64    `json:"created_at"`
	}{
		Name: image.Name, Tags: image.Tags, OS: image.OS, Architecture: image.Architecture,
		User: image.User, Entrypoint: image.Entrypoint, Command: image.Command,
		Env: image.Env, WorkingDir: image.WorkingDir, CreatedAt: image.CreatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("生成镜像摘要输入失败: %w", err)
	}
	hash := sha256.New()
	if _, err := hash.Write(configData); err != nil {
		return nil, err
	}
	if err := m.writeRootfsTar(hash, stageRootfs); err != nil {
		return nil, fmt.Errorf("计算镜像摘要失败: %w", err)
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	image.ID = "sha256:" + digest
	image.Rootfs = filepath.Join(m.imagesRoot, digest)
	if err := m.ensureInside(m.imagesRoot, image.Rootfs); err != nil {
		return nil, fmt.Errorf("镜像 rootfs 路径不安全: %w", err)
	}
	if _, err := os.Lstat(image.Rootfs); err == nil {
		return nil, fmt.Errorf("镜像已存在: %s", image.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("检查镜像 rootfs 失败: %w", err)
	}
	if err := os.Rename(stageRootfs, image.Rootfs); err != nil {
		return nil, fmt.Errorf("保存镜像 rootfs 失败: %w", err)
	}
	metadataPath := filepath.Join(m.imagesRoot, digest+".json")
	if err := m.writeJSON(metadataPath, image); err != nil {
		_ = os.RemoveAll(image.Rootfs)
		return nil, fmt.Errorf("保存镜像元数据失败: %w", err)
	}
	m.mu.Lock()
	m.images[image.ID] = m.cloneImage(image)
	m.mu.Unlock()
	return m.cloneImage(image), nil
}

func (m *Manager) writeSave(output io.Writer, configName string, configData, manifestData, repositoriesData, layerMetadata []byte, layerPath string, layerSize int64) error {
	writer := tar.NewWriter(output)
	writeBytes := func(name string, data []byte) error {
		if err := writer.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			return err
		}
		_, err := writer.Write(data)
		return err
	}
	if err := writeBytes(configName, configData); err != nil {
		return fmt.Errorf("写入镜像配置失败: %w", err)
	}
	if err := writeBytes("manifest.json", manifestData); err != nil {
		return fmt.Errorf("写入镜像清单失败: %w", err)
	}
	if err := writeBytes("repositories", repositoriesData); err != nil {
		return fmt.Errorf("写入镜像仓库索引失败: %w", err)
	}
	if err := writer.WriteHeader(&tar.Header{Name: "layer", Mode: 0755, Typeflag: tar.TypeDir}); err != nil {
		return err
	}
	if err := writeBytes("layer/VERSION", []byte("1.0\n")); err != nil {
		return fmt.Errorf("写入镜像层版本失败: %w", err)
	}
	if err := writeBytes("layer/json", layerMetadata); err != nil {
		return fmt.Errorf("写入镜像层元数据失败: %w", err)
	}
	layerHeader := &tar.Header{Name: "layer/layer.tar", Mode: 0644, Size: layerSize, Typeflag: tar.TypeReg}
	if err := writer.WriteHeader(layerHeader); err != nil {
		return err
	}
	layer, err := os.Open(layerPath)
	if err != nil {
		return fmt.Errorf("读取镜像层失败: %w", err)
	}
	_, copyErr := io.Copy(writer, layer)
	closeErr := layer.Close()
	if copyErr != nil {
		return fmt.Errorf("写入镜像层失败: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("关闭镜像层失败: %w", closeErr)
	}
	return writer.Close()
}

func (m *Manager) writeRootfsTar(output io.Writer, rootfs string) error {
	writer := tar.NewWriter(output)
	hardlinks := make(map[string]string)
	if err := m.writeRootfsTarEntry(writer, rootfs, ".", hardlinks); err != nil {
		return err
	}
	return writer.Close()
}

func (m *Manager) writeRootfsTarEntry(writer *tar.Writer, source, name string, hardlinks map[string]string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(name)
	if uid, gid, ok := archiveOwnership(info); ok {
		uid, gid, err = m.unmapOwnership(uid, gid)
		if err != nil {
			return fmt.Errorf("导出 rootfs 所有者失败 %s: %w", name, err)
		}
		header.Uid, header.Gid = uid, gid
	}
	if info.Mode()&os.ModeSymlink != 0 {
		header.Linkname, err = os.Readlink(source)
		if err != nil {
			return err
		}
	}
	if !info.Mode().IsRegular() && !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("rootfs 包含暂不支持的特殊文件: %s", name)
	}
	identity := ""
	if info.Mode().IsRegular() {
		identity = archiveFileIdentity(info)
		if previous, exists := hardlinks[identity]; identity != "" && exists {
			header.Typeflag = tar.TypeLink
			header.Linkname = previous
			header.Size = 0
		}
	}
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	if info.Mode().IsRegular() && header.Typeflag != tar.TypeLink {
		if identity != "" {
			hardlinks[identity] = header.Name
		}
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if !info.IsDir() {
		return nil
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		childName := filepath.Join(name, entry.Name())
		if err := m.writeRootfsTarEntry(writer, filepath.Join(source, entry.Name()), childName, hardlinks); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) copyRootfsSnapshot(source, lower, destination string, mounts []Mount) error {
	normalizedMounts := make([]string, 0, len(mounts))
	seenMounts := make(map[string]struct{}, len(mounts))
	hardlinks := make(map[string]string)
	for _, mount := range mounts {
		value := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(mount.Destination)), "/")
		if value == "" || value == "." {
			continue
		}
		value = filepath.FromSlash(value)
		if _, exists := seenMounts[value]; exists {
			continue
		}
		seenMounts[value] = struct{}{}
		normalizedMounts = append(normalizedMounts, value)
	}
	sort.Slice(normalizedMounts, func(i, j int) bool {
		return strings.Count(normalizedMounts[i], string(filepath.Separator)) < strings.Count(normalizedMounts[j], string(filepath.Separator))
	})
	return m.copySnapshotPath(source, destination, lower, ".", normalizedMounts, hardlinks)
}

func (m *Manager) copySnapshotPath(source, destination, lower, relative string, mounts []string, hardlinks map[string]string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	switch {
	case info.IsDir():
		if err := os.MkdirAll(destination, 0755); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			childRelative := entry.Name()
			if relative != "." {
				childRelative = filepath.Join(relative, entry.Name())
			}
			mount := ""
			for _, candidate := range mounts {
				if childRelative == candidate || strings.HasPrefix(childRelative, candidate+string(filepath.Separator)) {
					mount = candidate
					break
				}
			}
			if mount == childRelative {
				lowerPath := filepath.Join(lower, filepath.FromSlash(mount))
				if _, lowerErr := os.Lstat(lowerPath); lowerErr == nil {
					if err := m.copySnapshotPath(lowerPath, filepath.Join(destination, entry.Name()), "", mount, nil, hardlinks); err != nil {
						return err
					}
				}
				continue
			}
			if err := m.copySnapshotPath(filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name()), lower, childRelative, mounts, hardlinks); err != nil {
				return err
			}
		}
	case info.Mode()&os.ModeSymlink != 0:
		link, err := os.Readlink(source)
		if err != nil {
			return err
		}
		if err := os.Symlink(link, destination); err != nil {
			return err
		}
	case info.Mode().IsRegular():
		identity := archiveFileIdentity(info)
		linked := false
		if previous, exists := hardlinks[identity]; identity != "" && exists {
			if err := os.Link(previous, destination); err == nil {
				linked = true
			}
		}
		if linked {
			break
		}
		input, err := os.Open(source)
		if err != nil {
			return err
		}
		output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, archiveMode(info))
		if err != nil {
			_ = input.Close()
			return err
		}
		_, copyErr := io.Copy(output, input)
		closeOutputErr := output.Close()
		closeInputErr := input.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeOutputErr != nil {
			return closeOutputErr
		}
		if closeInputErr != nil {
			return closeInputErr
		}
		if identity != "" {
			hardlinks[identity] = destination
		}
	default:
		return fmt.Errorf("rootfs 包含暂不支持的特殊文件: %s", relative)
	}
	if err := archiveChown(destination, info); err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		if err := os.Chmod(destination, archiveMode(info)); err != nil {
			return err
		}
		if err := os.Chtimes(destination, info.ModTime(), info.ModTime()); err != nil {
			return err
		}
	}
	return nil
}

func archiveMode(info os.FileInfo) os.FileMode {
	mode := info.Mode().Perm()
	if info.Mode()&os.ModeSetuid != 0 {
		mode |= os.ModeSetuid
	}
	if info.Mode()&os.ModeSetgid != 0 {
		mode |= os.ModeSetgid
	}
	if info.Mode()&os.ModeSticky != 0 {
		mode |= os.ModeSticky
	}
	return mode
}

func archiveStatUint(info os.FileInfo, name string) (uint64, bool) {
	if info == nil || info.Sys() == nil {
		return 0, false
	}
	value := reflect.ValueOf(info.Sys())
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return 0, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return 0, false
	}
	field := value.FieldByName(name)
	if !field.IsValid() {
		return 0, false
	}
	switch field.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() >= 0 {
			return uint64(field.Int()), true
		}
	}
	return 0, false
}

func archiveOwnership(info os.FileInfo) (int, int, bool) {
	uid, uidOK := archiveStatUint(info, "Uid")
	gid, gidOK := archiveStatUint(info, "Gid")
	if !uidOK || !gidOK || uint64(int(uid)) != uid || uint64(int(gid)) != gid {
		return 0, 0, false
	}
	return int(uid), int(gid), true
}

func archiveFileIdentity(info os.FileInfo) string {
	dev, devOK := archiveStatUint(info, "Dev")
	ino, inoOK := archiveStatUint(info, "Ino")
	if !devOK || !inoOK || ino == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", dev, ino)
}

func archiveChown(path string, info os.FileInfo) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	uid, gid, ok := archiveOwnership(info)
	if !ok {
		return nil
	}
	var err error
	if info.Mode()&os.ModeSymlink != 0 {
		err = os.Lchown(path, uid, gid)
	} else {
		err = os.Chown(path, uid, gid)
	}
	if err != nil {
		return fmt.Errorf("恢复文件所有者失败 %s: %w", path, err)
	}
	return nil
}
