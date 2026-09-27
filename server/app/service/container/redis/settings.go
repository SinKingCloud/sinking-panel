package redis

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

// GetConfig 返回已保存的目标配置，需要重启的改动可能尚未生效。
func (m *Manager) GetConfig() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := m.readFile("redis.conf")
	return string(data), err
}

func (m *Manager) GetPerformance() (*Performance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	config, err := m.readConfig("redis.conf")
	if err != nil {
		return nil, err
	}
	return &config.Performance, nil
}

func (m *Manager) GetPersistence() (*Persistence, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	config, err := m.readConfig("redis.conf")
	if err != nil {
		return nil, err
	}
	return &config.Persistence, nil
}

// performanceValues 保留 nil 与零值的区别，空密码、零超时和零内存限制均可正常提交。
func (m *Manager) performanceValues(data *PerformanceUpdate) map[string][]string {
	values := make(map[string][]string)
	if data == nil {
		return values
	}
	if data.Bind != nil {
		values["bind"] = strings.Fields(*data.Bind)
	}
	if data.Port != nil {
		values["port"] = []string{strconv.Itoa(*data.Port)}
	}
	if data.Timeout != nil {
		values["timeout"] = []string{strconv.Itoa(*data.Timeout)}
	}
	if data.Maxclients != nil {
		values["maxclients"] = []string{strconv.Itoa(*data.Maxclients)}
	}
	if data.Databases != nil {
		values["databases"] = []string{strconv.Itoa(*data.Databases)}
	}
	if data.Requirepass != nil {
		values["requirepass"] = []string{*data.Requirepass}
	}
	if data.Maxmemory != nil {
		values["maxmemory"] = []string{strconv.FormatInt(*data.Maxmemory, 10) + "mb"}
	}
	return values
}

func (m *Manager) UpdatePerformance(ctx context.Context, data *PerformanceUpdate) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data == nil {
		return nil, errors.New("性能配置不能为空")
	}
	return m.updateConfigLocked(ctx, "", m.performanceValues(data))
}

func (m *Manager) UpdatePersistence(ctx context.Context, data *PersistenceUpdate) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data == nil {
		return nil, errors.New("持久化配置不能为空")
	}
	values := make(map[string][]string)
	if data.Save != nil {
		values["save"] = []string{""}
		if len(*data.Save) > 0 {
			values["save"] = nil
			for _, rule := range *data.Save {
				values["save"] = append(values["save"], strconv.Itoa(rule.Seconds), strconv.Itoa(rule.Changes))
			}
		}
	}
	if data.AppendOnly != nil {
		value := "no"
		if *data.AppendOnly {
			value = "yes"
		}
		values["appendonly"] = []string{value}
	}
	if data.AppendFsync != nil {
		values["appendfsync"] = []string{*data.AppendFsync}
	}
	if data.AutoAofRewritePercentage != nil {
		values["auto-aof-rewrite-percentage"] = []string{strconv.Itoa(*data.AutoAofRewritePercentage)}
	}
	if data.AutoAofRewriteMinSize != nil {
		values["auto-aof-rewrite-min-size"] = []string{strconv.FormatInt(*data.AutoAofRewriteMinSize, 10) + "mb"}
	}
	return m.updateConfigLocked(ctx, "", values)
}

func (m *Manager) UpdateConfig(ctx context.Context, content string) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateConfigLocked(ctx, content, nil)
}

// updateConfigLocked 统一保存和应用配置：values 非 nil 时局部更新，否则替换完整文本。
// 只读取一次旧文件；应用失败同时恢复目标配置，运行配置由 applyLocked 回滚。
func (m *Manager) updateConfigLocked(ctx context.Context, content string, values map[string][]string) (*ApplyResult, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	previous, err := m.readFile("redis.conf")
	if err != nil {
		return nil, err
	}
	var config *redisConfig
	if values != nil {
		config, err = m.patchConfig(string(previous), values)
	} else {
		config, err = m.parseConfig(content)
	}
	if err != nil {
		return nil, err
	}
	if err = m.writeFile("redis.conf", []byte(config.content)); err != nil {
		return nil, err
	}
	result, err := m.applyLocked(ctx, config, false)
	if err != nil {
		return nil, errors.Join(err, m.writeFile("redis.conf", previous))
	}
	return result, nil
}

