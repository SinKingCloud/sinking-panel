package timezone

import (
	"errors"
	"fmt"
	"slices"
	"time"
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

// Get 读取系统时区和当前当地时间，不受进程 TZ 环境变量或 time.Local 缓存影响。
func Get() (string, time.Time, error) {
	library, err := loadWindowsLibrary("kernel32.dll")
	if err != nil {
		return "", time.Time{}, err
	}
	defer library.Release()
	get, err := library.FindProc("GetDynamicTimeZoneInformation")
	if err != nil {
		return "", time.Time{}, fmt.Errorf("系统不支持读取 Windows 动态时区: %w", err)
	}
	var info windowsTimezoneInformation
	result, _, err := get.Call(uintptr(unsafe.Pointer(&info)))
	if uint32(result) == 0xffffffff { // TIME_ZONE_ID_INVALID
		return "", time.Time{}, fmt.Errorf("读取 Windows 系统时区失败: %w", err)
	}
	name := windows.UTF16ToString(info.TimeZoneKeyName[:])
	if name == "" {
		return "", time.Time{}, errors.New("Windows 当前使用的自定义时区没有系统时区 ID")
	}
	// Windows 使用 UTC = 当地时间 + Bias，返回状态已包含当前夏令时判断。
	bias := info.Bias
	switch uint32(result) {
	case windows.TIME_ZONE_ID_STANDARD:
		bias += info.StandardBias
	case windows.TIME_ZONE_ID_DAYLIGHT:
		bias += info.DaylightBias
	}
	return name, time.Now().In(time.FixedZone(name, -int(bias)*60)), nil
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
