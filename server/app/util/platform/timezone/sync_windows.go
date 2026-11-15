package timezone

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// 域发现没有取消接口；最多保留一次原生查询，避免连续超时堆积系统线程。
var domainQuery = make(chan struct{}, 1)

// sources 优先采用组策略，未配置的项再读取本机 W32Time 设置。
func sources(ctx context.Context) ([]string, error) {
	read := func(path, name string, integer bool) (string, bool, error) {
		for _, base := range []string{`SOFTWARE\Policies\Microsoft\W32Time`, `SYSTEM\CurrentControlSet\Services\W32Time`} {
			if err := ctx.Err(); err != nil {
				return "", false, err
			}
			key, err := registry.OpenKey(registry.LOCAL_MACHINE, base+`\`+path, registry.QUERY_VALUE|registry.WOW64_64KEY)
			if errors.Is(err, registry.ErrNotExist) {
				continue
			}
			if err != nil {
				return "", false, fmt.Errorf("读取 Windows 时间配置 %s 失败: %w", name, err)
			}
			var value string
			if integer {
				var number uint64
				number, _, err = key.GetIntegerValue(name)
				value = strconv.FormatUint(number, 10)
			} else {
				value, _, err = key.GetStringValue(name)
			}
			_ = key.Close()
			if errors.Is(err, registry.ErrNotExist) {
				continue
			}
			if err != nil {
				return "", false, fmt.Errorf("读取 Windows 时间配置 %s 失败: %w", name, err)
			}
			return strings.TrimSpace(value), true, nil
		}
		return "", false, nil
	}
	if enabled, found, err := read(`TimeProviders\NtpClient`, "Enabled", true); err != nil {
		return nil, err
	} else if found && enabled != "1" {
		if enabled == "0" {
			return nil, errors.New("Windows 系统已禁用 NTP 客户端")
		}
		return nil, errors.New("Windows NTP 客户端启用配置不合法")
	}
	kind, configured, err := read("Parameters", "Type", false)
	if err != nil {
		return nil, err
	}
	kind = strings.ToUpper(kind)
	if configured && kind == "" {
		return nil, errors.New("Windows 时间同步类型配置为空")
	}
	if kind == "NOSYNC" {
		return nil, errors.New("Windows 系统已关闭网络时间同步")
	}
	if kind != "" && kind != "NTP" && kind != "NT5DS" && kind != "ALLSYNC" {
		return nil, fmt.Errorf("不支持的 Windows 时间同步类型: %s", kind)
	}
	if kind == "NT5DS" {
		server, joined, err := windowsDomainSource(ctx)
		if err != nil {
			return nil, err
		}
		if !joined {
			return nil, errors.New("Windows 配置为域时间同步，但当前计算机未加入域")
		}
		// NT5DS 明确只接受域时间源，不能使用注册表中遗留的公共 NTP 地址。
		return []string{server}, nil
	}
	peers, found, err := read("Parameters", "NtpServer", false)
	if err != nil {
		return nil, err
	}
	if !found && !configured {
		return nil, nil
	}
	var result, fallback []string
	for _, peer := range strings.Fields(peers) {
		server, flags, hasFlags := strings.Cut(peer, ",")
		var options uint64
		if hasFlags {
			options, err = strconv.ParseUint(flags, 0, 32)
			if err != nil || options&^0xf != 0 {
				return nil, fmt.Errorf("Windows 时间源标记不合法: %s", peer)
			}
		}
		if server == "" {
			return nil, fmt.Errorf("Windows 时间源地址为空: %s", peer)
		}
		// SpecialInterval 和通信模式由系统服务使用；单次查询统一采用 NTP 客户端模式。
		if options&0x2 != 0 {
			fallback = append(fallback, server)
		} else {
			result = append(result, server)
		}
	}
	result = append(result, fallback...)
	if kind == "ALLSYNC" {
		// 先保存明确配置的手动源，域发现耗尽期限后仍能使用它们。
		server, joined, err := windowsDomainSource(ctx)
		if err != nil {
			if len(result) == 0 || errors.Is(err, context.Canceled) {
				return nil, err
			}
			return result, nil
		}
		if joined {
			result = append([]string{server}, result...)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("Windows 已配置时间同步，但没有可用的时间源")
	}
	return result, nil
}

// windowsDomainSource 让系统定位提供时间服务的域控制器，并按调用方期限等待。
func windowsDomainSource(ctx context.Context) (string, bool, error) {
	select {
	case domainQuery <- struct{}{}:
	case <-ctx.Done():
		return "", false, ctx.Err()
	}
	type result struct {
		server string
		joined bool
		err    error
	}
	done := make(chan result, 1)
	go func() {
		defer func() { <-domainQuery }()
		data := result{}
		defer func() { done <- data }()
		if data.err = ctx.Err(); data.err != nil {
			return
		}
		var name *uint16
		var status uint32
		data.err = windows.NetGetJoinInformation(nil, &name, &status)
		if name != nil {
			_ = windows.NetApiBufferFree((*byte)(unsafe.Pointer(name)))
		}
		if data.err != nil {
			data.err = fmt.Errorf("读取 Windows 域成员状态失败: %w", data.err)
			return
		}
		data.joined = status == windows.NetSetupDomainName
		if !data.joined {
			if status == windows.NetSetupUnknownStatus {
				data.err = errors.New("无法确定 Windows 域成员状态")
			}
			return
		}
		if data.err = ctx.Err(); data.err != nil {
			return
		}
		handle, err := windows.LoadLibraryEx("netapi32.dll", 0, windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
		if err != nil {
			data.err = fmt.Errorf("加载 Windows 域时间服务失败: %w", err)
			return
		}
		library := &windows.DLL{
			Name:   "netapi32.dll",
			Handle: handle,
		}
		defer library.Release()
		procedure, err := library.FindProc("DsGetDcNameW")
		if err != nil {
			data.err = fmt.Errorf("Windows 域时间发现接口不可用: %w", err)
			return
		}
		// 仅使用 DOMAIN_CONTROLLER_INFOW 的首字段；完整缓冲由系统分配和释放。
		var info *struct{ Name *uint16 }
		const flags = 0x00000100 | 0x00000800 | 0x00004000 | 0x40000000 // BACKGROUND_ONLY | TIMESERV_REQUIRED | AVOID_SELF | RETURN_DNS_NAME
		code, _, _ := procedure.Call(0, 0, 0, 0, flags, uintptr(unsafe.Pointer(&info)))
		if info != nil {
			defer windows.NetApiBufferFree((*byte)(unsafe.Pointer(info)))
		}
		if code != 0 {
			data.err = fmt.Errorf("定位 Windows 域时间源失败: %w", windows.Errno(code))
			return
		}
		if info != nil {
			data.server = strings.TrimPrefix(windows.UTF16PtrToString(info.Name), `\\`)
		}
		if data.server == "" {
			data.err = errors.New("Windows 未返回域时间源地址")
		}
	}()
	select {
	case data := <-done:
		if err := ctx.Err(); err != nil {
			return "", false, err
		}
		return data.server, data.joined, data.err
	case <-ctx.Done():
		return "", false, fmt.Errorf("等待 Windows 域时间源失败: %w", ctx.Err())
	}
}

// setTime 调用系统原生接口写入 UTC，当前进程须具备 SeSystemtimePrivilege。
func setTime(value time.Time) error {
	value = value.UTC()
	if value.Year() < 1601 || value.Year() > 30827 {
		return errors.New("时间超出 Windows 系统支持范围")
	}
	handle, err := windows.LoadLibraryEx("kernel32.dll", 0, windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return fmt.Errorf("加载 Windows 时间接口失败: %w", err)
	}
	library := &windows.DLL{
		Name:   "kernel32.dll",
		Handle: handle,
	}
	defer library.Release()
	procedure, err := library.FindProc("SetSystemTime")
	if err != nil {
		return fmt.Errorf("Windows 设置时间接口不可用: %w", err)
	}
	systemTime := windows.Systemtime{
		Year:         uint16(value.Year()),
		Month:        uint16(value.Month()),
		Day:          uint16(value.Day()),
		Hour:         uint16(value.Hour()),
		Minute:       uint16(value.Minute()),
		Second:       uint16(value.Second()),
		Milliseconds: uint16(value.Nanosecond() / int(time.Millisecond)),
	}
	if result, _, err := procedure.Call(uintptr(unsafe.Pointer(&systemTime))); result == 0 {
		return fmt.Errorf("设置 Windows 系统时间失败: %w", err)
	}
	return nil
}
