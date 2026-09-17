package cert

import (
	"server/app/enum/cert_auto_renew"
	"server/app/enum/cert_type"
	"server/app/model"
	"server/app/util/page"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
	"gorm.io/gorm"
)

// SelectByIds 查询指定证书。
func (r *Repository) SelectByIds(ids []int64, tx ...*gorm.DB) ([]*model.Cert, error) {
	list := make([]*model.Cert, 0)
	if len(ids) == 0 {
		return list, nil
	}
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("id IN ?", ids).Order("id ASC").Find(&list).Error
	return list, err
}

// SelectAutoRenewTask 查询指定时间前到期的自动续签证书，包含已过期证书。
func (r *Repository) SelectAutoRenewTask(expireTime time.Time) ([]*Cert, error) {
	list := make([]*Cert, 0)
	err := r.Database.Db.Model(&model.Cert{}).
		Where("type = ? AND auto_renew = ? AND expire_time <= ?", cert_type.Auto, cert_auto_renew.Enabled, expireTime.Format("2006-01-02 15:04:05")).
		Order("expire_time ASC, id ASC").Find(&list).Error
	return list, err
}

// Select 分页查询证书。
func (r *Repository) Select(where *SelectCert, queryPage *page.Query) (*page.Result[*Cert], error) {
	query := r.Database.Db.Model(&model.Cert{})
	if where != nil {
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("name LIKE ? OR domains LIKE ?", keyword, keyword)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Domains != nil {
			query = query.Where("domains LIKE ?", "%"+*where.Domains+"%")
		}
		if where.Type != nil {
			query = query.Where("type = ?", *where.Type)
		}
		if where.Domain != nil {
			domain := *where.Domain
			// 连同 JSON 双引号匹配完整域名，避免命中其他域名的子串。
			pattern := "%" + strconv.Quote(domain) + "%"
			rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
			if err == nil && domain != rootDomain && !strings.HasPrefix(domain, "*.") {
				// 仅注册域名下的子域名匹配泛域名，且只替换最左侧一级。
				_, parentDomain, _ := strings.Cut(domain, ".")
				query = query.Where("domains LIKE ? OR domains LIKE ?", pattern, "%"+strconv.Quote("*."+parentDomain)+"%")
			} else {
				query = query.Where("domains LIKE ?", pattern)
			}
		}
		if where.Usable != nil && *where.Usable == 1 {
			// 与证书有效期写入时使用的服务器本地时间保持一致。
			now := time.Now().Format("2006-01-02 15:04:05")
			query = query.Where("start_time <= ? AND expire_time > ?", now, now)
		}
		if where.ExpireTimeStart != nil {
			query = query.Where("expire_time >= ?", *where.ExpireTimeStart)
		}
		if where.ExpireTimeEnd != nil {
			query = query.Where("expire_time <= ?", *where.ExpireTimeEnd)
		}
		if where.CreateTimeStart != nil {
			query = query.Where("create_time >= ?", *where.CreateTimeStart)
		}
		if where.CreateTimeEnd != nil {
			query = query.Where("create_time <= ?", *where.CreateTimeEnd)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
