package timezone

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// withWindowsTimezonePrivilege 串行启用时区特权，执行后恢复原状态。
// 特权属于进程，mutex 属于线程；锁定线程可确保等待和释放使用同一线程。
func withWindowsTimezonePrivilege(adjust *windows.Proc, apply func() error) (resultErr error) {
	mutexName, _ := windows.UTF16PtrFromString(fmt.Sprintf(`Local\sinking-panel.timezone.%d`, os.Getpid()))
	mutex, err := windows.CreateMutexEx(nil, mutexName, 0, windows.SYNCHRONIZE|windows.MUTEX_MODIFY_STATE)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return fmt.Errorf("创建 Windows 时区锁失败: %w", err)
	}
	defer windows.CloseHandle(mutex)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	event, err := windows.WaitForSingleObject(mutex, 5000)
	if err != nil {
		return fmt.Errorf("等待 Windows 时区锁失败: %w", err)
	}
	if event != windows.WAIT_OBJECT_0 && event != windows.WAIT_ABANDONED {
		return errors.New("系统时区正在修改，请稍后重试")
	}
	defer windows.ReleaseMutex(mutex)
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token); err != nil {
		return fmt.Errorf("打开 Windows 进程令牌失败: %w", err)
	}
	defer token.Close()
	privilege := windows.Tokenprivileges{PrivilegeCount: 1}
	privilege.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED
	privilegeName, _ := windows.UTF16PtrFromString("SeTimeZonePrivilege")
	if err := windows.LookupPrivilegeValue(nil, privilegeName, &privilege.Privileges[0].Luid); err != nil {
		return fmt.Errorf("查询 Windows 时区特权失败: %w", err)
	}
	var previous windows.Tokenprivileges
	var length uint32
	result, _, err := adjust.Call(uintptr(token), 0, uintptr(unsafe.Pointer(&privilege)), unsafe.Sizeof(previous), uintptr(unsafe.Pointer(&previous)), uintptr(unsafe.Pointer(&length)))
	// 返回成功也可能没有授予特权，需同时检查 ERROR_NOT_ALL_ASSIGNED。
	if result == 0 || err != windows.ERROR_SUCCESS {
		return fmt.Errorf("启用 Windows 时区特权失败: %w", err)
	}
	defer func() {
		result, _, err := adjust.Call(uintptr(token), 0, uintptr(unsafe.Pointer(&previous)), 0, 0, 0)
		if result == 0 || err != windows.ERROR_SUCCESS {
			resultErr = errors.Join(resultErr, fmt.Errorf("恢复 Windows 时区特权失败: %w", err))
		}
	}()
	return apply()
}