// Reload 通过 CONFIG SET 重载可变参数；监听地址、数据库数等变动提示显式重启。
func (m *Manager) Reload(ctx context.Context) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	config, err := m.readConfig("redis.conf")
	if err != nil {
		return nil, err
	}
	return m.applyLocked(ctx, config, true)
}

// applyLocked 保存下次启动配置；运行时可热更新的参数立即应用，其余等待通用容器重启。
func (m *Manager) applyLocked(ctx context.Context, next *redisConfig, reload bool) (*ApplyResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil && err.Error() != "实例不存在" {
		return nil, err
	}
	if instance == nil || instance.Status == container.StatusStopped || instance.Status == container.StatusFailed {
		if previous, err := m.readConfig("config/redis.conf"); err == nil && !previous.AppendOnly && next.AppendOnly {
			return nil, errors.New("请先启动 Redis，再在线开启 AOF 持久化")
		}
		if err = m.writeFile("config/redis.conf", []byte(next.content)); err != nil {
			return nil, err
		}
		return &ApplyResult{}, nil
	}
	if instance.Status != container.StatusRunning {
		return nil, errors.New("Redis 正在启动或停止，请稍后修改配置")
	}
	previous, err := m.readRunningConfig()
	if err != nil {
		return nil, err
	}
	generation := m.configGeneration
	startup, err := m.readFile("config/redis.conf")
	if err != nil {
		return nil, err
	}
	image, err := m.container.Image(instance.ImageID)
	if err != nil {
		return nil, err
	}
	maxclientsHot := false
	for _, env := range image.Env {
		if version, ok := strings.CutPrefix(env, "REDIS_VERSION="); ok {
			majorText, _, _ := strings.Cut(version, ".")
			major, _ := strconv.Atoi(majorText)
			maxclientsHot = major >= 3
		}
	}
	changes, rollback := make(map[string][]string), make(map[string][]string)
	restartRequired := !slices.Equal(previous.directives, next.directives)
	// include 可能覆盖主文件中的参数，必须由 Redis 重新加载完整配置。
	if (previous.includes || next.includes) && (reload || previous.content != next.content) {
		restartRequired = true
	}
	for key, value := range next.values {
		old, exists := previous.values[key]
		changed := !slices.Equal(old, value)
		// values 只包含面板识别的参数，其中监听地址、端口和数据库数量等待显式重启。
		hot := exists && key != "bind" && key != "port" && key != "databases"
		// Redis 2.x 的最大连接数只能在启动时设置。
		if key == "maxclients" && !maxclientsHot {
			hot = false
		}
		if changed && !hot {
			restartRequired = true
		}
		// 显式重载也重新下发未变化的值，覆盖外部 CONFIG SET 的临时调整。
		if hot && (changed || reload) {
			changes[key], rollback[key] = value, old
		}
	}
	for key := range previous.values {
		if _, exists := next.values[key]; !exists {
			restartRequired = true
		}
	}
	if restartRequired {
		if !previous.AppendOnly && next.AppendOnly {
			return nil, errors.New("请先单独开启 AOF 持久化，再修改需要重启的配置")
		}
		if err = m.writeFile("config/redis.conf", []byte(next.content)); err != nil {
			return nil, err
		}
		return &ApplyResult{RestartRequired: true}, nil
	}
	if len(changes) > 0 {
		enablingAOF := false
		if _, changed := changes["appendonly"]; changed && next.AppendOnly {
			output, readErr := m.command(ctx, previous, "CONFIG", "GET", "appendonly")
			if readErr != nil {
				return nil, errors.New("无法读取 Redis 持久化状态，请检查服务连接")
			}
			enablingAOF = strings.TrimSpace(output) != "appendonly\nyes"
		}
		if err = m.setConfig(ctx, previous, changes); err != nil {
			return nil, m.rollbackConfig(previous, next, rollback)
		}
		if enablingAOF {
			// 等待现有内存数据落入 AOF 后才保存配置，防止紧接着重启丢失数据。
			waitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			ticker := time.NewTicker(200 * time.Millisecond)
			complete := false
			for !complete && waitCtx.Err() == nil {
				info, infoErr := m.command(waitCtx, next, "INFO", "persistence")
				if infoErr != nil {
					break
				}
				values := make(map[string]string)
				for line := range strings.Lines(info) {
					if key, value, ok := strings.Cut(strings.TrimSpace(line), ":"); ok {
						values[key] = value
					}
				}
				if values["aof_last_bgrewrite_status"] == "err" {
					break
				}
				complete = values["aof_enabled"] == "1" && values["aof_rewrite_in_progress"] == "0" && values["aof_rewrite_scheduled"] == "0" && values["aof_last_bgrewrite_status"] == "ok"
				if !complete {
					select {
					case <-waitCtx.Done():
					case <-ticker.C:
					}
				}
			}
			ticker.Stop()
			cancel()
			if !complete {
				return nil, m.rollbackConfig(previous, next, rollback)
			}
		}
	}
	if err = m.writeFile("config/redis.conf", []byte(next.content)); err != nil {
		if len(changes) > 0 {
			return nil, errors.Join(err, m.rollbackConfig(previous, next, rollback))
		}
		return nil, err
	}
	if err = m.writeRunningConfig([]byte(next.content), generation); err != nil {
		restoreErr := m.writeFile("config/redis.conf", startup)
		if len(changes) > 0 {
			restoreErr = errors.Join(restoreErr, m.rollbackConfig(previous, next, rollback))
		}
		return nil, errors.Join(err, restoreErr)
	}
	return &ApplyResult{Applied: true}, nil
}

