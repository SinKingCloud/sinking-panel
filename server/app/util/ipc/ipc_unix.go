//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package ipc

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type listener struct {
	*net.UnixListener
	lock    *os.File
	address string
	socket  os.FileInfo
	once    sync.Once
	err     error
}

// Accept 只交付通过内核身份校验的连接，非法连接不会中断监听。
func (l *listener) Accept() (net.Conn, error) {
	for {
		connection, err := l.AcceptUnix()
		if err != nil {
			return nil, err
		}
		uid, err := peerUID(connection)
		if err == nil && uid == uint32(os.Geteuid()) {
			return connection, nil
		}
		_ = connection.Close()
	}
}

func (l *listener) Close() error {
	l.once.Do(func() {
		l.err = l.UnixListener.Close()
		// 不删除监听期间被替换的文件，也不提前释放互斥锁。
		if info, err := os.Lstat(l.address); err == nil && os.SameFile(l.socket, info) {
			if err = os.Remove(l.address); l.err == nil {
				l.err = err
			}
		}
		_ = l.lock.Close()
	})
	return l.err
}

// directoryInfo 逐层检查再创建，避免通过可写祖先或软链接进入其他用户的目录。
// root/当前账户拥有的 sticky 目录（如 /tmp）可以作为祖先，子目录仍单独校验。
func directoryInfo(path string, create bool) (os.FileInfo, error) {
	links := 0
	var check func(string) (os.FileInfo, error)
	check = func(path string) (os.FileInfo, error) {
		if parent := filepath.Dir(path); parent != path {
			if _, err := check(parent); err != nil {
				return nil, err
			}
		}
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) && create {
			if err = os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				return nil, err
			}
			info, err = os.Lstat(path)
		}
		if err != nil {
			return nil, err
		}
		uid := info.Sys().(*syscall.Stat_t).Uid
		if uid != 0 && uid != uint32(os.Geteuid()) {
			return nil, fmt.Errorf("本机通信目录不属于可信账户: %s", path)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			links++
			if links > 255 {
				return nil, errors.New("本机通信目录软链接过多")
			}
			target, err := os.Readlink(path)
			if err != nil {
				return nil, err
			}
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(path), target)
			}
			return check(filepath.Clean(target))
		}
		if !info.IsDir() || (info.Mode().Perm()&0022 != 0 && info.Mode()&os.ModeSticky == 0) {
			return nil, fmt.Errorf("本机通信目录允许其他用户修改: %s", path)
		}
		return info, nil
	}
	return check(path)
}

func socketDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	// 先检查原路径，不能在解析软链接后遗漏它所在的不安全目录。
	if _, err = directoryInfo(absolute, false); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	directory, err := endpoint(absolute)
	if err != nil {
		return "", err
	}
	// macOS 套接字路径上限为 104 字节，较长的安装目录使用固定短路径。
	if len(filepath.Join(directory, "control.sock")) >= 100 {
		digest := sha256.Sum256([]byte(directory))
		directory = filepath.Join("/tmp", fmt.Sprintf("sinking-ipc-%d-%x", os.Geteuid(), digest[:20]))
		if _, err = directoryInfo(directory, false); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		directory, err = endpoint(directory)
		if err != nil {
			return "", err
		}
	}
	return directory, nil
}

// openLock 不跟随软链接，不接受硬链接、特殊文件或其他用户可访问的锁文件。
func openLock(directory string, create bool) (*os.File, error) {
	flags := syscall.O_RDWR | syscall.O_NOFOLLOW | syscall.O_CLOEXEC | syscall.O_NONBLOCK
	if create {
		flags |= syscall.O_CREAT
	}
	fd, err := syscall.Open(filepath.Join(directory, "control.lock"), flags, 0600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), "control.lock")
	info, err := file.Stat()
	if err == nil {
		stat := info.Sys().(*syscall.Stat_t)
		if !info.Mode().IsRegular() || stat.Uid != uint32(os.Geteuid()) || stat.Nlink != 1 || info.Mode().Perm()&0077 != 0 {
			err = errors.New("本机通信锁文件权限或类型不正确")
		}
	}
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

