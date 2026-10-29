package psutil

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"slices"
	"strings"
	"syscall"

	gnet "github.com/shirou/gopsutil/v4/net"
)

// Connections 查询当前权限可见的 TCP/UDP 连接，按端点及进程稳定排序。
// 该方法仅查询，不扫描网络，不修改监听端口；权限不足时可能无法取得连接所属 PID。
func Connections(ctx context.Context, filter ConnectionFilter) ([]ConnectionInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if filter.PID < 0 || filter.LocalPort > 65535 || filter.RemotePort > 65535 {
		return nil, errors.New("进程 ID 或端口不合法")
	}
	kind := strings.ToLower(strings.TrimSpace(filter.Kind))
	switch kind {
	case "", "all":
		// gopsutil 的 all 在 Unix 系统包含 Unix 套接字，统一使用 inet。
		kind = "inet"
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
	default:
		return nil, errors.New("连接类型必须为 all、tcp、tcp4、tcp6、udp、udp4 或 udp6")
	}
	var records []gnet.ConnectionStat
	var err error
	if runtime.GOOS == "linux" {
		// 返回格式不包含 UID，避免为每个连接额外读取进程用户信息。
		records, err = gnet.ConnectionsPidWithoutUidsWithContext(ctx, kind, filter.PID)
	} else if filter.PID > 0 {
		records, err = gnet.ConnectionsPidWithContext(ctx, kind, filter.PID)
	} else {
		records, err = gnet.ConnectionsWithContext(ctx, kind)
	}
	if err != nil {
		return nil, fmt.Errorf("读取网络连接失败: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Windows 与 Unix 的状态名统一为同一格式，筛选也接受各平台的原始名称。
	normalizeStatus := func(value string) string {
		value = strings.ToUpper(strings.TrimSpace(value))
		switch value {
		case "SYN_RECEIVED":
			return "SYN_RECV"
		case "FIN_WAIT_1":
			return "FIN_WAIT1"
		case "FIN_WAIT_2":
			return "FIN_WAIT2"
		case "CLOSED":
			return "CLOSE"
		default:
			return value
		}
	}
	status := normalizeStatus(filter.Status)
	result := make([]ConnectionInfo, 0, len(records))
	for _, record := range records {
		if filter.PID > 0 && record.Pid != filter.PID ||
			filter.LocalPort > 0 && record.Laddr.Port != filter.LocalPort ||
			filter.RemotePort > 0 && record.Raddr.Port != filter.RemotePort {
			continue
		}
		item := ConnectionInfo{
			PID: record.Pid,
			FD:  record.Fd,
			Local: AddressInfo{
				IP:   record.Laddr.IP,
				Port: record.Laddr.Port,
			},
			Remote: AddressInfo{
				IP:   record.Raddr.IP,
				Port: record.Raddr.Port,
			},
			Status: normalizeStatus(record.Status),
		}
		switch record.Family {
		case syscall.AF_INET:
			item.Family = "ipv4"
		case syscall.AF_INET6:
			item.Family = "ipv6"
		default:
			continue
		}
		switch record.Type {
		case syscall.SOCK_STREAM:
			item.Protocol = "tcp"
		case syscall.SOCK_DGRAM:
			item.Protocol = "udp"
			item.Status = "BOUND"
			if item.Remote.Port > 0 {
				item.Status = "CONNECTED"
			}
		default:
			continue
		}
		if status == "" || status == item.Status {
			result = append(result, item)
		}
	}
	slices.SortFunc(result, func(a, b ConnectionInfo) int {
		return cmp.Or(
			cmp.Compare(a.Protocol, b.Protocol),
			cmp.Compare(a.Family, b.Family),
			cmp.Compare(a.Local.IP, b.Local.IP),
			cmp.Compare(a.Local.Port, b.Local.Port),
			cmp.Compare(a.Remote.IP, b.Remote.IP),
			cmp.Compare(a.Remote.Port, b.Remote.Port),
			cmp.Compare(a.PID, b.PID),
			cmp.Compare(a.FD, b.FD),
			cmp.Compare(a.Status, b.Status),
		)
	})
	return result, nil
}

// Ports 查询正在监听的 TCP 和已绑定端口的 UDP，保留每个套接字及所属进程。
// UDP 没有 LISTEN 状态，已连接 UDP 也会占用本地端口，因此一并返回。
func Ports(ctx context.Context, filter ConnectionFilter) ([]ConnectionInfo, error) {
	connections, err := Connections(ctx, filter)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(connections, func(item ConnectionInfo) bool {
		return item.Local.Port == 0 || item.Protocol == "tcp" && item.Status != "LISTEN"
	}), nil
}

// Interfaces 返回全部网卡（包括环回），按名称排序；个别地址读取失败时返回部分数据与错误。
func Interfaces(ctx context.Context) ([]InterfaceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("读取网卡信息失败: %w", err)
	}
	result := make([]InterfaceInfo, 0, len(interfaces))
	var failures []error
	for _, network := range interfaces {
		if err := ctx.Err(); err != nil {
			return result, errors.Join(append(failures, err)...)
		}
		item := InterfaceInfo{
			Index:      network.Index,
			Name:       network.Name,
			MTU:        network.MTU,
			MAC:        network.HardwareAddr.String(),
			Flags:      make([]string, 0, 6),
			Addresses:  make([]string, 0),
			IsUp:       network.Flags&net.FlagUp != 0,
			IsLoopback: network.Flags&net.FlagLoopback != 0,
		}
		for _, flag := range []net.Flags{net.FlagUp, net.FlagBroadcast, net.FlagLoopback, net.FlagPointToPoint, net.FlagMulticast, net.FlagRunning} {
			if network.Flags&flag != 0 {
				item.Flags = append(item.Flags, flag.String())
			}
		}
		addresses, err := network.Addrs()
		if err != nil {
			item.Error = err.Error()
			failures = append(failures, fmt.Errorf("读取网卡 %s 地址失败: %w", network.Name, err))
		} else {
			for _, address := range addresses {
				item.Addresses = append(item.Addresses, address.String())
			}
			slices.Sort(item.Addresses)
		}
		result = append(result, item)
	}
	slices.SortFunc(result, func(a, b InterfaceInfo) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Index, b.Index))
	})
	return result, errors.Join(failures...)
}
