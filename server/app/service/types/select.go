package types

import (
	"errors"
	"server/app/enum/type_module"
	"server/app/model"
	repositoryTypes "server/app/repository/types"
	"server/app/util/page"
)

// FindById 查询指定ID的类型
func (s *service) FindById(id int64) (*model.Type, error) {
	data, err := s.repositoryTypes.SelectByIds([]int64{id})
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("类型不存在")
	}
	return data[0], nil
}

// Select 获取数据
func (s *service) Select(where *repositoryTypes.SelectType, queryPage *page.Query) (*page.Result[*model.Type], error) {
	if where != nil && where.Module != "" {
		if _, ok := type_module.Map()[where.Module]; !ok {
			return nil, errors.New("所属模块不合法")
		}
	}
	return s.repositoryTypes.Select(where, queryPage)
}
