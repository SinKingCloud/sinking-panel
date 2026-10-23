package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/util/database"
	"server/global"
	"server/public"
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
}
