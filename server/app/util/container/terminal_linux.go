//go:build linux && cgo

package container

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/containerd/console"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/utils"
	"golang.org/x/sys/unix"
)

type linuxTerminalSession struct {
	console      console.Console
	process      *libcontainer.Process
	closeOnce    sync.Once
	consoleOnce  sync.Once
	consoleError error
}

func (s *linuxTerminalSession) Read(payload []byte) (int, error) {
	return s.console.Read(payload)
}

func (s *linuxTerminalSession) Write(payload []byte) (int, error) {
	return s.console.Write(payload)
}

func (s *linuxTerminalSession) Resize(height, width int) error {
	if height < 1 || height > 1000 || width < 1 || width > 1000 {
		return errors.New("终端尺寸必须在 1-1000 范围内")
	}
	return s.console.Resize(console.WinSize{Height: uint16(height), Width: uint16(width)})
}

func (s *linuxTerminalSession) Close() error {
	var signalErr error
	s.closeOnce.Do(func() {
		if err := s.process.Signal(unix.SIGKILL); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) && !errors.Is(err, os.ErrProcessDone) {
			signalErr = err
		}
		s.closeConsole()
	})
	return errors.Join(signalErr, s.consoleError)
}

func (s *linuxTerminalSession) closeConsole() {
	s.consoleOnce.Do(func() {
		s.consoleError = s.console.Close()
	})
}

func (m *Manager) openPlatformTerminal(id string, height, width int) (TerminalSession, error) {
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	if instance == nil {
		m.mu.RUnlock()
		return nil, errors.New("实例不存在")
	}
	image := m.cloneImage(m.images[instance.ImageID])
	runtime, _ := m.runtime[id].(*linuxRuntime)
	m.mu.RUnlock()
	if image == nil {
		return nil, errors.New("实例镜像不存在")
	}

	var runtimeContainer *libcontainer.Container
	if runtime != nil && runtime.container != nil {
		runtimeContainer = runtime.container
	} else {
		loaded, err := libcontainer.Load(m.runtimeRoot, id)
		if err != nil {
			return nil, fmt.Errorf("加载容器运行状态失败: %w", err)
		}
		runtimeContainer = loaded
	}

	rootfs := instance.Rootfs
	if rootfs == "" {
		rootfs = image.Rootfs
	}
	uid, gid, groups, err := m.resolveUser(rootfs, image.User)
	if err != nil {
		return nil, fmt.Errorf("解析镜像运行用户失败: %w", err)
	}
	workingDir := instance.WorkingDir
	if workingDir == "" {
		workingDir = image.WorkingDir
	}
	if workingDir == "" {
		workingDir = "/"
	}
	env := m.mergeEnv(
		[]string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		image.Env,
		instance.Env,
		[]string{"TERM=xterm-256color", "COLORTERM=truecolor"},
	)
	parent, child, err := utils.NewSockPair("container-terminal")
	if err != nil {
		return nil, fmt.Errorf("创建容器终端通信通道失败: %w", err)
	}
	closeSockets := func() {
		_ = child.Close()
		_ = parent.Close()
	}
	timeout := unix.NsecToTimeval(m.managerOptions().TerminalConsoleTimeout.Nanoseconds())
	if err := unix.SetsockoptTimeval(int(parent.Fd()), unix.SOL_SOCKET, unix.SO_RCVTIMEO, &timeout); err != nil {
		closeSockets()
		return nil, fmt.Errorf("设置容器终端接收超时失败: %w", err)
	}
	process := &libcontainer.Process{
		Args:             []string{"/bin/sh", "-c", "if [ -x /bin/bash ]; then exec /bin/bash; fi; exec /bin/sh"},
		Env:              env,
		UID:              uid,
		GID:              gid,
		AdditionalGroups: groups,
		Cwd:              workingDir,
		ConsoleSocket:    child,
		ConsoleHeight:    uint16(height),
		ConsoleWidth:     uint16(width),
	}
	if err := runtimeContainer.Run(process); err != nil {
		closeSockets()
		return nil, fmt.Errorf("启动容器终端失败: %w", err)
	}
	_ = child.Close()
	processDone := make(chan struct{})
	go func() {
		_, _ = process.Wait()
		close(processDone)
	}()
	cleanupProcess := func() {
		_ = process.Signal(unix.SIGKILL)
		timer := time.NewTimer(m.managerOptions().TerminalCleanupTimeout)
		defer timer.Stop()
		select {
		case <-processDone:
		case <-timer.C:
		}
	}

	file, receiveErr := utils.RecvFile(parent)
	_ = parent.Close()
	if receiveErr != nil {
		cleanupProcess()
		if errors.Is(receiveErr, unix.EAGAIN) || errors.Is(receiveErr, unix.EWOULDBLOCK) {
			return nil, errors.New("等待容器终端就绪超时")
		}
		return nil, fmt.Errorf("接收容器终端失败: %w", receiveErr)
	}
	terminal, err := console.ConsoleFromFile(file)
	if err != nil {
		_ = file.Close()
		cleanupProcess()
		return nil, fmt.Errorf("打开容器终端失败: %w", err)
	}
	session := &linuxTerminalSession{console: terminal, process: process}
	if err := session.Resize(height, width); err != nil {
		_ = session.Close()
		cleanupProcess()
		return nil, fmt.Errorf("设置容器终端尺寸失败: %w", err)
	}
	return session, nil
}

var _ io.ReadWriteCloser = (*linuxTerminalSession)(nil)
