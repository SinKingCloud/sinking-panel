package database

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

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

// SyncSqlite 同步sqlite数据库结构，schemaSql必须包含完整表结构
func (d *Database) SyncSqlite(schemaSql string) error {
	if d == nil || d.Db == nil {
		return gorm.ErrInvalidDB
	}
	if strings.TrimSpace(schemaSql) == "" {
		return errors.New("数据库结构SQL不能为空")
	}
	type schemaObject struct {
		Type string `gorm:"column:type"`
		Name string `gorm:"column:name"`
		SQL  string `gorm:"column:sql"`
	}
	normalizeSql := func(sql string) string {
		var builder strings.Builder
		chars := []rune(sql)
		for i := 0; i < len(chars); {
			char := chars[i]
			if unicode.IsSpace(char) {
				i++
				continue
			}
			switch char {
			case '\'':
				builder.WriteString("\x00s")
				builder.WriteRune(char)
				i++
				for i < len(chars) {
					builder.WriteRune(chars[i])
					if chars[i] != '\'' {
						i++
						continue
					}
					if i+1 < len(chars) && chars[i+1] == '\'' {
						builder.WriteRune(chars[i+1])
						i += 2
						continue
					}
					i++
					break
				}
			case '"', '`':
				builder.WriteString("\x00w")
				quote := char
				i++
				for i < len(chars) {
					if chars[i] != quote {
						builder.WriteRune(unicode.ToLower(chars[i]))
						i++
						continue
					}
					if i+1 < len(chars) && chars[i+1] == quote {
						builder.WriteRune(quote)
						i += 2
						continue
					}
					i++
					break
				}
			case '[':
				builder.WriteString("\x00w")
				i++
				for i < len(chars) && chars[i] != ']' {
					builder.WriteRune(unicode.ToLower(chars[i]))
					i++
				}
				if i < len(chars) {
					i++
				}
			default:
				if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' || char == '$' {
					builder.WriteString("\x00w")
					for i < len(chars) {
						char = chars[i]
						if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '$' {
							break
						}
						builder.WriteRune(unicode.ToLower(char))
						i++
					}
					continue
				}
				builder.WriteString("\x00p")
				builder.WriteRune(char)
				i++
			}
		}
		return builder.String()
	}

	expectedDb := NewSqlite(":memory:")
	if expectedDb.DbError != nil {
		return expectedDb.DbError
	}
	sqlDb, err := expectedDb.Db.DB()
	if err != nil {
		return err
	}
	sqlDb.SetMaxOpenConns(1)
	defer sqlDb.Close()
	if err = expectedDb.Transaction(func(tx *gorm.DB) error {
		if executeErr := tx.Exec(schemaSql).Error; executeErr != nil {
			return fmt.Errorf("执行数据库结构SQL失败: %w", executeErr)
		}
		return nil
	}); err != nil {
		return err
	}
	var expected []*schemaObject
	if err = expectedDb.Db.Raw(`SELECT type, name, sql FROM sqlite_master
		WHERE type IN ('table', 'index') AND name NOT LIKE 'sqlite_%' AND sql IS NOT NULL
		ORDER BY rowid`).Scan(&expected).Error; err != nil {
		return err
	}
	expectedTables := make(map[string]*schemaObject)
	expectedIndexes := make(map[string]*schemaObject)
	for _, object := range expected {
		if object.Type == "table" {
			expectedTables[strings.ToLower(object.Name)] = object
		} else if object.Type == "index" {
			expectedIndexes[strings.ToLower(object.Name)] = object
		}
	}
	if len(expectedTables) == 0 {
		return errors.New("数据库结构SQL未定义数据表")
	}
	return d.Db.Connection(func(db *gorm.DB) (err error) {
		var foreignKeys int
		if err = db.Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error; err != nil {
			return err
		}
		if foreignKeys > 0 {
			if err = db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
				return err
			}
			defer func() {
				if enableErr := db.Exec("PRAGMA foreign_keys = ON").Error; err == nil && enableErr != nil {
					err = enableErr
				}
			}()
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			var actual []*schemaObject
			if queryErr := tx.Raw(`SELECT type, name, sql FROM sqlite_master
				WHERE type IN ('table', 'index') AND name NOT LIKE 'sqlite_%' AND sql IS NOT NULL
				ORDER BY rowid`).Scan(&actual).Error; queryErr != nil {
				return queryErr
			}
			actualTables := make(map[string]*schemaObject)
			for _, object := range actual {
				if object.Type == "table" {
					actualTables[strings.ToLower(object.Name)] = object
				}
			}
			for _, object := range actual {
				if object.Type != "table" || expectedTables[strings.ToLower(object.Name)] != nil {
					continue
				}
				if dropErr := tx.Exec("DROP TABLE " + tx.Statement.Quote(object.Name)).Error; dropErr != nil {
					return fmt.Errorf("删除多余数据表%s失败: %w", object.Name, dropErr)
				}
				delete(actualTables, strings.ToLower(object.Name))
			}
			for _, object := range expected {
				if object.Type != "table" {
					continue
				}
				current := actualTables[strings.ToLower(object.Name)]
				if current == nil {
					if createErr := tx.Exec(object.SQL).Error; createErr != nil {
						return fmt.Errorf("创建数据表%s失败: %w", object.Name, createErr)
					}
					continue
				}
				if normalizeSql(current.SQL) == normalizeSql(object.SQL) {
					continue
				}
				newName := "__schema_new_" + object.Name
				prefix := "CREATE TABLE " + object.Name
				if !strings.HasPrefix(object.SQL, prefix) {
					return fmt.Errorf("无法识别数据表%s的建表语句", object.Name)
				}
				createSql := "CREATE TABLE " + tx.Statement.Quote(newName) + object.SQL[len(prefix):]
				if createErr := tx.Exec(createSql).Error; createErr != nil {
					return fmt.Errorf("同步数据表%s失败: %w", object.Name, createErr)
				}
				var oldColumns []string
				if queryErr := tx.Raw("SELECT name FROM pragma_table_info(?) ORDER BY cid", current.Name).Scan(&oldColumns).Error; queryErr != nil {
					return fmt.Errorf("同步数据表%s失败: %w", object.Name, queryErr)
				}
				var newColumns []string
				if queryErr := tx.Raw("SELECT name FROM pragma_table_info(?) ORDER BY cid", newName).Scan(&newColumns).Error; queryErr != nil {
					return fmt.Errorf("同步数据表%s失败: %w", object.Name, queryErr)
				}
				oldColumnMap := make(map[string]string, len(oldColumns))
				for _, column := range oldColumns {
					oldColumnMap[strings.ToLower(column)] = column
				}
				targetColumns := make([]string, 0, len(newColumns))
				sourceColumns := make([]string, 0, len(newColumns))
				for _, column := range newColumns {
					if sourceColumn, ok := oldColumnMap[strings.ToLower(column)]; ok {
						targetColumns = append(targetColumns, tx.Statement.Quote(column))
						sourceColumns = append(sourceColumns, tx.Statement.Quote(sourceColumn))
					}
				}
				if len(targetColumns) > 0 {
					copySql := "INSERT INTO " + tx.Statement.Quote(newName) + " (" + strings.Join(targetColumns, ",") + ") SELECT " + strings.Join(sourceColumns, ",") + " FROM " + tx.Statement.Quote(current.Name)
					if copyErr := tx.Exec(copySql).Error; copyErr != nil {
						return fmt.Errorf("同步数据表%s失败: %w", object.Name, copyErr)
					}
				} else {
					var total int64
					if queryErr := tx.Raw("SELECT COUNT(*) FROM " + tx.Statement.Quote(current.Name)).Scan(&total).Error; queryErr != nil {
						return fmt.Errorf("同步数据表%s失败: %w", object.Name, queryErr)
					}
					if total > 0 {
						return fmt.Errorf("同步数据表%s失败: 没有可迁移的同名字段", object.Name)
					}
				}
				if dropErr := tx.Exec("DROP TABLE " + tx.Statement.Quote(current.Name)).Error; dropErr != nil {
					return fmt.Errorf("同步数据表%s失败: %w", object.Name, dropErr)
				}
				if renameErr := tx.Exec("ALTER TABLE " + tx.Statement.Quote(newName) + " RENAME TO " + tx.Statement.Quote(object.Name)).Error; renameErr != nil {
					return fmt.Errorf("同步数据表%s失败: %w", object.Name, renameErr)
				}
			}
			actual = nil
			if queryErr := tx.Raw(`SELECT type, name, sql FROM sqlite_master
				WHERE type IN ('table', 'index') AND name NOT LIKE 'sqlite_%' AND sql IS NOT NULL
				ORDER BY rowid`).Scan(&actual).Error; queryErr != nil {
				return queryErr
			}
			actualIndexes := make(map[string]*schemaObject)
			for _, object := range actual {
				if object.Type == "index" {
					actualIndexes[strings.ToLower(object.Name)] = object
				}
			}
			for name, object := range actualIndexes {
				expectedIndex := expectedIndexes[name]
				if expectedIndex != nil && normalizeSql(object.SQL) == normalizeSql(expectedIndex.SQL) {
					continue
				}
				if dropErr := tx.Exec("DROP INDEX " + tx.Statement.Quote(object.Name)).Error; dropErr != nil {
					return fmt.Errorf("删除多余索引%s失败: %w", object.Name, dropErr)
				}
				delete(actualIndexes, name)
			}
			for _, object := range expected {
				if object.Type != "index" || actualIndexes[strings.ToLower(object.Name)] != nil {
					continue
				}
				if createErr := tx.Exec(object.SQL).Error; createErr != nil {
					return fmt.Errorf("创建索引%s失败: %w", object.Name, createErr)
				}
			}
			var violations int64
			if checkErr := tx.Raw("SELECT COUNT(*) FROM pragma_foreign_key_check").Scan(&violations).Error; checkErr != nil {
				return checkErr
			}
			if violations > 0 {
				return fmt.Errorf("数据库存在%d条外键异常", violations)
			}
			return nil
		})
		return err
	})
}
