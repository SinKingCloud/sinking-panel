package plugin

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/traefik/yaegi/interp"
)

// Manager 管理插件包与解释器。相同目录串行调用，不同插件互不阻塞。
type Manager struct {
	root      string
	options   Options
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	instances map[string]*instance
	imports   sync.WaitGroup
}

// instance 的全部运行状态由 gate 保护；卸载后只保留轻量锁，避免删除与重新加载竞态。
type instance struct {
	gate        chan struct{}
	manifest    *Manifest
	plugin      *Plugin
	interpreter *interp.Interpreter
	root        *os.Root
	runCancel   context.CancelFunc
	runContext  context.Context
}

// NewManager 创建插件管理器，不读取或执行任何插件代码。
func NewManager(root string, options Options) *Manager {
	if options.ManifestName == "" {
		options.ManifestName = "plugin.json"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		root:      root,
		options:   options,
		ctx:       ctx,
		cancel:    cancel,
		instances: make(map[string]*instance),
	}
}

// acquire 使用可取消的等待队列串行化同一插件的加载、调用和文件删除。
func (m *Manager) acquire(ctx context.Context, directory string) (*instance, func(), error) {
	if !filepath.IsLocal(directory) || directory == "." || strings.ContainsAny(directory, `\:`) || filepath.ToSlash(filepath.Clean(directory)) != directory {
		return nil, nil, errors.New("插件目录必须是包内相对路径")
	}
	m.mu.Lock()
	if m.ctx.Err() != nil {
		m.mu.Unlock()
		return nil, nil, errors.New("插件管理器已关闭")
	}
	i := m.instances[directory]
	if i == nil {
		i = &instance{gate: make(chan struct{}, 1)}
		m.instances[directory] = i
	}
	m.mu.Unlock()
	select {
	case i.gate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-i.gate
			return nil, nil, err
		}
		if m.ctx.Err() != nil {
			<-i.gate
			return nil, nil, errors.New("插件管理器已关闭")
		}
		return i, func() { <-i.gate }, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-m.ctx.Done():
		return nil, nil, errors.New("插件管理器已关闭")
	}
}

// Unload 等待当前调用结束后释放解释器，下次调用会重新读取插件代码。
func (m *Manager) Unload(directory string) {
	m.mu.Lock()
	i := m.instances[directory]
	m.mu.Unlock()
	if i == nil {
		return
	}
	i.gate <- struct{}{}
	defer func() { <-i.gate }()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := i.stop(ctx, "{}"); err != nil {
		log.Printf("停止插件 %s 失败: %v", directory, err)
		return
	}
	i.unload()
}

// Rename 移动插件目录，先释放其解释器和目录句柄，不覆盖目标目录。
func (m *Manager) Rename(from, to string) error {
	for _, directory := range []string{from, to} {
		if !filepath.IsLocal(directory) || directory == "." || strings.ContainsAny(directory, `\:`) || filepath.Base(directory) != directory {
			return errors.New("插件目录必须位于管理器根目录下")
		}
	}
	if from == to {
		return nil
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return err
	}
	defer root.Close()
	source, err := root.Lstat(from)
	if err != nil {
		return err
	}
	if !source.IsDir() {
		return errors.New("插件源路径必须是目录")
	}
	if _, err = root.Lstat(to); err == nil {
		return &os.PathError{
			Op:   "rename",
			Path: to,
			Err:  os.ErrExist,
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	ordered := []string{from, to}
	slices.Sort(ordered)
	instances := make([]*instance, 0, len(ordered))
	for _, name := range ordered {
		i, release, err := m.acquire(context.Background(), name)
		if err != nil {
			return err
		}
		defer release()
		instances = append(instances, i)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, i := range instances {
		if _, err = i.stop(ctx, "{}"); err != nil {
			return err
		}
		i.unload()
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// 目录切换期间不接收新的加载请求。
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ctx.Err() != nil {
		return errors.New("插件管理器已关闭")
	}
	if _, err = root.Lstat(to); err == nil {
		return &os.PathError{
			Op:   "rename",
			Path: to,
			Err:  os.ErrExist,
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return root.Rename(from, to)
}

// Remove 停止并删除插件代码目录，不删除目录之外的业务数据。
func (m *Manager) Remove(directory string) error {
	i, release, err := m.acquire(context.Background(), directory)
	if err != nil {
		return err
	}
	defer release()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := i.stop(ctx, "{}"); err != nil {
		return err
	}
	i.unload()
	root, err := os.OpenRoot(m.root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer root.Close()
	return root.RemoveAll(directory)
}

// Close 停止接收调用，取消正在执行的方法，并等待其退出后释放解释器。
func (m *Manager) Close() {
	m.mu.Lock()
	m.cancel()
	instances := make([]*instance, 0, len(m.instances))
	for _, i := range m.instances {
		instances = append(instances, i)
	}
	m.mu.Unlock()
	m.imports.Wait()
	var wait sync.WaitGroup
	for _, i := range instances {
		wait.Add(1)
		go func() {
			defer wait.Done()
			i.gate <- struct{}{}
			defer func() { <-i.gate }()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if _, err := i.stop(ctx, "{}"); err != nil {
				log.Printf("停止插件 %s 失败: %v", i.manifest.Id, err)
			}
			i.unload()
		}()
	}
	wait.Wait()
	m.mu.Lock()
	clear(m.instances)
	m.mu.Unlock()
}

func (i *instance) unload() {
	if i.runCancel != nil {
		i.runCancel()
	}
	i.runCancel, i.runContext = nil, nil
	i.manifest, i.plugin, i.interpreter = nil, nil, nil
	if i.root != nil {
		_ = i.root.Close()
		i.root = nil
	}
}
