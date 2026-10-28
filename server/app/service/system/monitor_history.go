package system

import (
	"errors"
	"fmt"
	"log"
	"math"
	"server/app/constant"
	"server/app/util/recordstore"
	"time"
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

func (s *service) GetMonitorHistory(start, end int64) (recordstore.RecordResult, error) {
	recorder, err := s.monitorStorage()
	if err != nil {
		return recordstore.RecordResult{}, err
	}
	result, err := recorder.Query(start, end, recordstore.SeriesQueryOptions{MaxPoints: monitorHistoryMaxPoints})
	if err != nil {
		return result, err
	}
	return fillMonitorHistory(result)
}

// recordMonitor 只负责采集系统指标；排队、汇总、启停及落盘由 recordstore.Recorder 处理。
func (s *service) recordMonitor(now time.Time, cpu, memory, load map[string]interface{}, netErr, diskErr error) {
	recorder := s.monitorRecorder
	if recorder == nil || !recorder.Enabled() {
		return
	}
	if cpu["available"] != true || memory["available"] != true || load["available"] != true {
		recorder.Report(errors.New("CPU、内存或系统负载采集不可用，本次采样已跳过"))
		return
	}
	if err := errors.Join(netErr, diskErr); err != nil {
		recorder.Report(fmt.Errorf("系统 I/O 采集不可用: %w", err))
		return
	}
	values := map[string]float64{
		"cpu": monitorNumber(cpu["usage"]), "cpu_max": monitorNumber(cpu["usage"]),
		"memory": monitorNumber(memory["used_percent"]), "memory_used": monitorNumber(memory["used"]), "memory_total": monitorNumber(memory["total"]),
		"disk_usage": 0, "disk_used": 0, "disk_total": 0, "disk_read": 0, "disk_write": 0,
		"net_sent": 0, "net_recv": 0,
		"load1": monitorNumber(load["load1"]), "load5": monitorNumber(load["load5"]), "load15": monitorNumber(load["load15"]),
	}
	s.monitorDeviceMu.RLock()
	if !s.monitorDiskReady {
		s.monitorDeviceMu.RUnlock()
		recorder.Report(errors.New("磁盘容量采集不可用，本次采样已跳过"))
		return
	}
	s.ioCacheLock.RLock()
	for name, rates := range s.netRateCache {
		if s.monitorNetworks[name] {
			values["net_sent"] += monitorNumber(rates["send_rate"])
			values["net_recv"] += monitorNumber(rates["recv_rate"])
		}
	}
	for name, rates := range s.diskRateCache {
		if s.monitorDisks[name] {
			values["disk_read"] += monitorNumber(rates["read_bytes_rate"])
			values["disk_write"] += monitorNumber(rates["write_bytes_rate"])
		}
	}
	s.ioCacheLock.RUnlock()
	s.monitorDeviceMu.RUnlock()
	s.disksInfoCacheLock.RLock()
	values["disk_used"], values["disk_total"] = monitorDiskCapacity(s.disksInfoCache)
	s.disksInfoCacheLock.RUnlock()
	if values["disk_total"] > 0 {
		values["disk_usage"] = values["disk_used"] / values["disk_total"] * 100
	}
	// 失败和队列溢出由记录器统一保留到 Settings.LastError。
	_ = recorder.Record(now.UnixMilli(), values)
}

func monitorNumber(value interface{}) float64 {
	var result float64
	switch number := value.(type) {
	case float64:
		result = number
	case uint64:
		result = float64(number)
	case int64:
		result = float64(number)
	case int:
		result = float64(number)
	}
	if math.IsNaN(result) || math.IsInf(result, 0) || result < 0 {
		return 0
	}
	return result
}
