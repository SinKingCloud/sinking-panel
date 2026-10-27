package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DatabaseInfo 读取真实数据库的所有者、编码、排序规则和占用空间。
func (m *Manager) DatabaseInfo(ctx context.Context, name string) (_ map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err = m.quote(name); err != nil || strings.TrimSpace(name) == "" || len(name) > 63 {
		return nil, errors.New("数据库名称不合法，最多允许 63 字节")
	}
	var databaseName, owner, encoding, collation string
	var size int64
	var allowConnections bool
	query := "SELECT datname, pg_get_userbyid(datdba), pg_encoding_to_char(encoding), datcollate, datallowconn, pg_database_size(oid) FROM pg_database WHERE datname = $1 AND has_database_privilege(datname, 'CONNECT')"
	err = m.db.QueryRowContext(ctx, query, name).Scan(&databaseName, &owner, &encoding, &collation, &allowConnections, &size)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("数据库不存在或当前账户无权访问")
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"name":              databaseName,
		"owner":             owner,
		"encoding":          encoding,
		"collation":         collation,
		"allow_connections": allowConnections,
		"size":              size,
	}, nil
}

// CreateDatabase 创建真实数据库，继承服务器模板库的默认配置。
func (m *Manager) CreateDatabase(ctx context.Context, name string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	name = strings.TrimSpace(name)
	databaseName, err := m.quote(name)
	if err != nil || len(name) > 63 {
		return errors.New("数据库名称不合法，最多允许 63 字节")
	}
	switch strings.ToLower(name) {
	case "postgres", "template0", "template1":
		return errors.New("不能创建系统数据库")
	}
	_, err = m.db.ExecContext(ctx, "CREATE DATABASE "+databaseName)
	return err
}

// RenameDatabase 从其他数据库连接重命名业务数据库，不终止已有连接。
func (m *Manager) RenameDatabase(ctx context.Context, name, newName string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	newName = strings.TrimSpace(newName)
	databaseName, err := m.quote(name)
	if err != nil || strings.TrimSpace(name) == "" || len(name) > 63 {
		return errors.New("数据库名称不合法，最多允许 63 字节")
	}
	newDatabaseName, err := m.quote(newName)
	if err != nil || len(newName) > 63 {
		return errors.New("新数据库名称不合法，最多允许 63 字节")
	}
	for _, value := range []string{name, newName} {
		switch strings.ToLower(value) {
		case "postgres", "template0", "template1":
			return errors.New("不能重命名系统数据库或使用系统数据库名称")
		}
	}
	if name == m.database {
		return errors.New("重命名数据库必须连接到其他数据库")
	}
	if name == newName {
		_, err = m.DatabaseInfo(ctx, name)
		return err
	}
	_, err = m.db.ExecContext(ctx, "ALTER DATABASE "+databaseName+" RENAME TO "+newDatabaseName)
	return err
}

// DropDatabase 删除真实业务数据库，不终止已有连接；不存在时允许重试。
func (m *Manager) DropDatabase(ctx context.Context, name string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	databaseName, err := m.dropDatabaseTarget(name)
	if err != nil {
		return err
	}
	_, err = m.db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+databaseName)
	return err
}

// DropDatabases 在发送任何删除前检查整批目标，并报告失败前已完成的数量。
func (m *Manager) DropDatabases(ctx context.Context, names []string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if len(names) == 0 {
		return errors.New("请选择数据库")
	}
	for _, name := range names {
		if _, err = m.dropDatabaseTarget(name); err != nil {
			return err
		}
	}
	for completed, name := range names {
		if err = m.DropDatabase(ctx, name); err != nil {
			return fmt.Errorf("已删除 %d 个数据库，删除 %s 失败：%w", completed, name, err)
		}
	}
	return nil
}

func (m *Manager) dropDatabaseTarget(name string) (string, error) {
	databaseName, err := m.quote(name)
	if err != nil || strings.TrimSpace(name) == "" || len(name) > 63 {
		return "", errors.New("数据库名称不合法，最多允许 63 字节")
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "postgres", "template0", "template1":
		return "", errors.New("不能删除系统数据库")
	}
	if name == m.database {
		return "", errors.New("删除数据库必须连接到其他数据库")
	}
	return databaseName, nil
}
