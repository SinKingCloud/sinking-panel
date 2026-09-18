package str

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DateTime 返回日期格式
type DateTime time.Time

func (t DateTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}
func (t *DateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var err error
	str := string(data)
	timeStr := strings.Trim(str, "\"")
	t1, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	*t = DateTime(t1)
	return err
}

func (t DateTime) Value() (driver.Value, error) {
	tTime := time.Time(t)
	return tTime.Format("2006-01-02 15:04:05"), nil
}

func (t *DateTime) Scan(v interface{}) error {
	var value string
	switch vt := v.(type) {
	case time.Time:
		// 数据库存的是无时区的本地时间，保留驱动返回的年月日时分秒。
		value = vt.Format("2006-01-02 15:04:05")
	case string:
		value = vt
	case []byte:
		value = string(vt)
	case nil:
		*t = DateTime{}
		return nil
	default:
		return errors.New("类型处理错误")
	}
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		return err
	}
	*t = DateTime(parsedTime)
	return nil
}

func (t *DateTime) String() string {
	return time.Time(*t).Format("2006-01-02 15:04:05")
}
