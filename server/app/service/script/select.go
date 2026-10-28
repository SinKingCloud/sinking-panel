package script

import (
	"errors"
	"server/app/enum/script_module"
	repositoryScript "server/app/repository/script"
	"server/app/util/page"
	"strconv"
	"strings"
)

// Select 查询常用脚本。
func (s *service) Select(where *repositoryScript.SelectScript, queryPage *page.Query) (*page.Result[*repositoryScript.Script], error) {
	if where != nil && where.Module != nil {
		module := *where.Module
		if index := strings.LastIndexByte(module, '_'); index > 0 {
			if _, err := strconv.ParseUint(module[index+1:], 10, 64); err == nil {
				module = module[:index]
			}
		}
		if _, ok := script_module.Map()[module]; !ok {
			return nil, errors.New("脚本所属模块不合法")
		}
	}
	return s.repositoryScript.Select(where, queryPage)
}
