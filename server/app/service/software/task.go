package software

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"server/app/enum/software_state"
	"server/app/enum/software_status"
	"server/app/enum/system_task_status"
	"server/app/model"
	repositorySoftware "server/app/repository/software"
	"server/app/util/plugin"
	"server/app/util/str"
)

// Install 只读取清单并提交任务；大文件打包、解压和插件生命周期均在后台执行。
func (s *service) Install(path string) (int64, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	var input io.ReadCloser
	if info.IsDir() {
		input, err = os.Open(filepath.Join(path, "software.json"))
	} else if info.Mode().IsRegular() {
		archive, openErr := zip.OpenReader(path)
		if openErr != nil {
			return 0, openErr
		}
		defer archive.Close()
		for _, file := range archive.File {
			if file.Name == "software.json" {
				input, err = file.Open()
				break
			}
		}
	} else {
		return 0, errors.New("软件包必须是普通 ZIP 文件或目录")
	}
	if err != nil {
		return 0, err
	}
	if input == nil {
		return 0, errors.New("软件包缺少 software.json")
	}
	defer input.Close()
	var manifest plugin.Manifest
	if err = json.NewDecoder(io.LimitReader(input, 256<<10)).Decode(&manifest); err != nil {
		return 0, fmt.Errorf("读取软件清单失败: %w", err)
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`).MatchString(manifest.Id) || strings.TrimSpace(manifest.Name) == "" || len(manifest.Name) > 100 || strings.TrimSpace(manifest.Version) == "" || len(manifest.Version) > 50 {
		return 0, errors.New("软件标识、名称或版本不合法")
	}
	// 全局锁仅覆盖登记阶段，耗时安装按软件互斥，不阻塞其他软件。
	s.importing.Lock()
	defer s.importing.Unlock()
	record, err := s.repository.FindByCode(manifest.Id)
	if err != nil {
		return 0, err
	}
	isNew := record == nil
	if isNew {
		record = &model.Software{
			Id:      str.GetSnowWorkIns().GetId(),
			Name:    manifest.Name,
			Code:    manifest.Id,
			Version: manifest.Version,
			Status:  software_status.Enabled,
			State:   software_state.Installing,
			Config:  "{}",
		}
	}
	e, err := s.entry(record.Id)
	if err != nil {
		return 0, err
	}
	if !e.mu.TryLock() {
		return 0, errors.New("软件正在执行操作，请稍后重试")
	}
	defer e.mu.Unlock()
	if e.busy.Load() {
		return 0, errors.New("软件任务正在执行")
	}
	// entry 可能恢复已取消的任务，重新读取其记录。
	current, err := s.repository.FindById(record.Id)
	if err != nil {
		return 0, err
	}
	if current == nil {
		isNew = true
		record = &model.Software{
			Id:      record.Id,
			Name:    manifest.Name,
			Code:    manifest.Id,
			Version: manifest.Version,
			Status:  software_status.Enabled,
			State:   software_state.Installing,
			Config:  "{}",
		}
	} else {
		record = current
	}
	if !isNew && record.State != software_state.Installed {
		return 0, errors.New("软件尚未就绪，请先处理未完成的任务")
	}
	if _, err = s.dependencies(&manifest); err != nil {
		return 0, err
	}
	// 升级提交后旧代码可能仍在运行，回滚也需要旧依赖，完成前同时保留两版依赖。
	reserved := manifest
	reserved.Requires = slices.Clone(manifest.Requires)
	if !isNew {
		previous, err := s.manifest(record.Code)
		if err != nil {
			return 0, err
		}
		for _, code := range previous.Requires {
			if !slices.Contains(reserved.Requires, code) {
				reserved.Requires = append(reserved.Requires, code)
			}
		}
	}
	if isNew {
		err = s.repository.Create(record)
	} else {
		state := software_state.Upgrading
		err = s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{State: &state})
	}
	s.invalidate(record.Code)
	if err != nil {
		return 0, err
	}
	taskID := str.GetSnowWorkIns().GetId()
	key := strconv.FormatInt(taskID, 10)
	e.created, e.pending, e.taskID = isNew, record, key
	s.mu.Lock()
	s.manifests[record.Code] = &reserved
	s.mu.Unlock()
	e.busy.Store(true)
	s.system.TaskCreate(key, "安装 "+record.Name, map[string]interface{}{"software_id": record.Id}, func(ctx context.Context, _ interface{}, progress func(int, float64, string)) {
		e.mu.Lock()
		defer e.mu.Unlock()
		if e.taskID != key {
			return
		}
		if ctx.Err() != nil {
			s.system.TaskCancel(key)
			if err := s.restoreTask(e); err != nil {
				progress(system_task_status.Failed, 100, err.Error())
			}
			return
		}
		e.pending = nil
		defer e.busy.Store(false)
		defer s.invalidate(record.Code)
		defer s.invalidateManifest(record.Code)
		progress(system_task_status.Running, 10, "正在准备软件包并执行安装")
		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		// 生命周期可能回写内部配置，失败时由 install 恢复代码及配置，再恢复安装状态。
		runErr := s.install(ctx, path, record, isNew)
		if runErr != nil {
			var restoreErr error
			if isNew {
				restoreErr = s.repository.DeleteById(record.Id)
			} else {
				state := software_state.Installed
				restoreErr = s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{State: &state})
			}
			progress(system_task_status.Failed, 100, errors.Join(runErr, restoreErr).Error())
			return
		}
		progress(system_task_status.Completed, 100, "安装完成")
	})
	if s.system.GetTask(key) == nil {
		return 0, errors.Join(errors.New("创建后台任务失败，面板正在停止"), s.restoreTask(e))
	}
	return taskID, nil
}
