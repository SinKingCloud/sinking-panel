package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/util/database"
	"server/app/util/str"
	"server/global"
	"server/public"
	"strings"

	"gorm.io/gorm"
)

// LoadDatabase 初始化数据库
func LoadDatabase() {
	if global.App.Database != nil {
		return
	}
	if err := os.MkdirAll(constant.DBPath, 0755); err != nil {
		panic(fmt.Errorf("创建数据库目录失败: %w", err))
	}
	global.App.SetDataBase(database.NewSqlite(filepath.Join(constant.DBPath, constant.DBFile)))
	if global.App.Database.DbError != nil {
		panic(global.App.Database.DbError)
	}
	// 先重命名实例表，保留原 ID 和加密配置，再同步新的单表结构。
	migrator := global.App.Database.Db.Migrator()
	if migrator.HasTable("cloud_database_servers") {
		if migrator.HasTable("cloud_databases") {
			panic("检测到两张数据库实例表，停止结构同步以避免覆盖连接信息")
		}
		if err := migrator.RenameTable("cloud_database_servers", "cloud_databases"); err != nil {
			panic(fmt.Errorf("重命名数据库实例表失败: %w", err))
		}
	}
	if err := global.App.Database.SyncSqlite(public.Sql); err != nil {
		panic(err)
	}
	// 数据库脚本和分类统一使用 database 或 database_实例ID 保存范围。
	engines := []string{"mysql", "postgres", "sqlserver", "sqlite", "redis", "mongo", "elasticsearch"}
	conditions := []string{"module = 'script'"}
	for _, engine := range engines {
		conditions = append(conditions, fmt.Sprintf("module = '%s' OR module GLOB '%s_*'", engine, engine))
	}
	legacyModules := strings.Join(conditions, " OR ")
	var pending bool
	if err := global.App.Database.Db.Raw("SELECT EXISTS(SELECT 1 FROM cloud_scripts WHERE " + legacyModules + ") OR EXISTS(SELECT 1 FROM cloud_types WHERE " + legacyModules + ")").Scan(&pending).Error; err != nil {
		panic(fmt.Errorf("检查脚本模块失败: %w", err))
	}
	if !pending {
		return
	}
	if err := global.App.Database.Transaction(func(tx *gorm.DB) error {
		for _, engine := range engines {
			module := gorm.Expr("? || substr(module, ?)", type_module.Database, len(engine)+1)
			if err := tx.Model(&model.Script{}).Where("module = ? OR module GLOB ?", engine, engine+"_*").
				UpdateColumn("module", module).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Type{}).Where("module = ? OR module GLOB ?", engine, engine+"_*").UpdateColumn("module", module).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.Script{}).Where("module = ?", "script").UpdateColumn("module", type_module.Ssh).Error; err != nil {
			return err
		}
		// 旧的共用分类只有同时被终端和数据库引用时才复制，保留原名称和时间。
		var categories []*model.Type
		if err := tx.Where("module = ?", "script").Find(&categories).Error; err != nil || len(categories) == 0 {
			return err
		}
		var scopes []struct {
			TypeId int64
			Module string
		}
		if err := tx.Table("cloud_scripts AS scripts").Distinct("scripts.type_id", "scripts.module").
			Joins("JOIN cloud_types AS types ON types.id = scripts.type_id").Where("types.module = ?", "script").Find(&scopes).Error; err != nil {
			return err
		}
		usage := make(map[int64]map[string]bool)
		for _, scope := range scopes {
			module, _, _ := strings.Cut(scope.Module, "_")
			if usage[scope.TypeId] == nil {
				usage[scope.TypeId] = make(map[string]bool)
			}
			usage[scope.TypeId][module] = true
		}
		categoryIds := make(map[string][]int64)
		for _, category := range categories {
			module := type_module.Ssh
			if usage[category.Id][type_module.Database] {
				if !usage[category.Id][type_module.Ssh] {
					module = type_module.Database
				} else {
					copy := *category
					copy.Id = str.GetSnowWorkIns().GetId()
					copy.Module = type_module.Database
					if err := tx.Session(&gorm.Session{SkipHooks: true}).Create(&copy).Error; err != nil {
						return err
					}
					if err := tx.Model(&model.Script{}).Where("type_id = ? AND (module = ? OR module GLOB ?)", category.Id, type_module.Database, "database_*").UpdateColumn("type_id", copy.Id).Error; err != nil {
						return err
					}
				}
			}
			categoryIds[module] = append(categoryIds[module], category.Id)
		}
		for module, ids := range categoryIds {
			if err := global.App.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
				return tx.Model(&model.Type{}).Where("id IN ?", batch.([]int64)).UpdateColumn("module", module).Error
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("迁移脚本模块失败: %w", err))
	}
}
