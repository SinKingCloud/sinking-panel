package script

import (
	"context"
	"errors"
	"server/app/constant"
	"server/app/util/cmd"
	"strings"
)

// Validate 校验脚本内容，保留原始缩进和换行。
func (m *Manager) Validate(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("任务内容不能为空")
	}
	return content, nil
}

// Run 执行脚本并实时写入日志，上下文结束时停止脚本进程。
func (m *Manager) Run(ctx context.Context, content string, writeLog func(string)) error {
	executor := cmd.NewScriptExec(constant.CmdPath, 43200, writeLog)
	_, _, err := executor.ExecuteContext(ctx, content)
	return err
}
