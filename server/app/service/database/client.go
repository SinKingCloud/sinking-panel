package database

import (
	"context"
	"encoding/json"
	"errors"
	"server/app/enum/database_kind"
	clientElasticsearch "server/app/service/database/elasticsearch"
	clientMongo "server/app/service/database/mongo"
	clientMySQL "server/app/service/database/mysql"
	clientPostgres "server/app/service/database/postgres"
	clientRedis "server/app/service/database/redis"
	clientSqlite "server/app/service/database/sqlite"
	clientSqlServer "server/app/service/database/sqlserver"
	"strings"
	"time"
)

// Test 解密提交的凭据并测试连接，整个过程最多 5 秒，不修改实例记录。
func (s *service) Test(ctx context.Context, id int64, databaseType int, config string) (result map[string]interface{}, err error) {
	// 面板类型由业务层补充，数据库管理器不依赖面板枚举。
	defer func() {
		if result != nil {
			result["type"] = databaseType
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if strings.TrimSpace(config) == "" {
		return nil, errors.New("请填写连接配置")
	}
	if id < 0 {
		return nil, errors.New("实例 ID 不合法")
	}
	stored := ""
	if id > 0 {
		data, err := s.repository.FindById(id)
		if err != nil {
			return nil, err
		}
		if data == nil || data.Id <= 0 {
			return nil, errors.New("实例不存在")
		}
		if data.Type != databaseType {
			return nil, errors.New("数据库实例类型与当前操作不匹配")
		}
		stored = data.Config
	}
	values, err := s.Config(databaseType, config, id, stored, "decrypt")
	if err != nil {
		return nil, err
	}
	content, _ := json.Marshal(values)
	config = string(content)
	switch databaseType {
	case database_kind.MySQL:
		manager, err := clientMySQL.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Test(ctx)
	case database_kind.Postgres:
		manager, err := clientPostgres.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Test(ctx)
	case database_kind.SqlServer:
		manager, err := clientSqlServer.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Test(ctx)
	case database_kind.Sqlite:
		manager, err := clientSqlite.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Test(ctx)
	case database_kind.Redis:
		manager, err := clientRedis.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Test(ctx)
	case database_kind.MongoDB:
		manager, err := clientMongo.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.CloseContext(ctx)
		return manager.Test(ctx)
	case database_kind.Elasticsearch:
		manager, err := clientElasticsearch.NewManager(ctx, config, "")
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Test(ctx)
	default:
		return nil, errors.New("数据库类型不合法")
	}
}

// Mysql 按实例和库名创建独立连接，使用后由调用方关闭。
func (s *service) Mysql(ctx context.Context, id int64, databaseName string) (*clientMySQL.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.MySQL)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	return clientMySQL.NewManager(ctx, connection.Config, databaseName)
}

// Postgres 按实例和库名创建独立连接，使用后由调用方关闭。
func (s *service) Postgres(ctx context.Context, id int64, databaseName string) (*clientPostgres.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.Postgres)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	return clientPostgres.NewManager(ctx, connection.Config, databaseName)
}

// Sqlserver 按实例和库名创建独立连接，使用后由调用方关闭。
func (s *service) Sqlserver(ctx context.Context, id int64, databaseName string) (*clientSqlServer.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.SqlServer)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	return clientSqlServer.NewManager(ctx, connection.Config, databaseName)
}

// Sqlite 按实例和库名创建独立连接，使用后由调用方关闭。
func (s *service) Sqlite(ctx context.Context, id int64, databaseName string) (*clientSqlite.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.Sqlite)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	return clientSqlite.NewManager(ctx, connection.Config, databaseName)
}

// Redis 按实例和库名创建独立连接，使用后由调用方关闭。
func (s *service) Redis(ctx context.Context, id int64, databaseName string) (*clientRedis.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.Redis)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	return clientRedis.NewManager(ctx, connection.Config, databaseName)
}

// Mongo 按实例和库名创建独立连接，使用后由调用方关闭。
func (s *service) Mongo(ctx context.Context, id int64, databaseName string) (*clientMongo.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.MongoDB)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	return clientMongo.NewManager(ctx, connection.Config, databaseName)
}

// Elasticsearch 将工作区数据库名称解释为具体索引，使用后由调用方关闭。
func (s *service) Elasticsearch(ctx context.Context, id int64, databaseName string) (*clientElasticsearch.Manager, error) {
	connection, err := s.FindConnection(id, database_kind.Elasticsearch)
	if err != nil {
		return nil, err
	}
	return clientElasticsearch.NewManager(ctx, connection.Config, databaseName)
}
