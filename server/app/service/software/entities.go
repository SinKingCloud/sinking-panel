package software

import (
	"server/app/model"
	repositorySoftware "server/app/repository/software"
	"sync"
	"sync/atomic"
)

// Info 返回软件公开信息及页面入口。
type Info struct {
	repositorySoftware.Software
	Frontend string `json:"frontend"`
}

type entry struct {
	mu      sync.RWMutex
	busy    atomic.Bool
	taskID  string          // 仅用于恢复排队中的卸载任务，由 mu 保护。
	pending *model.Software // 回调开始前保留原记录，排队任务取消时恢复。
}
