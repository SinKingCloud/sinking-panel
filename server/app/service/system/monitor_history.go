package system

import (
	"errors"
	"fmt"
	"log"
	"server/app/constant"
	"server/app/util/psutil"
	"server/app/util/recordstore"
)

// 字段顺序与已保存的历史文件保持一致；聚合和存储由通用记录器负责。
var systemMonitorFields = []recordstore.Field{
	{Name: "cpu", Aggregate: recordstore.Mean},
	{Name: "cpu_max", Aggregate: recordstore.Max},
	{Name: "memory", Aggregate: recordstore.Mean},
	{Name: "memory_used", Aggregate: recordstore.Mean},
	{Name: "memory_total", Aggregate: recordstore.Mean},
	{Name: "disk_usage", Aggregate: recordstore.Mean},
	{Name: "disk_used", Aggregate: recordstore.Mean},
	{Name: "disk_total", Aggregate: recordstore.Mean},
	{Name: "disk_read", Aggregate: recordstore.Mean},
	{Name: "disk_write", Aggregate: recordstore.Mean},
	{Name: "net_sent", Aggregate: recordstore.Mean},
	{Name: "net_recv", Aggregate: recordstore.Mean},
	{Name: "load1", Aggregate: recordstore.Mean},
	{Name: "load5", Aggregate: recordstore.Mean},
	{Name: "load15", Aggregate: recordstore.Mean},
}

func (s *service) startMonitorHistory() {
	configuration, err := s.loadMonitorConfiguration()
	if err != nil {
		s.monitorInitError = err
		log.Printf("读取系统监控配置失败: %v", err)
		return
	}
	s.monitorRecorder, s.monitorInitError = recordstore.OpenRecorder(constant.MonitorPath, systemMonitorFields, recordstore.SeriesOptions{
		Configuration:     &configuration,
		SaveConfiguration: s.saveMonitorConfiguration,
	})
	if s.monitorInitError != nil {
		log.Printf("打开系统监控历史失败: %v", s.monitorInitError)
		return
	}
}

func (s *service) monitorStorage() (*recordstore.Recorder, error) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.closed {
		return nil, errors.New("系统监控服务已停止")
	}
	if s.monitorInitError != nil {
		return nil, fmt.Errorf("监控存储初始化失败: %w", s.monitorInitError)
	}
	if s.monitorRecorder == nil {
		return nil, errors.New("系统监控服务尚未启动")
	}
	return s.monitorRecorder, nil
}

func (s *service) GetMonitorSettings() (recordstore.SeriesSettings, error) {
	recorder, err := s.monitorStorage()
	if err != nil {
		return recordstore.SeriesSettings{}, err
	}
	return recorder.Settings()
}

func (s *service) SetMonitorSettings(enabled bool, days int) error {
	recorder, err := s.monitorStorage()
	if err != nil {
		return err
	}
	return recorder.UpdateSettings(enabled, days)
}

// GetMonitorHistory 按展示宽度控制查询点数，省略时使用默认预算；原始分钟数据保持不变。
func (s *service) GetMonitorHistory(start, end int64, maxPoints ...int) (recordstore.RecordResult, error) {
	if len(maxPoints) > 1 {
		return recordstore.RecordResult{}, errors.New("监控历史只接受一个查询点数")
	}
	maximum := monitorHistoryDefaultPoints
	if len(maxPoints) == 1 {
		maximum = maxPoints[0]
	}
	if maximum < monitorHistoryMinPoints || maximum > monitorHistoryMaxPoints {
		return recordstore.RecordResult{}, fmt.Errorf("监控历史查询点数必须为 %d 至 %d", monitorHistoryMinPoints, monitorHistoryMaxPoints)
	}
	recorder, err := s.monitorStorage()
	if err != nil {
		return recordstore.RecordResult{}, err
	}
	result, err := recorder.Query(start, end, recordstore.SeriesQueryOptions{MaxPoints: maximum})
	if err != nil {
		return result, err
	}
	return s.fillMonitorHistory(result)
}

// recordMonitor 将本轮资源快照汇总成业务监控字段；排队、聚合及落盘由记录器处理。
func (s *service) recordMonitor(snapshot psutil.Snapshot, netErr, diskErr error) {
	recorder := s.monitorRecorder
	if recorder == nil || !recorder.Enabled() {
		return
	}
	if !snapshot.CPU.Available || !snapshot.Memory.Available || !snapshot.Load.Available {
		recorder.Report(errors.New("CPU、内存或系统负载采集不可用，本次采样已跳过"))
		return
	}
	if err := errors.Join(netErr, diskErr); err != nil {
		recorder.Report(fmt.Errorf("系统 I/O 采集不可用: %w", err))
		return
	}
	values := map[string]float64{
		"cpu":          snapshot.CPU.Usage,
		"cpu_max":      snapshot.CPU.Usage,
		"memory":       snapshot.Memory.UsedPercent,
		"memory_used":  float64(snapshot.Memory.Used),
		"memory_total": float64(snapshot.Memory.Total),
		"disk_usage":   0,
		"disk_used":    0,
		"disk_total":   0,
		"disk_read":    0,
		"disk_write":   0,
		"net_sent":     0,
		"net_recv":     0,
		"load1":        snapshot.Load.Load1,
		"load5":        snapshot.Load.Load5,
		"load15":       snapshot.Load.Load15,
	}
	s.monitorMu.RLock()
	defer s.monitorMu.RUnlock()
	if !s.monitorDiskReady {
		recorder.Report(errors.New("磁盘容量采集不可用，本次采样已跳过"))
		return
	}
	for _, value := range snapshot.Network {
		if s.monitorNetworks[value.Name] {
			if !value.RateAvailable {
				recorder.Report(errors.New("网卡采样基线已重置，本次采样已跳过"))
				return
			}
			values["net_sent"] += value.SendRate
			values["net_recv"] += value.RecvRate
		}
	}
	for _, value := range snapshot.Disk {
		if s.monitorDisks[value.Name] {
			if !value.RateAvailable {
				recorder.Report(errors.New("磁盘采样基线已重置，本次采样已跳过"))
				return
			}
			values["disk_read"] += value.ReadBytesRate
			values["disk_write"] += value.WriteBytesRate
		}
	}
	values["disk_used"], values["disk_total"] = s.monitorDiskCapacity(s.disksInfoCache)
	if values["disk_total"] > 0 {
		values["disk_usage"] = values["disk_used"] / values["disk_total"] * 100
	}
	_ = recorder.Record(snapshot.Time, values)
}
