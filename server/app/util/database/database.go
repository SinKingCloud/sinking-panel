package database

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// Database 数据库连接
type Database struct {
	Db      *gorm.DB
	DbError error
}

func newLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), //io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second,   // 慢 SQL 阈值
			LogLevel:                  logger.Silent, // 日志级别
			IgnoreRecordNotFoundError: true,          // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,         // 禁用彩色打印
		},
	)
}

// NewMysql 实例化一个mysql连接
func NewMysql(host string, port string, user string, pwd string, database string) *Database {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pwd, host, port, database)
	Db, DbError := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger(),
	})
	if DbError != nil {
		return &Database{Db: Db, DbError: errors.New("sql connect error")}
	}
	sqlDB, err := Db.DB()
	if err != nil {
		return &Database{Db: Db, DbError: errors.New("set sql run config failed")}
	}
	sqlDB.SetMaxIdleConns(1000)
	sqlDB.SetMaxOpenConns(10000)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	return &Database{Db: Db, DbError: DbError}
}

// NewSqlite 实例化一个sqlite连接
func NewSqlite(file string) *Database {
	Db, DbError := gorm.Open(sqlite.Open(file), &gorm.Config{
		Logger: newLogger(),
	})
	if DbError != nil {
		return &Database{Db: Db, DbError: errors.New("sql connect error")}
	}
	return &Database{Db: Db, DbError: DbError}
}
