package recordstore

// Aggregate 定义一个字段在分钟记录和查询时间桶中的聚合方式。
type Aggregate string

const (
	Mean Aggregate = "mean"
	Max  Aggregate = "max"
	Min  Aggregate = "min"
	Sum  Aggregate = "sum"
	Last Aggregate = "last"
)

// Field 字段顺序也是磁盘记录中 Values 的顺序；省略 Aggregate 时使用 Mean。
type Field struct {
	Name      string    `json:"name"`
	Aggregate Aggregate `json:"aggregate"`
}

// Point 支持负数的通用数值记录。Samples 是均值权重，Sum 值已经是该批样本的合计。
type Point struct {
	Time    int64     `json:"time"`
	Samples uint32    `json:"samples"`
	Values  []float64 `json:"values"`
}

// SeriesQueryOptions 控制数值查询的输出密度；包含首尾边界时最多多返回一个桶。
// MaxPoints 为 1 至 1440 时按常用时间间隔向上选档；零值沿用默认约 1200 点的分钟间隔算法。
type SeriesQueryOptions struct {
	MaxPoints int
}

type SeriesResult struct {
	Points   []Point `json:"points"`
	Interval int64   `json:"interval"`
	Start    int64   `json:"start"`
	End      int64   `json:"end"`
}

type SeriesSettings struct {
	Enabled       bool   `json:"enabled"`
	RetentionDays int    `json:"retention_days"`
	Interval      int    `json:"interval"`
	StorageBytes  int64  `json:"storage_bytes"`
	LastError     string `json:"last_error,omitempty"`
}

// SeriesOptions 允许调用方接管配置持久化。
type SeriesOptions struct {
	Configuration     *SeriesConfiguration
	SaveConfiguration func(SeriesConfiguration) error
}

// SeriesConfiguration 是可持久化的采集配置，不包含运行状态。
// 外部存储需同时提供初始 Configuration 和 SaveConfiguration；否则使用 config.json。
// 保存回调成功后配置才生效；回调在内部锁下执行，不应再次调用当前序列或记录器。
type SeriesConfiguration struct {
	Enabled       bool `json:"enabled"`
	RetentionDays int  `json:"retention_days"`
}
