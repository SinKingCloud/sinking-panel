package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	serviceContainer "server/app/service/container"
	"server/app/util/context"
	"strings"
	"time"
)

// Images 查询列表、查看详情、导入或删除容器镜像。
func Images(c *context.Context) {
	var form struct {
		Action string `json:"action" default:"list" validate:"required,oneof=list info import remove" label:"操作类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	switch form.Action {
	case "list":
		query := c.ValidatePage("create_time", "desc", "id,name,create_time")
		var form struct {
			Keyword         string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
			Id              string `json:"id" default:"" validate:"omitempty,max=255" label:"镜像ID"`
			Name            string `json:"name" default:"" validate:"omitempty,max=255" label:"镜像名称"`
			Tag             string `json:"tag" default:"" validate:"omitempty,max=255" label:"镜像标签"`
			CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
			CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
		}
		if ok, msg := c.ValidatorAll(&form); !ok {
			c.Error(msg)
			return
		}
		where := &serviceContainer.SelectImage{}
		if form.CreateTimeStart != "" {
			value, err := time.ParseInLocation(time.DateTime, form.CreateTimeStart, time.Local)
			if err != nil {
				c.Error("创建起始时间格式错误")
				return
			}
			timestamp := value.Unix()
			where.CreateTimeStart = &timestamp
		}
		if form.CreateTimeEnd != "" {
			value, err := time.ParseInLocation(time.DateTime, form.CreateTimeEnd, time.Local)
			if err != nil {
				c.Error("创建结束时间格式错误")
				return
			}
			timestamp := value.Unix()
			where.CreateTimeEnd = &timestamp
		}
		if form.Keyword != "" {
			where.Keyword = &form.Keyword
		}
		if form.Id != "" {
			where.ID = &form.Id
		}
		if form.Name != "" {
			where.Name = &form.Name
		}
		if form.Tag != "" {
			where.Tag = &form.Tag
		}
		data, err := service.Container.Images(where, query)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看容器镜像", "查看容器镜像列表")
			c.SuccessWithData("获取成功", data)
		}
	case "info":
		var form struct {
			Id string `json:"id" validate:"required" label:"镜像ID"`
		}
		if ok, msg := c.ValidatorAll(&form); !ok {
			c.Error(msg)
			return
		}
		data, err := service.Container.Image(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			c.SuccessWithData("获取成功", data)
		}
	case "import":
		var form struct {
			ImagePath string `json:"image_path" default:"" validate:"required" label:"镜像路径"`
		}
		if ok, msg := c.ValidatorAll(&form); !ok {
			c.Error(msg)
			return
		}
		data, err := service.Container.Import(form.ImagePath)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "导入容器镜像", "导入容器镜像["+data.Id+"]")
			c.SuccessWithData("导入成功", data)
		}
	case "remove":
		var form struct {
			Ids []string `json:"ids" default:"" validate:"required,min=1,max=1000,unique,dive,required" label:"镜像ID列表"`
		}
		if ok, msg := c.ValidatorAll(&form); !ok {
			c.Error(msg)
			return
		}
		if err := service.Container.RemoveImage(form.Ids); err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除容器镜像", "删除容器镜像["+strings.Join(form.Ids, ",")+"]")
			c.Success("删除成功")
		}
	}
}
