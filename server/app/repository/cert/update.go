package cert

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateById 更新证书。
func (r *Repository) UpdateById(id int64, data *UpdateCert, tx ...*gorm.DB) error {
	if data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.Type != nil {
		updates["type"] = *data.Type
	}
	if data.Challenge != nil {
		updates["challenge"] = *data.Challenge
	}
	if data.SecretId != nil {
		updates["secret_id"] = *data.SecretId
	}
	if data.AutoRenew != nil {
		updates["auto_renew"] = *data.AutoRenew
	}
	if data.Domains != nil {
		updates["domains"] = *data.Domains
	}
	if data.Certificate != nil {
		updates["certificate"] = *data.Certificate
	}
	if data.PrivateKey != nil {
		updates["private_key"] = *data.PrivateKey
	}
	if data.StartTime != nil {
		updates["start_time"] = *data.StartTime
	}
	if data.ExpireTime != nil {
		updates["expire_time"] = *data.ExpireTime
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	execute := func(db *gorm.DB) error {
		result := db.Model(&model.Cert{}).
			Where("id = ?", id).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return execute(r.Database.Db)
}

// ClearSecretId 解除指定密钥的证书关联，并关闭自动续签。
func (r *Repository) ClearSecretId(secretId int64, tx ...*gorm.DB) error {
	if secretId <= 0 {
		return nil
	}
	updates := make(map[string]interface{})
	updates["secret_id"] = int64(0)
	updates["auto_renew"] = 0
	updates["update_time"] = str.DateTime(time.Now())
	execute := func(db *gorm.DB) error {
		return db.Model(&model.Cert{}).
			Where("secret_id = ?", secretId).
			Updates(updates).Error
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return execute(r.Database.Db)
}

// Restore 恢复证书快照，保留原始时间，不触发更新钩子。
func (r *Repository) Restore(data *model.Cert, tx ...*gorm.DB) error {
	execute := func(db *gorm.DB) error {
		return db.Model(&model.Cert{}).
			Where("id = ?", data.Id).
			UpdateColumns(map[string]interface{}{
				"name":        data.Name,
				"type":        data.Type,
				"challenge":   data.Challenge,
				"secret_id":   data.SecretId,
				"auto_renew":  data.AutoRenew,
				"domains":     data.Domains,
				"certificate": data.Certificate,
				"private_key": data.PrivateKey,
				"start_time":  data.StartTime,
				"expire_time": data.ExpireTime,
				"create_time": data.CreateTime,
				"update_time": data.UpdateTime,
			}).Error
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return execute(r.Database.Db)
}
