package redis

import (
	stdContext "context"
	"encoding/json"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
	"time"
)

// Command 执行命令。
func (Controller) Command(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Command      string `json:"command" validate:"omitempty,max=1048576" label:"命令"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var values []interface{}
	decoder := json.NewDecoder(strings.NewReader(form.Command))
	decoder.UseNumber()
	if !json.Valid([]byte(form.Command)) || decoder.Decode(&values) != nil || len(values) == 0 || len(values) > 10000 {
		c.Error("命令必须是非空 JSON 数组，例如 [\"GET\",\"key\"]")
		return
	}
	arguments := make([]string, len(values))
	for i, value := range values {
		switch value := value.(type) {
		case string:
			arguments[i] = value
		case json.Number:
			arguments[i] = value.String()
		default:
			c.Error("命令参数必须是字符串或数字")
			return
		}
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Redis(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.Command(ctx, arguments)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Redis 执行命令", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
