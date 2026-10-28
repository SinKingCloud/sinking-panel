package system

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"server/app/constant"
	"server/app/util/recordstore"
)

// loadMonitorConfiguration 从配置表读取设置，为缺失项填入默认值。
// 查询失败和非法配置均直接返回错误，不能当成未配置并覆盖用户已有数据。
func (s *service) loadMonitorConfiguration() (recordstore.SeriesConfiguration, error) {
	configuration := recordstore.SeriesConfiguration{Enabled: true, RetentionDays: 30}
	if s.configService == nil {
		return configuration, errors.New("系统配置服务未初始化")
	}
	configs, err := s.configService.GroupWithError(constant.MonitorGroup)
	if err != nil {
		return configuration, fmt.Errorf("读取监控配置表失败: %w", err)
	}
	enabled, hasEnabled := configs[constant.MonitorEnabled]
	days, hasDays := configs[constant.MonitorRetentionDays]
	if hasEnabled {
		configuration.Enabled, err = strconv.ParseBool(strings.TrimSpace(enabled))
		if err != nil {
			return configuration, fmt.Errorf("配置 %s 必须是有效布尔值", constant.MonitorEnabled)
		}
	}
	if hasDays {
		configuration.RetentionDays, err = strconv.Atoi(strings.TrimSpace(days))
		if err != nil || configuration.RetentionDays < 1 || configuration.RetentionDays > 365 {
			return configuration, fmt.Errorf("配置 %s 必须是 1 至 365 的整数", constant.MonitorRetentionDays)
		}
	}
	if hasEnabled && hasDays {
		return configuration, nil
	}

	missing := make(map[string]string, 2)
	if !hasEnabled {
		missing[constant.MonitorEnabled] = strconv.FormatBool(configuration.Enabled)
	}
	if !hasDays {
		missing[constant.MonitorRetentionDays] = strconv.Itoa(configuration.RetentionDays)
	}
	// 只提交缺失项，保留已保存的配置。
	if err := s.configService.Sets(missing); err != nil {
		return configuration, fmt.Errorf("保存初始监控配置失败: %w", err)
	}
	return configuration, nil
}

// saveMonitorConfiguration 在记录器的设置锁内调用，只保存配置表，不反向调用记录器。
func (s *service) saveMonitorConfiguration(configuration recordstore.SeriesConfiguration) error {
	if s.configService == nil {
		return errors.New("系统配置服务未初始化")
	}
	if configuration.RetentionDays < 1 || configuration.RetentionDays > 365 {
		return errors.New("监控保留天数必须为 1 至 365 天")
	}
	if err := s.configService.Sets(map[string]string{
		constant.MonitorEnabled:       strconv.FormatBool(configuration.Enabled),
		constant.MonitorRetentionDays: strconv.Itoa(configuration.RetentionDays),
	}); err != nil {
		return fmt.Errorf("保存监控配置失败: %w", err)
	}
	return nil
}
