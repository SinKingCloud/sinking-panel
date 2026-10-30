package software

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"server/app/enum/software_state"
	"server/app/enum/software_status"
	"server/app/model"
	repositorySoftware "server/app/repository/software"
)

// install 在后台完成打包、解释、生命周期及代码替换；调用方持有软件独占锁。
func (s *service) install(ctx context.Context, path string, record *model.Software, isNew bool) (err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		// 仅打包运行所需内容，根目录中的开发 SDK 和 Git 元数据不随软件安装。
		archive, err := os.CreateTemp("", "sinking-plugin-*.zip")
		if err != nil {
			return err
		}
		defer os.Remove(archive.Name())
		err = s.packDirectory(ctx, path, archive)
		err = errors.Join(err, archive.Close())
		if err != nil {
			return err
		}
		path = archive.Name()
	} else if !info.Mode().IsRegular() {
		return errors.New("软件包必须是普通 ZIP 文件或目录")
	}

	pkg, err := s.plugins.Import(ctx, path)
	if err != nil {
		return err
	}
	if pkg.Manifest.Id != record.Code {
		_ = s.plugins.Remove(pkg.Directory)
		return errors.New("软件包标识在提交任务后发生变动，请重新安装")
	}
	if err = ctx.Err(); err != nil {
		_ = s.plugins.Remove(pkg.Directory)
		return err
	}
	staging := pkg.Directory
	defer func() {
		if cleanupErr := s.plugins.Remove(staging); cleanupErr != nil {
			log.Printf("清理软件临时目录失败: %v", cleanupErr)
		}
	}()
	if !pkg.Available {
		return errors.New(pkg.UnavailableReason)
	}
	dependencies, err := s.dependencies(pkg.Manifest)
	if err != nil {
		return err
	}
	if s.booted.Load() {
		release, err := s.start(dependencies)
		if err != nil {
			return err
		}
		defer release()
	}
	directory, backup := pkg.Manifest.Id, ".backup-"+pkg.Manifest.Id
	committed, replaced, backedUp := false, false, false
	var previousConfig string
	restorePrevious := s.booted.Load() && !isNew && record.Status == software_status.Enabled
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
		if !isNew && replaced {
			if restoreErr := s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{Config: &previousConfig}); restoreErr != nil {
				err = errors.Join(err, fmt.Errorf("恢复软件配置失败: %w", restoreErr))
				return
			}
		}
		s.invalidate(record.Code)
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
			return err
		}
		backedUp = true
		// Stop 可能保存最后的运行配置，快照应在旧版本停止后读取。
		previousConfig, err = s.GetConfig(record.Code)
		if err != nil {
			return err
		}
	}
	if err = s.plugins.Rename(staging, directory); err != nil {
		return err
	}
	replaced = true
	pkg.Directory = directory
	if err = ctx.Err(); err != nil {
		return err
	}
	if !isNew {
		params, _ := json.Marshal(map[string]string{
			"previous_version": record.Version,
			"version":          pkg.Manifest.Version,
		})
		if _, err = s.plugins.Lifecycle(ctx, pkg.Directory, "upgrade", string(params)); err != nil {
			return err
		}
	} else {
		if _, err = s.plugins.Lifecycle(ctx, pkg.Directory, "install", "{}"); err != nil {
			return err
		}
	}
	if s.booted.Load() && record.Status == software_status.Enabled {
		startup, release := context.WithTimeout(ctx, 30*time.Second)
		_, err = s.plugins.Lifecycle(startup, pkg.Directory, "start", "{}")
		release()
		if err != nil {
			return err
		}
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// 只提交安装元数据，不覆盖生命周期通过 SetConfig 保存的插件配置。
	record.State = software_state.Installed
	record.Version = pkg.Manifest.Version
	err = s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{
		Version: &record.Version,
		State:   &record.State,
	})
	s.invalidate(record.Code)
	if err != nil {
		return err
	}
	committed = true
	s.mu.Lock()
	s.icons[record.Code] = pkg.Manifest.Icon
	s.mu.Unlock()
	if backedUp {
		if err = s.plugins.Remove(backup); err != nil {
			log.Printf("软件已更新，但清理旧代码失败: %v", err)
		}
	}
	return nil
}
