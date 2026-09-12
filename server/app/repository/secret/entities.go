package secret

import "server/app/util/str"

// SelectSecret 密钥凭据查询条件。
type SelectSecret struct {
	Keyword  string
	Provider string
}

// UpdateSecret 密钥凭据更新内容。
type UpdateSecret struct {
	Name     *string
	Provider *int
	Data     *string
}

// Secret 密钥凭据列表数据，不返回凭据正文。
type Secret struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Name       string       `gorm:"column:name" json:"name"`
	Provider   int          `gorm:"column:provider" json:"provider"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
