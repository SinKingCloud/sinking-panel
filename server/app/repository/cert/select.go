package cert

import (
	"net"
	"server/app/model"
	"server/app/util/page"
	"strconv"
	"strings"
	"time"

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
		if where.Type != nil {
			query = query.Where("type = ?", *where.Type)
		}
		if where.Domain != nil {
			domain := *where.Domain
			// 连同 JSON 双引号匹配完整域名，避免命中其他域名的子串。
			if _, suffix, ok := strings.Cut(domain, "."); ok && !strings.HasPrefix(domain, "*.") && net.ParseIP(domain) == nil {
				query = query.Where("domains LIKE ? OR domains LIKE ?", "%"+strconv.Quote(domain)+"%", "%"+strconv.Quote("*."+suffix)+"%")
			} else {
				query = query.Where("domains LIKE ?", "%"+strconv.Quote(domain)+"%")
			}
		}
		if where.Usable != nil && *where.Usable == 1 {
			// X.509 有效期以 UTC 写入数据库。
			now := time.Now().UTC().Format("2006-01-02 15:04:05")
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
