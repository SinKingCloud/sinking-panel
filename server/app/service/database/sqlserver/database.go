package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// DatabaseInfo 读取真实数据库的排序规则、状态、兼容级别和文件占用空间。
func (m *Manager) DatabaseInfo(ctx context.Context, name string) (_ map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err = m.quote(name); err != nil || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 128 {
		return nil, errors.New("数据库名称不合法，最多允许 128 个字符")
	}
	var databaseName, status, recovery string
	var collation sql.NullString
	var compatibility int
	var readOnly bool
	var size int64
	query := "SELECT d.name, d.collation_name, d.state_desc, d.recovery_model_desc, d.compatibility_level, d.is_read_only, COALESCE((SELECT SUM(CAST(f.size AS bigint)) * 8192 FROM sys.master_files f WHERE f.database_id = d.database_id), 0) FROM sys.databases d WHERE d.name = @p1"
	err = m.db.QueryRowContext(ctx, query, name).Scan(&databaseName, &collation, &status, &recovery, &compatibility, &readOnly, &size)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("数据库不存在或当前账户无权访问")
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": databaseName, "collation": collation.String, "status": status, "recovery_model": recovery, "compatibility_level": compatibility, "read_only": readOnly, "size": size}, nil
}

// CreateDatabase 创建真实数据库，使用服务器的默认数据库配置。
func (m *Manager) CreateDatabase(ctx context.Context, name string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	name = strings.TrimSpace(name)
	databaseName, err := m.quote(name)
	if err != nil || utf8.RuneCountInString(name) > 128 {
		return errors.New("数据库名称不合法，最多允许 128 个字符")
	}
	switch strings.ToLower(strings.TrimRight(name, " ")) {
	case "master", "model", "msdb", "tempdb":
		return errors.New("不能创建系统数据库")
	}
	_, err = m.db.ExecContext(ctx, "CREATE DATABASE "+databaseName)
	return err
}

// RenameDatabase 重命名真实业务数据库，不强制断开其他连接。
func (m *Manager) RenameDatabase(ctx context.Context, name, newName string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	newName = strings.TrimSpace(newName)
	databaseName, err := m.quote(name)
	if err != nil || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 128 {
		return errors.New("数据库名称不合法，最多允许 128 个字符")
	}
	newDatabaseName, err := m.quote(newName)
	if err != nil || utf8.RuneCountInString(newName) > 128 {
		return errors.New("新数据库名称不合法，最多允许 128 个字符")
	}
	for _, value := range []string{name, newName} {
		switch strings.ToLower(strings.TrimRight(value, " ")) {
		case "master", "model", "msdb", "tempdb":
			return errors.New("不能重命名系统数据库或使用系统数据库名称")
		}
	}
	if strings.EqualFold(name, m.database) {
		return errors.New("重命名数据库必须连接到其他数据库")
	}
	// 实例排序规则可能忽略名称大小写或重音，按真实 ID 保护系统库。
	var databaseID sql.NullInt64
	if err = m.db.QueryRowContext(ctx, "SELECT DB_ID(@p1)", name).Scan(&databaseID); err != nil {
		return err
	}
	if databaseID.Valid && databaseID.Int64 <= 4 {
		return errors.New("不能重命名系统数据库")
	}
	if name == newName {
		_, err = m.DatabaseInfo(ctx, name)
		return err
	}
	_, err = m.db.ExecContext(ctx, "ALTER DATABASE "+databaseName+" MODIFY NAME = "+newDatabaseName)
	return err
}

// DropDatabase 删除真实业务数据库；兼容不支持 DROP DATABASE IF EXISTS 的旧版本。
func (m *Manager) DropDatabase(ctx context.Context, name string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	databaseName, err := m.dropDatabaseTarget(ctx, name)
	if err != nil || databaseName == "" {
		return err
	}
	_, err = m.db.ExecContext(ctx, "DROP DATABASE "+databaseName)
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
		if _, err = m.dropDatabaseTarget(ctx, name); err != nil {
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

// dropDatabaseTarget 同时检查实例排序规则解析后的真实 ID；不存在时返回空目标。
func (m *Manager) dropDatabaseTarget(ctx context.Context, name string) (string, error) {
	databaseName, err := m.quote(name)
	if err != nil || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 128 {
		return "", errors.New("数据库名称不合法，最多允许 128 个字符")
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "master", "model", "msdb", "tempdb":
		return "", errors.New("不能删除系统数据库")
	}
	if strings.EqualFold(name, m.database) {
		return "", errors.New("删除数据库必须连接到其他数据库")
	}
	var databaseID sql.NullInt64
	if err = m.db.QueryRowContext(ctx, "SELECT DB_ID(@p1)", name).Scan(&databaseID); err != nil {
		return "", err
	}
	if databaseID.Valid && databaseID.Int64 <= 4 {
		return "", errors.New("不能删除系统数据库")
	}
	if !databaseID.Valid {
		// DB_ID 也会因权限不足返回 NULL，只有能查看全部库时才按缺失处理。
		var canInspect, existing int
		query := "SELECT CASE WHEN HAS_PERMS_BY_NAME(NULL, NULL, 'ALTER ANY DATABASE') = 1 OR HAS_PERMS_BY_NAME('master', 'DATABASE', 'CREATE DATABASE') = 1 THEN 1 ELSE 0 END, (SELECT COUNT(*) FROM sys.databases WHERE name = @p1)"
		if err = m.db.QueryRowContext(ctx, query, name).Scan(&canInspect, &existing); err != nil {
			return "", err
		}
		if canInspect != 1 || existing != 0 {
			return "", errors.New("数据库不存在或当前账户无权访问，无法确认删除结果")
		}
		return "", nil
	}
	return databaseName, nil
}
