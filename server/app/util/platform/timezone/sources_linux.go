package timezone

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/godbus/dbus/v5"
)

// sources 优先使用正在运行的 timesyncd 时间源，再读取标准配置文件。
func sources(ctx context.Context) ([]string, error) {
	if values, active, err := systemdSources(ctx); active || err != nil {
		return values, err
	}
	return fileSources(ctx)
}

// systemdSources 只读取正在运行的服务，不通过查询隐式启动 timesyncd。
func systemdSources(ctx context.Context) ([]string, bool, error) {
	socket, err := (&net.Dialer{}).DialContext(ctx, "unix", "/var/run/dbus/system_bus_socket")
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ECONNREFUSED) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("连接系统时间服务失败: %w", err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = socket.SetDeadline(deadline)
	}
	connection, err := dbus.NewConn(socket, dbus.WithContext(ctx))
	if err != nil {
		_ = socket.Close()
		return nil, false, err
	}
	defer connection.Close()
	if err := connection.Auth(nil); err != nil {
		return nil, false, fmt.Errorf("认证系统时间服务失败: %w", err)
	}
	if err := connection.Hello(); err != nil {
		return nil, false, fmt.Errorf("初始化系统时间服务失败: %w", err)
	}
	var active bool
	if err := connection.BusObject().CallWithContext(ctx, "org.freedesktop.DBus.NameHasOwner", 0, "org.freedesktop.timesync1").Store(&active); err != nil {
		return nil, false, fmt.Errorf("查询系统时间服务失败: %w", err)
	}
	if !active {
		return nil, false, nil
	}
	var properties map[string]dbus.Variant
	if err := connection.Object("org.freedesktop.timesync1", "/org/freedesktop/timesync1").CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", dbus.FlagNoAutoStart, "org.freedesktop.timesync1.Manager").Store(&properties); err != nil {
		return nil, true, fmt.Errorf("读取系统时间源失败: %w", err)
	}
	var values []string
	if name, ok := properties["ServerName"].Value().(string); ok && name != "" {
		values = append(values, name)
	}
	for _, key := range []string{"RuntimeNTPServers", "SystemNTPServers", "LinkNTPServers"} {
		if names, ok := properties[key].Value().([]string); ok {
			values = append(values, names...)
		}
	}
	// FallbackNTP 只在没有普通时间源时生效，不能用于绕过已配置源的故障。
	if len(values) == 0 {
		if names, ok := properties["FallbackNTPServers"].Value().([]string); ok {
			values = append(values, names...)
		}
	}
	if len(values) == 0 {
		return nil, true, errors.New("系统时间服务未提供可用的 NTP 时间源")
	}
	return values, true, nil
}
