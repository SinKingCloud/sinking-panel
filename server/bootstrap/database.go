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
	if err := global.App.Database.SyncSqlite(public.Sql); err != nil {
		panic(err)
	}
}
