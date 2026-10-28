package oss

const (
	MinPartSize     int64 = 5 << 20
	MaxPartSize     int64 = 5 << 30
	MaxParts              = 10000
	DefaultPartSize int64 = 16 << 20
	// MaxObjectSize 采用五家厂商均可接受的保守上限 5 TiB。
	MaxObjectSize int64 = 5 << 40
)
