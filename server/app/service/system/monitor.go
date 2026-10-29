package system

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"server/app/constant"
	"server/app/util/psutil"
	"server/app/util/recordstore"
)

var panelStartedAt = time.Now().UnixMilli()

// GetInfo 组合主机信息和面板自身的运行时信息。
func (s *service) GetInfo() map[string]interface{} {
	s.monitorMu.RLock()
	info := s.systemBaseCache
	s.monitorMu.RUnlock()
	return map[string]interface{}{
		"hostname":         info.Hostname,
		"os":               info.OS,
		"platform":         info.Platform,
		"platform_version": info.PlatformVersion,
		"kernel_version":   info.KernelVersion,
		"kernel_arch":      info.KernelArch,
		"uptime":           info.Uptime,
		"boot_time":        info.BootTime,
		"version":          constant.Version,
		"go_version":       runtime.Version(),
		"num_goroutine":    runtime.NumGoroutine(),
		"panel_pid":        os.Getpid(),
		"panel_start_time": panelStartedAt,
	}
}

// GetStatis 获取同一轮缓存中的 CPU、内存、负载和网卡，以及低频刷新的磁盘容量。
func (s *service) GetStatis() map[string]interface{} {
	s.monitorMu.RLock()
	defer s.monitorMu.RUnlock()
	return map[string]interface{}{
		"cpu":       s.resourceCache.CPU,
		"memory":    s.resourceCache.Memory,
		"load":      s.resourceCache.Load,
		"network":   s.resourceCache.Network,
		"disks":     s.disksInfoCache,
		"processes": s.processCache.Clone(),
	}
}

// GetIO 获取网卡和磁盘的 I/O 统计及采样历史。
func (s *service) GetIO(netInterface, diskName string, after int64) map[string]interface{} {
	s.monitorMu.RLock()
	defer s.monitorMu.RUnlock()
	if after > s.monitorSequence {
		after = 0
	}
	filter := func(sample monitorSample) monitorSample {
		if netInterface != "" {
			network := make(map[string]psutil.NetworkIO, 1)
			if value, ok := sample.Network[netInterface]; ok {
				network[netInterface] = value
			}
			sample.Network = network
		}
		if diskName != "" {
			disk := make(map[string]psutil.DiskIO, 1)
			if value, ok := sample.Disk[diskName]; ok {
				disk[diskName] = value
			}
			sample.Disk = disk
		}
		return sample
	}
	current := filter(monitorSample{
		ID:      s.monitorSequence,
		Time:    s.monitorUpdatedAt,
		Network: s.netRateCache,
		Disk:    s.diskRateCache,
	})
	history := make([]monitorSample, 0)
	for _, sample := range s.monitorHistory {
		if sample.ID > after {
			history = append(history, filter(sample))
		}
	}
	return map[string]interface{}{
		"sample_id":   current.ID,
		"sample_time": current.Time,
		"network":     current.Network,
		"disk":        current.Disk,
		"history":     history,
	}
}

