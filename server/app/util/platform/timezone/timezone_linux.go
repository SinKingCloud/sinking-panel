package timezone

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/godbus/dbus/v5"
)

func setTimezone(name string) error {
	database, err := openDatabase()
	if err != nil {
		return err
	}
	source, _, err := database.read(name)
	if err != nil {
		return err
	}
	if handled, err := setSystemTimezone(name); handled || err != nil {
		return err
	}
	return writeTimezone(name, source)
}

// setSystemTimezone 优先使用 timedated，由系统处理权限、RTC 和变更通知。
// 只有服务不存在时返回 false, nil；拒绝或超时不得再回退修改文件。
func setSystemTimezone(name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	socket, err := (&net.Dialer{}).DialContext(ctx, "unix", "/var/run/dbus/system_bus_socket")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ECONNREFUSED) {
			return false, nil
		}
		return false, fmt.Errorf("连接系统时区服务失败: %w", err)
	}
	connection, err := dbus.NewConn(socket, dbus.WithContext(ctx))
	if err != nil {
		_ = socket.Close()
		return false, err
	}
	defer connection.Close()
	if err := connection.Auth(nil); err != nil {
		return false, fmt.Errorf("认证系统时区服务失败: %w", err)
	}
	if err := connection.Hello(); err != nil {
		return false, fmt.Errorf("初始化系统时区服务失败: %w", err)
	}
	err = connection.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").CallWithContext(ctx, "org.freedesktop.timedate1.SetTimezone", 0, name, false).Err
	if err == nil {
		return true, nil
	}
	var busError dbus.Error
	if errors.As(err, &busError) {
		switch busError.Name {
		case "org.freedesktop.DBus.Error.ServiceUnknown", "org.freedesktop.DBus.Error.NameHasNoOwner", "org.freedesktop.DBus.Error.UnknownMethod":
			return false, nil
		}
	}
	return false, fmt.Errorf("通过系统服务修改时区失败: %w", err)
}
