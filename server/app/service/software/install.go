package software

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"server/app/enum/software_status"
	"server/app/model"
	repositorySoftware "server/app/repository/software"
	"server/app/util/str"
)

// Install 从 ZIP 或源码目录安装、升级软件，统一完成打包、校验与插件生命周期。
func (s *service) Install(path string) (result *Info, err error) {
	if !s.importing.TryLock() {
		return nil, errors.New("软件正在安装，请稍后重试")
	}
	defer s.importing.Unlock()

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		// 仅打包运行所需内容，根目录中的开发 SDK 和 Git 元数据不随软件安装。
		archive, err := os.CreateTemp("", "sinking-plugin-*.zip")
		if err != nil {
			return nil, err
		}
		defer os.Remove(archive.Name())
		err = packDirectory(path, archive)
		err = errors.Join(err, archive.Close())
		if err != nil {
			return nil, err
		}
		path = archive.Name()
	} else if !info.Mode().IsRegular() {
		return nil, errors.New("软件包必须是普通 ZIP 文件或目录")
	}

	pkg, err := s.plugins.Import(path)
	if err != nil {
		return nil, err
	}
	staging := pkg.Directory
	defer func() {
		err = errors.Join(err, s.plugins.Remove(staging))
	}()
	if !pkg.Available {
		return nil, errors.New(pkg.UnavailableReason)
	}
	dependencies, err := s.dependencies(pkg.Manifest)
	if err != nil {
		return nil, err
	}
	if s.booted.Load() {
		if err = s.start(dependencies); err != nil {
			return nil, err
		}
	}
	record, err := s.repository.FindByCode(pkg.Manifest.Id)
	if err != nil {
		return nil, err
	}
	isNew := record == nil
	if isNew {
		record = &model.Software{
			Id:      str.GetSnowWorkIns().GetId(),
			Name:    pkg.Manifest.Name,
			Code:    pkg.Manifest.Id,
			Version: pkg.Manifest.Version,
			Status:  software_status.Installed,
			Config:  "{}",
		}
	}
	e, err := s.entry(record.Id)
	if err != nil {
		return nil, err
	}
	if !e.mu.TryLock() {
		return nil, errors.New("软件正在执行操作，请稍后重试")
	}
	defer e.mu.Unlock()
	if e.busy.Load() {
		return nil, errors.New("软件任务正在执行，暂时不能更新软件包")
	}
	if !isNew {
		record, err = s.repository.FindById(record.Id)
		if err != nil {
			return nil, err
		}
		if record == nil {
			return nil, errors.New("软件不存在")
		}
	}
	directory, backup := pkg.Manifest.Id, ".backup-"+pkg.Manifest.Id
	committed, replaced, backedUp, created := false, false, false, false
	var previousConfig string
	restorePrevious := s.booted.Load() && !isNew && record.Status == software_status.Installed
	// 临时备份只用于本次回滚，正式代码始终放在插件标识目录。
	defer func() {
		if committed {
			return
		}
		if replaced {
			if removeErr := s.plugins.Remove(directory); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("清理未完成安装的插件失败: %w", removeErr))
				return
			}
		}
		// 先停止新代码，再恢复配置，旧版本启动时只能读到升级前的数据。
		if created {
			err = errors.Join(err, s.repository.DeleteById(record.Id))
		} else if !isNew && replaced {
			if restoreErr := s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{Config: &previousConfig}); restoreErr != nil {
				err = errors.Join(err, fmt.Errorf("恢复软件配置失败: %w", restoreErr))
				return
			}
		}
		if backedUp {
			if restoreErr := s.plugins.Rename(backup, directory); restoreErr != nil {
				err = errors.Join(err, fmt.Errorf("恢复原插件目录失败: %w", restoreErr))
				return
			}
		}
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if restorePrevious {
			_, restoreErr := s.plugins.Lifecycle(cleanup, directory, "start", "{}")
			err = errors.Join(err, restoreErr)
		}
	}()
	if !isNew {
		if err = s.plugins.Rename(directory, backup); err != nil {
			return nil, err
		}
		backedUp = true
		// Stop 可能保存最后的运行配置，快照应在旧版本停止后读取。
		previousConfig, err = s.GetConfig(record.Code)
		if err != nil {
			return nil, err
		}
	}
	if err = s.plugins.Rename(staging, directory); err != nil {
		return nil, err
	}
	replaced = true
	pkg.Directory = directory
	// 先建立内部配置记录，Install/Start 可直接读写；失败时回滚记录和代码。
	if isNew {
		if err = s.repository.Create(record); err != nil {
			return nil, err
		}
		created = true
	}
	// 内置镜像首次解包和导入需要比纯源码安装更长的处理时间。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if !isNew {
		params, _ := json.Marshal(map[string]string{
			"previous_version": record.Version,
			"version":          pkg.Manifest.Version,
		})
		if _, err = s.plugins.Lifecycle(ctx, pkg.Directory, "upgrade", string(params)); err != nil {
			return nil, err
		}
	} else {
		if _, err = s.plugins.Lifecycle(ctx, pkg.Directory, "install", "{}"); err != nil {
			return nil, err
		}
	}
	if s.booted.Load() {
		startup, release := context.WithTimeout(ctx, 30*time.Second)
		_, err = s.plugins.Lifecycle(startup, pkg.Directory, "start", "{}")
		release()
		if err != nil {
			return nil, err
		}
	}
	// 只提交安装元数据，不覆盖生命周期通过 SetConfig 保存的插件配置。
	record.Status = software_status.Installed
	record.Version = pkg.Manifest.Version
	if !isNew {
		err = s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{
			Version: &record.Version,
			Status:  &record.Status,
		})
	}
	if err != nil {
		return nil, err
	}
	committed = true
	s.mu.Lock()
	s.icons[record.Code] = pkg.Manifest.Icon
	s.mu.Unlock()
	e.taskID = ""
	if backedUp {
		if err = s.plugins.Remove(backup); err != nil {
			return nil, fmt.Errorf("软件已更新，但清理旧代码失败: %w", err)
		}
	}
	return s.Info(record.Id)
}
