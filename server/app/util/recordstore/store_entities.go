package recordstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

var ErrNotFound = errors.New("记录不存在或已过期")

const (
	EncodingJSON  = "json"
	EncodingBytes = "bytes"
)

// Options 控制通用存储，默认单条数据最多 16 MiB；不限制业务字段或数据结构。
type Options struct {
	MaxValueBytes int
}

// WriteOptions 的 Time 为业务时间（Unix 毫秒），省略时使用当前时间。
// TTL 从本次写入时刻计算，零值表示永久保存；覆盖写入会替换时间与过期设置。
type WriteOptions struct {
	Time int64
	TTL  time.Duration
}

// Record 的 Data 是编码后的独立副本。JSON 使用 Decode，原始字节直接读取 Data。
type Record struct {
	Key       string `json:"key"`
	Time      int64  `json:"time"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
	Encoding  string `json:"encoding"`
	Data      []byte `json:"data"`
}

func (r Record) Decode(target interface{}) error {
	if r.Encoding != EncodingJSON {
		return errors.New("原始字节记录不能作为 JSON 解码")
	}
	decoder := json.NewDecoder(bytes.NewReader(r.Data))
	decoder.UseNumber()
	return decoder.Decode(target)
}

// Entry 用于原子批量写入 JSON 值；Key 为空时自动生成，已有 Key 则覆盖。
type Entry struct {
	Key     string
	Value   interface{}
	Options WriteOptions
}

// Query 按业务时间和 Key 排序，Start/End 为包含边界的 Unix 毫秒，零值表示不限。
// Match 在读事务之外执行，可自行解码、筛选任意业务数据；翻页时须保持筛选逻辑一致。
type Query struct {
	Start  int64                      `json:"start"`
	End    int64                      `json:"end"`
	Prefix string                     `json:"prefix"`
	Desc   bool                       `json:"desc"`
	Limit  int                        `json:"limit"`
	Cursor string                     `json:"cursor"`
	Match  func(Record) (bool, error) `json:"-"`
}

// Page 最多返回 Limit 条，另按 16 MiB 数据量分页；单条超过该大小时单独成页。
type Page struct {
	Items      []Record `json:"items"`
	NextCursor string   `json:"next_cursor,omitempty"`
}
