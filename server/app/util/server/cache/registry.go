package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/darkweak/storages/core"
)

var (
	errStoreClosed = errors.New("simplefs 存储已关闭")
	// ErrStoreInUse 表示仍有已加载的 HTTP 路由正在使用该目录。
	ErrStoreInUse = errors.New("simplefs 存储仍在使用")
)

// StoreRegistry 管理进程内共享的站点缓存数据库。
type StoreRegistry struct {
	sync.Mutex
	items    map[string]*store
	backends map[string]*backend
	modules  map[string][]*SimpleFS
}

// Registry 是进程级缓存注册表。
var Registry = &StoreRegistry{
	items:    map[string]*store{},
	backends: map[string]*backend{},
	modules:  map[string][]*SimpleFS{},
}

type store struct {
	backend *backend
	key     string
	maxSize int64
	stale   time.Duration
	refs    int
}

func (r *StoreRegistry) acquire(path string, maxSize int64, stale time.Duration) *store {
	canonical, err := r.canonical(path)
	if err != nil {
		canonical = filepath.Clean(path)
	}
	r.Lock()
	defer r.Unlock()
	key := fmt.Sprintf("%s\x00%d\x00%d", canonical, maxSize, stale)
	if current := r.items[key]; current != nil {
		current.refs++
		return current
	}
	currentBackend := r.backends[canonical]
	if currentBackend == nil {
		currentBackend = &backend{path: canonical}
		r.backends[canonical] = currentBackend
	}
	currentBackend.refs++
	current := &store{backend: currentBackend, key: key, maxSize: maxSize, stale: stale, refs: 1}
	r.items[key] = current
	return current
}

func (r *StoreRegistry) release(current *store) error {
	if current == nil {
		return nil
	}
	r.Lock()
	defer r.Unlock()
	registered := r.items[current.key]
	if registered != current || current.refs == 0 {
		return nil
	}
	current.refs--
	if current.refs > 0 {
		return nil
	}
	delete(r.items, current.key)
	current.backend.refs--
	if current.backend.refs > 0 {
		return nil
	}
	delete(r.backends, current.backend.path)
	return current.backend.close()
}

func (r *StoreRegistry) register(module *SimpleFS) {
	r.Lock()
	r.modules[module.uuid] = append(r.modules[module.uuid], module)
	r.Unlock()
}

func (r *StoreRegistry) unregister(module *SimpleFS) {
	r.Lock()
	registered := r.modules[module.uuid]
	for index, current := range registered {
		if current == module {
			registered = append(registered[:index], registered[index+1:]...)
			break
		}
	}
	if len(registered) == 0 {
		delete(r.modules, module.uuid)
		r.Unlock()
		return
	}
	r.modules[module.uuid] = registered
	active := registered[len(registered)-1]
	r.Unlock()
	core.RegisterStorage(active)
}

// CloseAll 关闭进程内所有缓存数据库。应在 HTTP 服务停止接收请求后调用。
func (r *StoreRegistry) CloseAll() error {
	r.Lock()
	defer r.Unlock()
	var result error
	for _, current := range r.backends {
		result = errors.Join(result, current.close())
	}
	r.items = map[string]*store{}
	r.backends = map[string]*backend{}
	r.modules = map[string][]*SimpleFS{}
	core.ResetRegisteredStorages()
	return result
}

// PurgeTree 清空受控根目录内的缓存，运行中的存储保持可用。
func (r *StoreRegistry) PurgeTree(allowedRoot, path string) error {
	root, err := r.treePath(allowedRoot, path)
	if err != nil {
		return err
	}
	r.Lock()
	defer r.Unlock()
	activeFiles := map[string]struct{}{}
	for currentPath, current := range r.backends {
		if !r.inside(root, currentPath) {
			continue
		}
		activeFiles[filepath.Join(currentPath, databaseName)] = struct{}{}
		if err = current.reset(); err != nil {
			return fmt.Errorf("清空 simplefs 数据库失败: %w", err)
		}
	}
	err = filepath.WalkDir(root, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, os.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if _, active := activeFiles[currentPath]; active {
			return nil
		}
		return os.Remove(currentPath)
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// RemoveTree 在站点路由卸载后关闭并删除受控根目录内的缓存目录。
func (r *StoreRegistry) RemoveTree(allowedRoot, path string) error {
	root, err := r.treePath(allowedRoot, path)
	if err != nil {
		return err
	}
	r.Lock()
	defer r.Unlock()
	for currentPath, current := range r.backends {
		if r.inside(root, currentPath) && current.refs > 0 {
			return ErrStoreInUse
		}
	}
	for currentPath, current := range r.backends {
		if !r.inside(root, currentPath) {
			continue
		}
		if err = current.close(); err != nil {
			return err
		}
		delete(r.backends, currentPath)
	}
	if err = os.RemoveAll(root); err != nil {
		return fmt.Errorf("删除 simplefs 缓存目录失败: %w", err)
	}
	return nil
}

func (r *StoreRegistry) treePath(allowedRoot, path string) (string, error) {
	root, err := r.canonical(allowedRoot)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("解析缓存目录失败: %w", err)
	}
	parent, err := r.canonical(filepath.Dir(target))
	if err != nil {
		return "", err
	}
	target = filepath.Join(parent, filepath.Base(target))
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", errors.New("缓存目录超出受控根目录")
	}
	if info, statErr := os.Lstat(target); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("缓存目录不能是符号链接")
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return "", fmt.Errorf("读取缓存目录失败: %w", statErr)
	}
	resolved, err := r.canonical(target)
	if err != nil {
		return "", err
	}
	if !r.inside(root, resolved) || resolved == root {
		return "", errors.New("缓存目录通过符号链接逃逸受控根目录")
	}
	return target, nil
}

func (r *StoreRegistry) canonical(path string) (string, error) {
	if path == "" {
		return "", errors.New("simplefs 路径不能为空")
	}
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("解析 simplefs 路径失败: %w", err)
	}
	current := absolute
	missing := make([]string, 0, 4)
	for {
		resolved, resolveErr := filepath.EvalSymlinks(current)
		if resolveErr == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return filepath.Clean(resolved), nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return filepath.Clean(absolute), nil
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func (*StoreRegistry) inside(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && (relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative))
}
