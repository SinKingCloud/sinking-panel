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
	created bool            // 排队安装创建的记录，取消时删除。
	taskID  string          // 提交时持 importing 和 mu 写入，任务结束后不清空。
	pending *model.Software // 回调开始前保留原记录，排队任务取消时恢复。
}