// setConfig 在 Redis 7 及以上批量更新，旧版逐项更新并最后修改密码，失败统一由调用方回滚。
func (m *Manager) setConfig(ctx context.Context, current *redisConfig, values map[string][]string) error {
	if len(values) == 0 {
		return nil
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil {
		return err
	}
	image, err := m.container.Image(instance.ImageID)
	if err != nil {
		return err
	}
	version := ""
	for _, value := range image.Env {
		if value, ok := strings.CutPrefix(value, "REDIS_VERSION="); ok {
			version = value
		}
	}
	majorText, _, _ := strings.Cut(version, ".")
	major, err := strconv.Atoi(majorText)
	if err != nil || major <= 0 {
		return errors.New("镜像缺少有效的 Redis 版本信息，请重新解析镜像")
	}
	// 旧版 CONFIG SET 只接受字节数；配置文件继续保留易读的内存单位。
	values = maps.Clone(values)
	for _, key := range []string{"maxmemory", "auto-aof-rewrite-min-size"} {
		if value, exists := values[key]; exists {
			megabytes, err := m.configMemoryMB(strings.Join(value, " "))
			if err != nil {
				return err
			}
			values[key] = []string{strconv.FormatInt(megabytes*1024*1024, 10)}
		}
	}
	keys := slices.Sorted(maps.Keys(values))
	if major < 7 {
		for _, key := range keys {
			if key == "requirepass" {
				continue
			}
			if _, err = m.command(ctx, current, "CONFIG", "SET", key, strings.Join(values[key], " ")); err != nil {
				return err
			}
		}
		if value, exists := values["requirepass"]; exists {
			_, err = m.command(ctx, current, "CONFIG", "SET", "requirepass", strings.Join(value, " "))
		}
		return err
	}
	args := []string{"CONFIG", "SET"}
	for _, key := range keys {
		args = append(args, key, strings.Join(values[key], " "))
	}
	_, err = m.command(ctx, current, args...)
	return err
}

// rollbackConfig 独立于请求取消执行，并兼容应用前后密码不同的情况。
func (m *Manager) rollbackConfig(previous, next *redisConfig, values map[string][]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if m.setConfig(ctx, next, values) == nil || m.setConfig(ctx, previous, values) == nil {
		return errors.New("Redis 配置应用失败，已恢复原配置")
	}
	return errors.New("Redis 配置应用失败且恢复未完成，请重启实例以恢复保存的配置")
}
