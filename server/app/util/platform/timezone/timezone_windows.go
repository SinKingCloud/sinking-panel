package timezone

import (
	"errors"
	"fmt"
	"slices"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// List 返回排序后的 Windows 原生时区标识，可直接传给 Set。
func List() ([]string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Time Zones`, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("打开 Windows 时区列表失败: %w", err)
	}
	defer key.Close()
	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("读取 Windows 时区列表失败: %w", err)
	}
	slices.Sort(names)
	return slices.Compact(names), nil
}

// Get 读取系统配置，不受进程的 TZ 环境变量或 time.Local 缓存影响。
func Get() (string, error) {
	library, err := loadWindowsLibrary("kernel32.dll")
	if err != nil {
		return "", err
	}
	defer library.Release()
	get, err := library.FindProc("GetDynamicTimeZoneInformation")
	if err != nil {
		return "", fmt.Errorf("系统不支持读取 Windows 动态时区: %w", err)
	}
	var info windowsTimezoneInformation
	result, _, err := get.Call(uintptr(unsafe.Pointer(&info)))
	if uint32(result) == 0xffffffff { // TIME_ZONE_ID_INVALID
		return "", fmt.Errorf("读取 Windows 系统时区失败: %w", err)
	}
	name := windows.UTF16ToString(info.TimeZoneKeyName[:])
	if name == "" {
		return "", errors.New("Windows 当前使用的自定义时区没有系统时区 ID")
	}
	return name, nil
}

// setTimezone 使用当前年度规则设置系统时区，并在调用后恢复进程特权。
func setTimezone(name string) error {
	api, err := openWindowsAPI()
	if err != nil {
		return err
	}
	defer api.close()
	info, err := api.information(name)
	if err != nil {
		return err
	}
	return withWindowsTimezonePrivilege(api.adjustPrivileges, func() error {
		result, _, err := api.setTimezone.Call(uintptr(unsafe.Pointer(info)))
		if result == 0 {
			return fmt.Errorf("设置 Windows 系统时区失败: %w", err)
		}
		return nil
	})
}
