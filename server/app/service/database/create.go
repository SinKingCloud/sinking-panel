package database

import (
	"encoding/json"
	"errors"
	"server/app/enum/database_kind"
	"server/app/enum/type_module"
	"server/app/model"
	repositoryDatabase "server/app/repository/database"
	"strconv"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

// Create 保存实例连接信息，配置保持 JSON 对象以兼容不同数据库的连接参数。
func (s *service) Create(data *model.Database) error {
	if data == nil {
		return errors.New("实例信息不能为空")
	}
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" || utf8.RuneCountInString(data.Name) > 100 {
		return errors.New("实例名称应为 1 至 100 个字符")
	}
	if _, ok := database_kind.Map()[data.Type]; !ok {
		return errors.New("数据库类型不合法")
	}
	if data.TypeId < 0 {
		return errors.New("分类不合法")
	}
	if data.TypeId > 0 {
		category, err := s.typeService.FindById(data.TypeId)
		if err != nil {
			return err
		}
		if category == nil || category.Module != type_module.Database {
			return errors.New("数据库分类不存在")
		}
	}
	// 创建时允许只登记实例；连接配置由随后打开的编辑表单补充。
	if strings.TrimSpace(data.Config) == "" {
		data.Config = ""
		return s.repository.Create(data)
	}
	config, err := s.Config(data.Type, data.Config, data.Id, "", "decrypt")
	if err != nil {
		return err
	}
	content, _ := json.Marshal(config)
	candidate := *data
	candidate.Config = ""
	// 先取得数据库生成的 ID，再用该 ID 加密；整个过程不写入明文配置。
	if err = s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repository.Create(&candidate, tx); err != nil {
			return err
		}
		candidate.Config = s.encrypt(strconv.FormatInt(candidate.Id, 10), string(content))
		if candidate.Config == "" {
			return errors.New("连接配置加密失败")
		}
		return s.repository.UpdateById(candidate.Id, &repositoryDatabase.UpdateDatabase{
			Config: &candidate.Config,
		}, tx)
	}); err != nil {
		return err
	}
	*data = candidate
	return nil
}
