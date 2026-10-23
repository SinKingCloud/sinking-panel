package mongo

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// Version 返回服务器版本和功能支持情况；读取失败时保留真实错误。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	var info bson.M
	if err := m.database.RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&info); err != nil {
		return nil, m.cleanError(err)
	}
	version, _ := info["version"].(string)
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return nil, errors.New("无法识别 MongoDB 服务器版本")
	}
	major, majorErr := strconv.Atoi(parts[0])
	minor, minorErr := strconv.Atoi(parts[1])
	if majorErr != nil || minorErr != nil || major < 1 || minor < 0 {
		return nil, errors.New("无法识别 MongoDB 服务器版本")
	}
	supports := func(requiredMajor, requiredMinor int) bool {
		return major > requiredMajor || major == requiredMajor && minor >= requiredMinor
	}
	return map[string]interface{}{
		"name":    "MongoDB",
		"version": version,
		"features": map[string]bool{
			"accounts":       supports(2, 6), // 当前账户操作使用 2.6 引入的用户和角色命令。
			"views":          supports(3, 4),
			"aggregate":      supports(2, 2),
			"aggregate_32":   supports(3, 2), // $lookup、$sample。
			"aggregate_34":   supports(3, 4), // $addFields、$facet、$count 等。
			"aggregate_42":   supports(4, 2), // $set、$unset、$replaceWith。
			"aggregate_44":   supports(4, 4), // $unionWith。
			"validation":     supports(3, 2),
			"collation":      supports(3, 4),
			"ttl_indexes":    supports(2, 2),
			"text_indexes":   supports(2, 4),
			"hashed_indexes": supports(2, 4),
			"sphere_indexes": supports(2, 4),
		},
	}, nil
}
