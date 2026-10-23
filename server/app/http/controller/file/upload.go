package file

import (
	"net/http"
	"server/app/util/context"
	"time"
)

// Upload 文件上传处理
func Upload(c *context.Context) {
	// 大文件上传和合并单独放宽超时，不影响其他接口的一分钟限制。
	deadline := time.Now().Add(15 * time.Minute)
	controller := http.NewResponseController(c.Writer)
	_ = controller.SetReadDeadline(deadline)
	_ = controller.SetWriteDeadline(deadline)
	switch c.DefaultQuery("action", "upload") {
	case "upload":
		handleFileUpload(c)
	case "merge":
		handleMergeChunks(c)
	case "check":
		handleCheckChunks(c)
	case "clear":
		handleClearUpload(c)
	default:
		c.Error("不支持的操作类型")
	}
}
