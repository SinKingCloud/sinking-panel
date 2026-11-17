package settings

import (
	"fmt"
	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
)

// Update 修改网站自动备份设置。
func Update(c *context.Context) {
	var form struct {
		Enabled   *bool  `json:"enabled" label:"自动备份"`
		Spec      string `json:"spec" validate:"omitempty,max=200" label:"执行周期"`
		Retries   *int   `json:"retries" validate:"omitempty,min=0,max=10" label:"重试次数"`
		LocalKeep *int   `json:"local_keep" validate:"omitempty,min=0,max=1000" label:"本地保留副本数"`
		UploadOss *bool  `json:"upload_oss" label:"上传云端"`
		OssId     int64  `json:"oss_id" default:"0" validate:"min=0" label:"存储实例"`
		OssKeep   int    `json:"oss_keep" default:"1" validate:"min=1,max=1000" label:"云端保留副本数"`
		Prefix    string `json:"prefix" default:"" validate:"omitempty,max=512" label:"云端备份目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Enabled == nil || form.Spec == "" || form.Retries == nil || form.LocalKeep == nil || form.UploadOss == nil {
		c.Error("请填写自动备份开关、执行周期、重试次数、本地保留副本数和上传云端选项")
		return
	}
	configs := map[string]string{
		constant.SiteBackupEnabled:   strconv.FormatBool(*form.Enabled),
		constant.SiteBackupSpec:      strings.TrimSpace(form.Spec),
		constant.SiteBackupRetries:   strconv.Itoa(*form.Retries),
		constant.SiteBackupLocalKeep: strconv.Itoa(*form.LocalKeep),
		constant.SiteBackupUploadOss: strconv.FormatBool(*form.UploadOss),
		constant.SiteBackupOssId:     strconv.FormatInt(form.OssId, 10),
		constant.SiteBackupOssKeep:   strconv.Itoa(form.OssKeep),
		constant.SiteBackupPrefix:    strings.TrimSpace(form.Prefix),
	}
	if err := service.Site.SetBackupSettings(configs); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站备份设置", fmt.Sprintf("自动备份：%t，执行周期：%s，重试次数：%d，本地副本：%d，上传云端：%t", *form.Enabled, configs[constant.SiteBackupSpec], *form.Retries, *form.LocalKeep, *form.UploadOss))
	c.Success("保存成功")
}
