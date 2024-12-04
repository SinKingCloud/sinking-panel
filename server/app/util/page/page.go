package page

type Page struct {
	Total    int64       `json:"total"`     //记录总数
	Page     int         `json:"page"`      //分页页码
	PageSize int         `json:"page_size"` //分页容量
	List     interface{} `json:"list"`      //记录详情
}

// NewPage 实例化新page
func NewPage(total int64, page int, pageSize int, list interface{}) *Page {
	return &Page{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     list,
	}
}