func socketInfo(address string) (os.FileInfo, error) {
	info, err := os.Lstat(address)
	if err == nil && (info.Mode()&os.ModeSocket == 0 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Geteuid()) || info.Mode().Perm()&0077 != 0) {
		err = errors.New("本机通信套接字权限或类型不正确")
	}
	return info, err
}

// Listen 创建私有套接字，并校验每条连接的系统账户；文件锁防止重复监听。
func Listen(path string) (net.Listener, error) {
	directory, err := socketDirectory(path)
	if err != nil {
		return nil, err
	}
	info, err := directoryInfo(directory, true)
	if err != nil {
		return nil, err
	}
	if info.Sys().(*syscall.Stat_t).Uid != uint32(os.Geteuid()) || info.Mode().Perm()&0022 != 0 {
		return nil, errors.New("本机通信目录权限不正确，请使用面板运行账户执行命令")
	}
	fd, err := syscall.Open(directory, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	dir := os.NewFile(uintptr(fd), directory)
	opened, err := dir.Stat()
	if err == nil && !os.SameFile(info, opened) {
		err = errors.New("本机通信目录已被替换")
	}
	if err == nil {
		err = dir.Chmod(0700)
	}
	_ = dir.Close()
	if err != nil {
		return nil, err
	}
	lock, err := openLock(directory, true)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("本机通信服务已运行: %w", err)
	}
	ready := false
	defer func() {
		if !ready {
			_ = lock.Close()
		}
	}()
	address := filepath.Join(directory, "control.sock")
	if _, err = socketInfo(address); err == nil {
		connection, probeErr := net.DialTimeout("unix", address, 200*time.Millisecond)
		if probeErr == nil {
			_ = connection.Close()
			return nil, errors.New("本机通信服务已运行")
		}
		if !errors.Is(probeErr, syscall.ECONNREFUSED) && !errors.Is(probeErr, os.ErrNotExist) {
			return nil, probeErr
		}
		if err = os.Remove(address); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	socket, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: address,
		Net:  "unix",
	})
	if err != nil {
		return nil, err
	}
	if err = os.Chmod(address, 0600); err != nil {
		_ = socket.Close()
		return nil, err
	}
	info, err = socketInfo(address)
	if err != nil {
		_ = socket.Close()
		return nil, err
	}
	socket.SetUnlinkOnClose(false)
	ready = true
	return &listener{
		UnixListener: socket,
		lock:         lock,
		address:      address,
		socket:       info,
	}, nil
}

// Dial 同时校验目录、套接字和服务端内核身份；认证失败不能作为离线模式处理。
func Dial(ctx context.Context, path string) (net.Conn, error) {
	directory, err := socketDirectory(path)
	if err != nil {
		return nil, err
	}
	info, err := directoryInfo(directory, false)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrUnavailable
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Geteuid()) || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("本机通信目录权限不正确，请使用面板运行账户执行命令")
	}
	address := filepath.Join(directory, "control.sock")
	if _, err = socketInfo(address); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	connection, err := (&net.Dialer{}).DialContext(ctx, "unix", address)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ECONNREFUSED) {
		// 套接字创建、关闭和文件锁释放之间仍属于在线进程，不能离线接管。
		lock, lockErr := openLock(directory, false)
		if lockErr == nil {
			lockErr = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
			_ = lock.Close()
			if lockErr != nil {
				return nil, fmt.Errorf("本机通信服务正在启动或停止，请稍后重试: %w", lockErr)
			}
		} else if !errors.Is(lockErr, os.ErrNotExist) {
			return nil, lockErr
		}
		return nil, ErrUnavailable
	}
	if err != nil {
		return nil, err
	}
	uid, err := peerUID(connection.(*net.UnixConn))
	if err == nil && uid != uint32(os.Geteuid()) {
		err = errors.New("本机通信服务端身份与当前账户不一致")
	}
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("本机通信身份验证失败: %w", err)
	}
	return connection, nil
}
