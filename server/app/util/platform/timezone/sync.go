package timezone

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var syncing sync.Mutex

// Sync 执行一次时间同步；优先使用系统时间源，未配置时使用公共 NTP。
// 整体最多等待 15 秒，不修改时区、时间服务配置或启动后台任务。
func Sync(ctx context.Context) error {
	if !syncing.TryLock() {
		return errors.New("时间正在同步，请稍后再试")
	}
	defer syncing.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	discovery, stop := context.WithTimeout(ctx, 3*time.Second)
	servers, err := sources(discovery)
	stop()
	if err != nil {
		return fmt.Errorf("读取系统时间源失败: %w", err)
	}
	if len(servers) == 0 {
		servers = []string{"time.cloudflare.com", "ntp.aliyun.com"}
	}
	seen := make(map[string]bool, len(servers))
	var failures []error
	for _, server := range servers {
		server = strings.TrimSpace(server)
		key := strings.ToLower(server)
		if seen[key] {
			continue
		}
		seen[key] = true
		if err = ctx.Err(); err != nil {
			return fmt.Errorf("时间同步已取消或超时: %w", errors.Join(append(failures, err)...))
		}
		offset, queryErr := query(ctx, server)
		if queryErr != nil {
			failures = append(failures, fmt.Errorf("%s: %w", server, queryErr))
			continue
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = setTime(time.Now().Add(offset)); err != nil {
			return fmt.Errorf("设置系统时间失败: %w", err)
		}
		return nil
	}
	return fmt.Errorf("同步时间失败: %w", errors.Join(failures...))
}
