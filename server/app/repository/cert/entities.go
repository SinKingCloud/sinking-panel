package cert

import "server/app/util/str"

// SelectCert 证书查询条件。
type SelectCert struct {
	Keyword         string
	Name            string
	Type            string
	ExpireTimeStart string
	ExpireTimeEnd   string
	CreateTimeStart string
	CreateTimeEnd   string
}

// UpdateCert 证书更新内容。
type UpdateCert struct {
	Name        *string
	Type        *int
	Challenge   *string
	SecretId    *int64
	AutoRenew   *int
	Domains     *string
	Certificate *string
	PrivateKey  *string
	StartTime   *str.DateTime
	ExpireTime  *str.DateTime
}

// Cert 证书列表数据，不在列表中返回证书和私钥正文。
type Cert struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Name       string       `gorm:"column:name" json:"name"`
	Type       int          `gorm:"column:type" json:"type"`
	Challenge  string       `gorm:"column:challenge" json:"challenge"`
	SecretId   int64        `gorm:"column:secret_id" json:"secret_id"`
	AutoRenew  int          `gorm:"column:auto_renew" json:"auto_renew"`
	Domains    string       `gorm:"column:domains" json:"domains"`
	StartTime  str.DateTime `gorm:"column:start_time" json:"start_time"`
	ExpireTime str.DateTime `gorm:"column:expire_time" json:"expire_time"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
