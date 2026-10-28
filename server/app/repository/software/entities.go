package software

import "server/app/util/str"

// SelectSoftware 软件列表查询条件。
type SelectSoftware struct {
	Keyword *string
	Name    *string
	Code    *string
	Status  *int
}

// UpdateSoftware 软件更新内容，唯一标识创建后不修改。
type UpdateSoftware struct {
	Name    *string
	Version *string
	Status  *int
	Config  *string
}

// Software 列表摘要，不读取或返回插件内部配置。
type Software struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Name       string       `gorm:"column:name" json:"name"`
	Code       string       `gorm:"column:code" json:"code"`
	Version    string       `gorm:"column:version" json:"version"`
	Icon       string       `gorm:"-" json:"icon"` // 由服务从包清单缓存填充，不写入数据库
	Status     int          `gorm:"column:status" json:"status"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
