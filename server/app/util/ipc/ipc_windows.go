package ipc

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/tailscale/go-winio"
	"golang.org/x/sys/windows"
)

type pipeListener struct {
	net.Listener
	user       *windows.SID
	privileged bool
}

// Accept 丢弃未授权连接，避免单个非法客户端中断整个通信服务。
func (l *pipeListener) Accept() (net.Conn, error) {
	for {
		connection, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		if err := checkPipePeer(connection, l.user, l.privileged, false); err == nil {
			return connection, nil
		}
		_ = connection.Close()
	}
}

func pipeName(path string) (string, error) {
	path, err := endpoint(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`\\.\pipe\sinking-panel-%x`, sha256.Sum256([]byte(strings.ToLower(path)))), nil
}

// Listen 创建本机命名管道；提权服务只允许 SYSTEM 和已提权管理员访问。
func Listen(path string) (net.Listener, error) {
	name, err := pipeName(path)
	if err != nil {
		return nil, err
	}
	user, privileged, err := pipeIdentity(windows.GetCurrentProcessToken())
	if err != nil {
		return nil, err
	}
	owner := user.String()
	dacl := "D:P(A;;GA;;;SY)(A;;GA;;;BA)"
	if privileged {
		// 管理员组作为属主，避免同账户未提权进程利用属主权限修改 DACL。
		owner = "BA"
		if user.IsWellKnown(windows.WinLocalSystemSid) {
			owner = "SY"
		}
	} else {
		dacl += "(A;;GA;;;" + owner + ")"
	}
	listener, err := winio.ListenPipe(name, &winio.PipeConfig{
		SecurityDescriptor: "O:" + owner + dacl,
	})
	if err != nil {
		return nil, err
	}
	return &pipeListener{
		Listener:   listener,
		user:       user,
		privileged: privileged,
	}, nil
}

// Dial 连接本机管道；占用和拒绝访问均不能按面板已停止处理。
func Dial(ctx context.Context, path string) (net.Conn, error) {
	name, err := pipeName(path)
	if err != nil {
		return nil, err
	}
	connection, err := winio.DialPipeContext(ctx, name)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
		return nil, ErrUnavailable
	}
	if err != nil {
		return nil, err
	}
	user, privileged, err := pipeIdentity(windows.GetCurrentProcessToken())
	if err == nil {
		err = checkPipePeer(connection, user, privileged, true)
	}
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("本机通信身份验证失败: %w", err)
	}
	return connection, nil
}

// pipeIdentity 读取真实账户和提权状态，不将 UAC 过滤后的管理员令牌视为管理员。
func pipeIdentity(token windows.Token) (*windows.SID, bool, error) {
	user, err := token.GetTokenUser()
	if err != nil {
		return nil, false, err
	}
	if user.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
		return user.User.Sid, true, nil
	}
	groups, err := token.GetTokenGroups()
	if err != nil {
		return nil, false, err
	}
	for _, group := range groups.AllGroups() {
		if group.Sid.IsWellKnown(windows.WinBuiltinAdministratorsSid) &&
			group.Attributes&windows.SE_GROUP_ENABLED != 0 &&
			group.Attributes&windows.SE_GROUP_USE_FOR_DENY_ONLY == 0 {
			if !token.IsElevated() {
				return nil, false, errors.New("无法确认本机通信进程的管理员提权状态")
			}
			return user.User.Sid, true, nil
		}
	}
	return user.User.Sid, false, nil
}

// checkPipePeer 以已连接管道的内核 PID 验证对端，不能信任请求中自行声明的身份。
func checkPipePeer(connection net.Conn, user *windows.SID, privileged, server bool) error {
	file, ok := connection.(interface{ Fd() uintptr })
	if !ok {
		return errors.New("无法读取本机通信句柄")
	}
	handle := windows.Handle(file.Fd())
	getPID := windows.GetNamedPipeClientProcessId
	if server {
		getPID = windows.GetNamedPipeServerProcessId
	}
	var pid uint32
	if err := getPID(handle, &pid); err != nil {
		return err
	}
	if pid == 0 {
		return errors.New("本机通信对端进程无效")
	}
	process, token, err := pipeProcess(pid, privileged)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(process)
	defer token.Close()
	peer, peerPrivileged, err := pipeIdentity(token)
	if err != nil {
		return err
	}
	if !peerPrivileged && (privileged || !user.Equals(peer)) {
		return errors.New("本机通信对端账户或提权状态不受信任")
	}
	if server {
		// 管道名可被提前占用，校验实际句柄的属主，不能只检查名称或路径。
		descriptor, err := windows.GetSecurityInfo(handle, windows.SE_KERNEL_OBJECT, windows.OWNER_SECURITY_INFORMATION)
		if err != nil {
			return err
		}
		owner, _, err := descriptor.Owner()
		if err != nil {
			return err
		}
		if owner == nil || (!owner.Equals(peer) && !(peerPrivileged &&
			(owner.IsWellKnown(windows.WinBuiltinAdministratorsSid) || owner.IsWellKnown(windows.WinLocalSystemSid)))) {
			return errors.New("本机通信管道属主与服务进程身份不匹配")
		}
	}
	// 持有进程句柄并复核身份，拒绝验证期间已经退出或变化的对端。
	var currentPID, exitCode uint32
	if err := getPID(handle, &currentPID); err != nil {
		return err
	}
	if err := windows.GetExitCodeProcess(process, &exitCode); err != nil {
		return err
	}
	if currentPID != pid || exitCode != 259 { // STILL_ACTIVE
		return errors.New("本机通信对端进程已经退出或变化")
	}
	return nil
}

// pipeProcess 只申请查询权限；管理员访问 SYSTEM 令牌时临时在独立线程启用调试权限。
func pipeProcess(pid uint32, privileged bool) (windows.Handle, windows.Token, error) {
	open := func() (windows.Handle, windows.Token, error) {
		process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			return 0, 0, err
		}
		var token windows.Token
		if err := windows.OpenProcessToken(process, windows.TOKEN_QUERY, &token); err != nil {
			_ = windows.CloseHandle(process)
			return 0, 0, err
		}
		return process, token, nil
	}
	process, token, err := open()
	if !privileged || !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return process, token, err
	}
	// 独立 goroutine 不携带调用方的线程令牌；还原失败时锁定线程随 goroutine 销毁。
	done := make(chan struct{})
	go func() {
		defer close(done)
		runtime.LockOSThread()
		if err = windows.ImpersonateSelf(windows.SecurityImpersonation); err != nil {
			runtime.UnlockOSThread()
			return
		}
		defer func() {
			if windows.RevertToSelf() == nil {
				runtime.UnlockOSThread()
			}
		}()
		var current windows.Token
		if err = windows.OpenThreadToken(windows.CurrentThread(), windows.TOKEN_QUERY|windows.TOKEN_ADJUST_PRIVILEGES, true, &current); err != nil {
			return
		}
		defer current.Close()
		name, _ := windows.UTF16PtrFromString("SeDebugPrivilege")
		privileges := windows.Tokenprivileges{PrivilegeCount: 1}
		if err = windows.LookupPrivilegeValue(nil, name, &privileges.Privileges[0].Luid); err != nil {
			return
		}
		privileges.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED
		if err = windows.AdjustTokenPrivileges(current, false, &privileges, 0, nil, nil); err != nil {
			return
		}
		process, token, err = open()
	}()
	<-done
	return process, token, err
}
