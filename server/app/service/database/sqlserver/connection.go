package sqlserver

import (
	"context"
	"database/sql"
	"strconv"
	"time"
)

// ConnectionInfo 返回当前连接的真实端点和认证模式，兼容 SqlServer 2005 及以后版本。
func (m *Manager) ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var address, transport sql.NullString
	var port sql.NullInt64
	var windowsOnly int
	err = m.db.QueryRowContext(ctx, "SELECT local_net_address, local_tcp_port, net_transport, CAST(SERVERPROPERTY('IsIntegratedSecurityOnly') AS int) FROM sys.dm_exec_connections WHERE session_id = @@SPID").Scan(&address, &port, &transport, &windowsOnly)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]string, 0, 4)
	if address.Valid {
		items = append(items, map[string]string{
			"label": "当前连接地址",
			"value": address.String,
		})
	}
	if port.Valid {
		items = append(items, map[string]string{
			"label": "当前连接端口",
			"value": strconv.FormatInt(port.Int64, 10),
		})
	}
	if transport.Valid {
		items = append(items, map[string]string{
			"label": "连接协议",
			"value": transport.String,
		})
	}
	authentication := "SQL Server 和 Windows 认证"
	if windowsOnly == 1 {
		authentication = "仅 Windows 认证"
	}
	items = append(items, map[string]string{
		"label": "认证模式",
		"value": authentication,
	})
	return map[string]interface{}{
		"items": items,
		"note":  "远程连接需在数据库服务器启用 TCP/IP 并放行端口；数据库用户与服务器登录关联，修改密码会影响该登录的其他连接。",
	}, nil
}
