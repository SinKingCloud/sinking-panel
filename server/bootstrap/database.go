package bootstrap

import (
	"path/filepath"
	"server/app/constant"
	"server/app/util/database"
	"server/app/util/file"
	"server/global"
	"server/public"
)

// LoadDatabase 初始化数据库
func LoadDatabase() {
	if global.App.Database != nil {
		return
	}
	global.App.SetDataBase(database.NewSqlite(getDbFile()))
	if global.App.Database.DbError != nil {
		panic(global.App.Database.DbError)
	}
	if err := global.App.Database.SyncSqlite(public.Sql); err != nil {
		panic(err)
	}
}

// 获取并初始化db文件
func getDbFile() string {
	path := constant.DBPath
	f := file.NewDisk(path)
	_ = f.AutoCreate(constant.DBFile)
	return filepath.Join(path, constant.DBFile)
}