// startMonitor 启动系统监控
func (s *service) startMonitor() {
	// 首次同步更新，确保服务启动后即可读取监控数据
	s.updateMonitor()
	s.updateStaticMonitor()
	s.startMonitorHistory()
	s.startMonitorProcesses()

	s.wait.Add(2)
	go func() {
		defer s.wait.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateMonitor()
			case <-s.ctx.Done():
				return
			}
		}
	}()

	// 磁盘容量等静态信息单独刷新，避免慢挂载点阻塞实时采样
	go func() {
		defer s.wait.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateStaticMonitor()
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// updateMonitor 同步采集由 util 完成，service 负责发布缓存和记录采样历史。
func (s *service) updateMonitor() {
	snapshot := s.resourceSampler.Sample(s.ctx)
	network := make([]psutil.NetworkIO, 0, len(snapshot.Network))
	netReady, diskReady := false, false
	for _, value := range snapshot.Network {
		netReady = netReady || value.RateAvailable
		if !value.IsLoopback {
			network = append(network, value)
		}
	}
	snapshot.Network = network
	var netErr, diskErr error
	if message := snapshot.Errors["network"]; message != "" {
		netErr = errors.New(message)
	}
	if message := snapshot.Errors["disk"]; message != "" {
		diskErr = errors.New(message)
	}
	s.monitorMu.Lock()
	if netErr == nil {
		s.netRateCache = make(map[string]psutil.NetworkIO, len(network))
		for _, value := range network {
			s.netRateCache[value.Name] = value
		}
	}
	if diskErr == nil {
		s.diskRateCache = make(map[string]psutil.DiskIO, len(snapshot.Disk))
		for _, value := range snapshot.Disk {
			s.diskRateCache[value.Name] = value
			diskReady = diskReady || value.RateAvailable
		}
	}
	s.resourceCache = snapshot
	if netReady || diskReady {
		s.monitorUpdatedAt = snapshot.Time
		s.monitorSequence = max(s.monitorSequence+1, snapshot.Time)
		sample := monitorSample{
			ID:      s.monitorSequence,
			Time:    snapshot.Time,
			Network: s.netRateCache,
			Disk:    s.diskRateCache,
		}
		if len(s.monitorHistory) == 120 {
			copy(s.monitorHistory, s.monitorHistory[1:])
			s.monitorHistory[len(s.monitorHistory)-1] = sample
		} else {
			s.monitorHistory = append(s.monitorHistory, sample)
		}
	}
	s.monitorMu.Unlock()
	if netErr == nil && !netReady {
		netErr = errors.New("网卡计数器正在建立采样基线")
	}
	if diskErr == nil && !diskReady {
		diskErr = errors.New("磁盘计数器正在建立采样基线")
	}
	// 汇总业务监控字段，排队、聚合及落盘交给记录器。
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
	// 同一文件系统的多次挂载只计算一次。
	var diskUsed, diskTotal float64
	seen := make(map[string]bool, len(s.disksInfoCache))
	for _, disk := range s.disksInfoCache {
		key := disk.Filesystem
		if key == "" {
			key = disk.Path
		}
		if seen[key] || disk.Size.Total <= 0 {
			continue
		}
		seen[key] = true
		diskTotal += float64(disk.Size.Total)
		diskUsed += float64(max(0, min(disk.Size.Used, disk.Size.Total)))
	}
	values["disk_used"], values["disk_total"] = diskUsed, diskTotal
	if values["disk_total"] > 0 {
		values["disk_usage"] = values["disk_used"] / values["disk_total"] * 100
	}
	_ = recorder.Record(snapshot.Time, values)
}

// updateStaticMonitor 低频刷新主机、磁盘容量与汇总设备，避免慢挂载点阻塞实时采样。
func (s *service) updateStaticMonitor() {
	systemBase, _ := psutil.Host(s.ctx)
	diskInfo, diskErr := psutil.Disks(s.ctx)
	s.monitorMu.Lock()
	s.systemBaseCache = systemBase
	s.disksInfoCache = diskInfo
	s.monitorMu.Unlock()
	// 低频刷新汇总设备，避免每秒遍历设备目录。
	s.monitorMu.RLock()
	diskNames := make([]string, 0, len(s.diskRateCache))
	for name := range s.diskRateCache {
		diskNames = append(diskNames, name)
	}
	networkNames := make([]string, 0, len(s.netRateCache))
	for name := range s.netRateCache {
		networkNames = append(networkNames, name)
	}
	s.monitorMu.RUnlock()
	diskNames, networkNames = psutil.AggregateDevices(diskNames, networkNames)
	disks := make(map[string]bool, len(diskNames))
	for _, name := range diskNames {
		disks[name] = true
	}
	networks := make(map[string]bool, len(networkNames))
	for _, name := range networkNames {
		networks[name] = true
	}
	s.monitorMu.Lock()
	s.monitorDisks = disks
	s.monitorNetworks = networks
	s.monitorDiskReady = diskErr == nil
	s.monitorMu.Unlock()
}

// startMonitorProcesses 独立低频 worker，不向每秒主机采样路径增加逐进程开销。
func (s *service) startMonitorProcesses() {
	s.wait.Add(1)
	go func() {
		defer s.wait.Done()
		for {
			if s.ctx.Err() != nil {
				return
			}
			snapshot := s.processSampler.Sample(s.ctx)
			if s.ctx.Err() != nil {
				return
			}
			s.monitorMu.Lock()
			s.processCache = snapshot
			s.monitorMu.Unlock()
			if s.monitorProcessStore != nil && s.monitorRecorder != nil && s.monitorRecorder.Enabled() {
				s.monitorProcessStore.Record(snapshot)
			}
			// 采样完成后再等待，慢扫描不会积压 ticker 或并发启动下一轮。
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-s.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()
}

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
	configuration, err := func() (recordstore.SeriesConfiguration, error) {
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
	}()
	if err != nil {
		s.monitorInitError = err
		log.Printf("读取系统监控配置失败: %v", err)
		return
	}
	s.monitorRecorder, s.monitorInitError = recordstore.OpenRecorder(constant.MonitorPath, systemMonitorFields, recordstore.SeriesOptions{
		Configuration: &configuration,
		SaveConfiguration: func(configuration recordstore.SeriesConfiguration) error {
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
		},
	})
	if s.monitorInitError != nil {
		log.Printf("打开系统监控历史失败: %v", s.monitorInitError)
		return
	}
	s.monitorProcessStore, s.monitorProcessError = openMonitorProcessStore(s.ctx, filepath.Join(constant.MonitorPath, "processes"), configuration)
	if s.monitorProcessError != nil {
		log.Printf("打开监控进程历史失败: %v", s.monitorProcessError)
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
	settings, err := recorder.Settings()
	if err != nil {
		return settings, err
	}
	var processError string
	if s.monitorProcessError != nil {
		processError = s.monitorProcessError.Error()
	} else if s.monitorProcessStore != nil {
		var size int64
		size, processError = s.monitorProcessStore.Settings()
		settings.StorageBytes += size
	}
	if processError != "" {
		if settings.LastError != "" {
			settings.LastError += "；"
		}
		settings.LastError += "进程历史：" + processError
	}
	return settings, nil
}

func (s *service) SetMonitorSettings(enabled bool, days int) error {
	s.monitorSettingsMu.Lock()
	defer s.monitorSettingsMu.Unlock()
	recorder, err := s.monitorStorage()
	if err != nil {
		return err
	}
	if err := recorder.UpdateSettings(enabled, days); err != nil {
		return err
	}
	if s.monitorProcessStore != nil {
		s.monitorProcessStore.UpdateSettings(enabled, days)
	}
	return nil
}

// GetMonitorHistory 按展示宽度控制查询点数，省略时使用默认预算；原始分钟数据保持不变。
func (s *service) GetMonitorHistory(ctx context.Context, start, end int64, maxPoints ...int) (MonitorHistory, error) {
	if err := ctx.Err(); err != nil {
		return MonitorHistory{}, err
	}
	if len(maxPoints) > 1 {
		return MonitorHistory{}, errors.New("监控历史只接受一个查询点数")
	}
	maximum := monitorHistoryDefaultPoints
	if len(maxPoints) == 1 {
		maximum = maxPoints[0]
	}
	if maximum < monitorHistoryMinPoints || maximum > monitorHistoryMaxPoints {
		return MonitorHistory{}, fmt.Errorf("监控历史查询点数必须为 %d 至 %d", monitorHistoryMinPoints, monitorHistoryMaxPoints)
	}
	recorder, err := s.monitorStorage()
	if err != nil {
		return MonitorHistory{}, err
	}
	result, err := recorder.Query(start, end, recordstore.SeriesQueryOptions{MaxPoints: maximum})
	if err != nil {
		return MonitorHistory{}, err
	}
	// 从首条真实采样开始补齐时间桶；中间及末尾空桶仍以 samples=0 标记。
	result, err = func(result recordstore.RecordResult) (recordstore.RecordResult, error) {
		if result.Start <= 0 || result.End <= result.Start || result.Interval <= 0 || result.Interval > math.MaxInt64/1000 {
			return result, errors.New("监控历史时间桶无效")
		}
		step := result.Interval * 1000
		first, last := result.Start/step, result.End/step
		count := last - first + 1
		if count < 1 || count > monitorHistoryMaxPoints+1 {
			return result, errors.New("监控历史时间桶数量超出限制")
		}

		if len(result.Points) == 0 {
			return result, nil
		}
		// Query 已按时间排序并聚合，顺序归位即可，无需再建立桶索引。
		previous := int64(-1)
		for index, point := range result.Points {
			timestamp, ok := point["time"].(int64)
			if !ok || timestamp < result.Start || timestamp > result.End {
				return result, errors.New("监控历史采样时间无效")
			}
			bucket := timestamp / step
			if bucket <= previous {
				return result, errors.New("监控历史时间桶重复或顺序无效")
			}
			if index == 0 {
				first = bucket
			}
			previous = bucket
		}

		result.Start = result.Points[0]["time"].(int64)
		points := make([]map[string]interface{}, 0, int(last-first+1))
		index := 0
		for bucket := first; bucket <= last; bucket++ {
			timestamp := max(result.Start, bucket*step)
			if index < len(result.Points) && result.Points[index]["time"].(int64)/step == bucket {
				row := maps.Clone(result.Points[index])
				row["time"] = timestamp
				points = append(points, row)
				index++
				continue
			}
			row := make(map[string]interface{}, len(systemMonitorFields)+2)
			row["time"], row["samples"] = timestamp, uint32(0)
			for _, field := range systemMonitorFields {
				row[field.Name] = float64(0)
			}
			points = append(points, row)
		}
		result.Points = points
		return result, nil
	}(result)
	if err != nil {
		return MonitorHistory{}, err
	}
	history := MonitorHistory{Data: result.Points, Interval: result.Interval, Start: result.Start, End: result.End}
	for _, point := range history.Data {
		point["process"] = monitorProcessData{}
	}
	if len(history.Data) == 0 {
		return history, nil
	}
	if s.monitorProcessStore == nil {
		history.ProcessError = "进程历史存储不可用"
		return history, nil
	}
	step := history.Interval * 1000
	peaks, err := s.monitorProcessStore.history(ctx, history.Start, history.End, step)
	if err != nil {
		if ctx.Err() != nil {
			return MonitorHistory{}, ctx.Err()
		}
		history.ProcessError = err.Error()
		return history, nil
	}
	for _, point := range history.Data {
		if samples, ok := point["samples"].(uint32); ok && samples > 0 {
			// 上游已排序，只返回各榜单实际展示的指标与实际采样时刻。
			var groups [3]*monitorProcessRanking
			for index, group := range peaks[point["time"].(int64)/step*step] {
				if !group.Available || len(group.Items) == 0 {
					continue
				}
				ranking := &monitorProcessRanking{Time: group.Time, Partial: group.Partial, Items: make([]monitorProcessItem, 0, len(group.Items))}
				for _, item := range group.Items {
					ranking.Items = append(ranking.Items, monitorProcessItem{PID: item.PID, Name: item.Name, Value: processMetricValue(item, index)})
				}
				groups[index] = ranking
			}
			point["process"] = monitorProcessData{CPU: groups[0], Memory: groups[1], DiskIO: groups[2]}
		}
	}
	return history, nil
}

var processGzipWriters = sync.Pool{New: func() interface{} {
	writer, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
	return writer
}}

var processGzipReaders = sync.Pool{New: func() interface{} { return new(gzip.Reader) }}

var processGzipBuffers = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}

func openMonitorProcessStore(ctx context.Context, dir string, configuration recordstore.SeriesConfiguration) (*monitorProcessStore, error) {
	store, err := recordstore.Open(dir, recordstore.Options{MaxValueBytes: 64 << 10})
	if err != nil {
		return nil, err
	}
	collection, err := store.Collection("process-peaks-v1")
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	h := &monitorProcessStore{
		store: store, collection: collection, path: filepath.Join(dir, "records.db"), cancel: cancel,
		done: make(chan struct{}), wake: make(chan struct{}, 1), queries: make(chan struct{}, 2),
		pending: make(map[int64]pendingProcessPeaks), enabled: configuration.Enabled, days: configuration.RetentionDays,
	}
	go h.run(ctx)
	return h, nil
}

func (h *monitorProcessStore) Record(snapshot psutil.ProcessTopSnapshot) {
	if !snapshot.CPUAvailable && !snapshot.MemoryAvailable && !snapshot.IOAvailable {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || !h.enabled || snapshot.Time < h.acceptFrom || snapshot.Time <= 0 {
		return
	}
	minute := snapshot.Time / processMinuteMillis * processMinuteMillis
	pending, exists := h.pending[minute]
	// 慢盘最多积压 8 分钟；丢弃最旧未落盘记录，不能让内存随故障时长增长。
	if !exists && len(h.pending) == 8 {
		oldest := minute
		for timestamp := range h.pending {
			oldest = min(oldest, timestamp)
		}
		if oldest == minute {
			h.dropped++
			return
		}
		delete(h.pending, oldest)
		h.dropped++
	}
	lists := [3][]psutil.ProcessTopItem{snapshot.CPU, snapshot.Memory, snapshot.IO}
	available := [3]bool{snapshot.CPUAvailable, snapshot.MemoryAvailable, snapshot.IOAvailable}
	for index, items := range lists {
		group := MonitorProcessGroup{Time: snapshot.Time, Items: items, Available: available[index], Partial: snapshot.Partial}
		if betterProcessPeak(group, pending.peaks[index], index) {
			group.Items = append([]psutil.ProcessTopItem{}, items[:min(len(items), 5)]...)
			pending.peaks[index] = group
		}
	}
	pending.version++
	h.pending[minute] = pending
}

func processPeakScore(group MonitorProcessGroup, metric int) float64 {
	var score float64
	for _, item := range group.Items {
		score += processMetricValue(item, metric)
	}
	return score
}

func betterProcessPeak(next, previous MonitorProcessGroup, metric int) bool {
	if !next.Available {
		return false
	}
	if !previous.Available {
		return true
	}
	a, b := processPeakScore(next, metric), processPeakScore(previous, metric)
	return a > b || a == b && next.Time > previous.Time
}

// UpdateSettings 不执行文件操作；启用边界排除上一轮在途采样，后台保留已接受的历史。
func (h *monitorProcessStore) UpdateSettings(enabled bool, days int) {
	h.mu.Lock()
	if h.enabled != enabled {
		h.acceptFrom = time.Now().UnixMilli()
	}
	h.enabled, h.days = enabled, days
	h.mu.Unlock()
	select {
	case h.wake <- struct{}{}:
	default:
	}
}

func (h *monitorProcessStore) run(ctx context.Context) {
	defer close(h.done)
	timer := time.NewTimer(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
	defer timer.Stop()
	var cleanedAt time.Time
	for {
		cleanup := time.Since(cleanedAt) >= time.Hour
		select {
		case <-ctx.Done():
			h.mu.Lock()
			h.closed = true
			h.mu.Unlock()
			h.closeError = errors.Join(h.flush(true), h.store.Close())
			return
		case <-h.wake:
			cleanup = true
		case <-timer.C:
			timer.Reset(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
		}
		err := h.flush(false)
		if cleanup {
			h.mu.Lock()
			days := h.days
			h.mu.Unlock()
			_, cleanupErr := h.collection.DeleteBefore(ctx, time.Now().UnixMilli()-int64(days)*processDayMillis)
			err = errors.Join(err, cleanupErr)
			cleanedAt = time.Now()
		}
		h.mu.Lock()
		h.lastError = err
		h.mu.Unlock()
	}
}

func (h *monitorProcessStore) flush(all bool) error {
	now := time.Now().UnixMilli()
	h.mu.Lock()
	cutoff := now - int64(h.days)*processDayMillis
	pending := make(map[int64]pendingProcessPeaks, len(h.pending))
	for minute, value := range h.pending {
		if minute < cutoff {
			delete(h.pending, minute)
		} else if all || !h.enabled || minute < now/processMinuteMillis*processMinuteMillis {
			pending[minute] = value
		}
	}
	h.mu.Unlock()
	values := make(map[string]processPeaks, len(pending)*2)
	times := make(map[string]int64, len(pending)*2)
	for minute, value := range pending {
		// 重启、开关或慢盘重试可能再次写入同一分钟/小时；与已保存峰值合并。
		for _, prefix := range []string{"m", "h"} {
			timestamp := minute
			if prefix == "h" {
				timestamp = minute / processHourMillis * processHourMillis
			}
			key := prefix + strconv.FormatInt(timestamp, 10)
			peaks, exists := values[key]
			if !exists {
				record, err := h.collection.Get(key)
				if err == nil {
					peaks, err = decodeProcessPeaks(record)
				}
				if err != nil && !errors.Is(err, recordstore.ErrNotFound) {
					return err
				}
			}
			for index, group := range value.peaks {
				if betterProcessPeak(group, peaks[index], index) {
					peaks[index] = group
				}
			}
			values[key], times[key] = peaks, timestamp
		}
	}
	entries := make([]recordstore.Entry, 0, len(values))
	for key, peaks := range values {
		data, err := func() ([]byte, error) {
			var buffer bytes.Buffer
			writer := processGzipWriters.Get().(*gzip.Writer)
			defer processGzipWriters.Put(writer)
			writer.Reset(&buffer)
			if err := json.NewEncoder(writer).Encode(peaks); err != nil {
				_ = writer.Close()
				return nil, err
			}
			if err := writer.Close(); err != nil {
				return nil, err
			}
			return buffer.Bytes(), nil
		}()
		if err != nil {
			return err
		}
		entries = append(entries, recordstore.Entry{Key: key, Value: data, Options: recordstore.WriteOptions{Time: times[key]}})
	}
	if _, err := h.collection.PutMany(entries); err != nil {
		return err
	}
	h.mu.Lock()
	for minute, value := range pending {
		if h.pending[minute].version == value.version {
			delete(h.pending, minute)
		}
	}
	h.mu.Unlock()
	return nil
}

func decodeProcessPeaks(record recordstore.Record) (processPeaks, error) {
	var peaks processPeaks
	if record.Encoding != recordstore.EncodingJSON {
		return peaks, errors.New("原始字节记录不能作为 JSON 解码")
	}
	var compressed []byte
	// 此处是固定的 JSON 字节串，不需要通用 Decode 的流式解码器及 UseNumber。
	if err := json.Unmarshal(record.Data, &compressed); err != nil {
		return peaks, err
	}
	reader := processGzipReaders.Get().(*gzip.Reader)
	defer processGzipReaders.Put(reader)
	if err := reader.Reset(bytes.NewReader(compressed)); err != nil {
		return peaks, err
	}
	defer reader.Close()
	// 复用解压缓冲区，长区间查询不为每条分钟记录重复扩容；仍读到 EOF 校验 CRC。
	buffer := processGzipBuffers.Get().(*bytes.Buffer)
	defer processGzipBuffers.Put(buffer)
	buffer.Reset()
	_, err := buffer.ReadFrom(io.LimitReader(reader, (64<<10)+1))
	if err != nil {
		return peaks, err
	}
	if buffer.Len() > 64<<10 {
		return peaks, errors.New("进程历史记录大小超过限制")
	}
	err = json.Unmarshal(buffer.Bytes(), &peaks)
	return peaks, err
}

func (h *monitorProcessStore) Settings() (int64, string) {
	h.mu.Lock()
	err, dropped := h.lastError, h.dropped
	h.mu.Unlock()
	info, statErr := os.Stat(h.path)
	message := ""
	if failure := errors.Join(err, statErr); failure != nil {
		message = failure.Error()
	}
	if dropped > 0 {
		if message != "" {
			message += "；"
		}
		message += fmt.Sprintf("累计跳过 %d 个进程历史分钟桶", dropped)
	}
	if info == nil {
		return 0, message
	}
	return info.Size(), message
}

func (h *monitorProcessStore) Close() error {
	h.cancel()
	<-h.done
	return h.closeError
}

// history 一次聚合整个历史区间；每条源记录仅解压一次，同时归并三个指标。
// 返回键是 UTC 时间桶起点，不是某个指标的实际峰值时间。
func (h *monitorProcessStore) history(ctx context.Context, start, end, step int64) (map[int64]processPeaks, error) {
	result := make(map[int64]processPeaks)
	if start <= 0 || end < start || end-start > 365*processDayMillis || end > time.Now().Add(time.Minute).UnixMilli() ||
		step < processMinuteMillis || step > 365*processDayMillis || step%processMinuteMillis != 0 ||
		(step > processHourMillis && step%processHourMillis != 0) || end/step-start/step+1 > monitorHistoryMaxPoints+1 {
		return result, errors.New("进程历史查询时间桶无效")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	select {
	case h.queries <- struct{}{}:
		defer func() { <-h.queries }()
	case <-ctx.Done():
		return result, ctx.Err()
	}
	h.mu.Lock()
	start = max(start, time.Now().UnixMilli()-int64(h.days)*processDayMillis)
	pending := make([]processPeaks, 0, len(h.pending))
	for minute, value := range h.pending {
		if minute <= end && minute+processMinuteMillis > start {
			pending = append(pending, value.peaks)
		}
	}
	h.mu.Unlock()
	if start > end {
		return result, nil
	}
	accept := func(peaks processPeaks) {
		for metric, group := range peaks {
			if !group.Available || group.Time < start || group.Time > end {
				continue
			}
			bucket := group.Time / step * step
			previous := result[bucket]
			if betterProcessPeak(group, previous[metric], metric) {
				previous[metric] = group
				result[bucket] = previous
			}
		}
	}
	scanMinutes := func(from, to int64) error {
		if from > to {
			return nil
		}
		return h.collection.Scan(ctx, recordstore.Query{Prefix: "m", Start: from / processMinuteMillis * processMinuteMillis, End: to, Limit: 128}, func(record recordstore.Record) error {
			peaks, err := decodeProcessPeaks(record)
			if err == nil {
				accept(peaks)
			}
			return err
		})
	}
	firstHour := (start + processHourMillis - 1) / processHourMillis * processHourMillis
	lastHour := (end + 1) / processHourMillis * processHourMillis
	if step >= processHourMillis && firstHour < lastHour {
		// 大桶整段直取小时主键，全年最多 8760 条；不遍历其间的全部分钟索引。
		for hour := firstHour; hour < lastHour; hour += processHourMillis {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			record, err := h.collection.Get("h" + strconv.FormatInt(hour, 10))
			if errors.Is(err, recordstore.ErrNotFound) {
				// 缩短保留期可能删掉边界小时汇总但留下部分分钟；扩大保留期后仍可查询这些记录。
				if err := scanMinutes(hour, hour+processHourMillis-1); err != nil {
					return nil, err
				}
				continue
			}
			if err != nil {
				return nil, err
			}
			peaks, err := decodeProcessPeaks(record)
			if err != nil {
				return nil, err
			}
			accept(peaks)
		}
		if err := scanMinutes(start, firstHour-1); err != nil {
			return nil, err
		}
		if err := scanMinutes(lastHour, end); err != nil {
			return nil, err
		}
	} else if err := scanMinutes(start, end); err != nil {
		return nil, err
	}
	for _, peaks := range pending {
		accept(peaks)
	}
	for bucket, peaks := range result {
		for metric, group := range peaks {
			// 克隆最终榜单并限五项，兼容旧记录；不修改共享缓存。
			peaks[metric] = func() MonitorProcessGroup {
				if !group.Available {
					return MonitorProcessGroup{}
				}
				items := make([]psutil.ProcessTopItem, 0, 5)
				for _, item := range group.Items[:min(len(group.Items), 5)] {
					value := processMetricValue(item, metric)
					if item.PID <= 0 || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 ||
						(metric == 0 && !item.CPUAvailable) || (metric == 2 && (!item.IOAvailable || item.ReadRate < 0 || item.WriteRate < 0)) {
						continue
					}
					items = append(items, item)
				}
				slices.SortFunc(items, func(a, b psutil.ProcessTopItem) int {
					left, right := processMetricValue(a, metric), processMetricValue(b, metric)
					if left > right {
						return -1
					} else if left < right {
						return 1
					}
					return int(a.PID) - int(b.PID)
				})
				if len(items) == 0 {
					return MonitorProcessGroup{}
				}
				group.Items = items
				return group
			}()
		}
		result[bucket] = peaks
	}
	return result, ctx.Err()
}

func processMetricValue(item psutil.ProcessTopItem, metric int) float64 {
	switch metric {
	case 0:
		return item.CPUPercent
	case 1:
		return float64(item.MemoryRSS)
	default:
		return item.ReadRate + item.WriteRate
	}
}
