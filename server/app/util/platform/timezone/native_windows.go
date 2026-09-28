package timezone

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// windowsTimezoneInformation 对应 DYNAMIC_TIME_ZONE_INFORMATION，前部保留原生结构布局。
type windowsTimezoneInformation struct {
	windows.Timezoneinformation
	TimeZoneKeyName             [128]uint16
	DynamicDaylightTimeDisabled uint8
}

// windowsAPI 持有本次设置所需的 DLL 和原生方法，使用后统一释放。
type windowsAPI struct {
	kernel, security *windows.DLL
	enumerate        *windows.Proc
	systemTime       *windows.Proc
	localTime        *windows.Proc
	annualRule       *windows.Proc
	setTimezone      *windows.Proc
	adjustPrivileges *windows.Proc
}

// loadWindowsLibrary 只从系统目录加载 DLL，避免搜索到工作目录中的同名文件。
func loadWindowsLibrary(name string) (*windows.DLL, error) {
	handle, err := windows.LoadLibraryEx(name, 0, windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return nil, fmt.Errorf("加载 Windows 系统库 %s 失败: %w", name, err)
	}
	return &windows.DLL{Name: name, Handle: handle}, nil
}

func openWindowsAPI() (*windowsAPI, error) {
	kernel, err := loadWindowsLibrary("kernel32.dll")
	if err != nil {
		return nil, err
	}
	security, err := loadWindowsLibrary("advapi32.dll")
	if err != nil {
		kernel.Release()
		return nil, err
	}
	api := &windowsAPI{kernel: kernel, security: security}
	for _, entry := range []struct {
		library *windows.DLL
		name    string
		target  **windows.Proc
	}{
		{kernel, "GetSystemTime", &api.systemTime},
		{kernel, "SystemTimeToTzSpecificLocalTimeEx", &api.localTime},
		{kernel, "GetTimeZoneInformationForYear", &api.annualRule},
		{kernel, "SetDynamicTimeZoneInformation", &api.setTimezone},
		{security, "EnumDynamicTimeZoneInformation", &api.enumerate},
		{security, "AdjustTokenPrivileges", &api.adjustPrivileges},
	} {
		*entry.target, err = entry.library.FindProc(entry.name)
		if err != nil {
			api.close()
			return nil, fmt.Errorf("Windows 时区管理所需的 %s 不可用: %w", entry.name, err)
		}
	}
	return api, nil
}

func (api *windowsAPI) close() {
	api.security.Release()
	api.kernel.Release()
}

// information 读取目标时区，并按目标当地年份载入对应的夏令时规则。
func (api *windowsAPI) information(name string) (*windowsTimezoneInformation, error) {
	var info windowsTimezoneInformation
	for index := uint32(0); ; index++ {
		result, _, _ := api.enumerate.Call(uintptr(index), uintptr(unsafe.Pointer(&info)))
		if uint32(result) == uint32(windows.ERROR_NO_MORE_ITEMS) {
			return nil, fmt.Errorf("Windows 系统不存在时区 %q", name)
		}
		if uint32(result) != 0 {
			return nil, fmt.Errorf("读取 Windows 时区列表失败: %w", windows.Errno(uint32(result)))
		}
		if windows.UTF16ToString(info.TimeZoneKeyName[:]) == name {
			break
		}
	}
	info.DynamicDaylightTimeDisabled = 0
	var utc, local windows.Systemtime
	api.systemTime.Call(uintptr(unsafe.Pointer(&utc)))
	result, _, err := api.localTime.Call(uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&utc)), uintptr(unsafe.Pointer(&local)))
	if result == 0 {
		return nil, fmt.Errorf("计算 Windows 目标时区年份失败: %w", err)
	}
	var annual windows.Timezoneinformation
	result, _, err = api.annualRule.Call(uintptr(local.Year), uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&annual)))
	if result == 0 {
		return nil, fmt.Errorf("读取 Windows 当前年度时区规则失败: %w", err)
	}
	info.Timezoneinformation = annual
	return &info, nil
}
