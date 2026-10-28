# log：按字节游标浏览和清空日志

导入时建议使用别名 `fileLog "server/app/util/log"`，避免和标准库 `log` 混淆。本包不生成日志、不负责滚动文件；它为前端日志查看器提供历史翻页、增量轮询和清空操作。

## Read API

```go
func Read(path string, after, before int64, pageSize int, previous ...string) (map[string]interface{}, error)
```

| 参数 | 用法 |
| --- | --- |
| `path` | 当前日志文件路径 |
| `after` | 大于 0 时，从此字节游标向后读新增内容；优先级高于 before |
| `before` | after 为 0 且 before 大于 0 时，读取此游标之前的历史 |
| 两个游标均为 0 | 读取最新一页 |
| `pageSize` | 目标行数；小于 1 使用 300，大于 10000 截为 10000 |
| `previous` | 当前文件之前的未压缩滚动文件，按旧到新的顺序传入 |

返回字段：

| 字段 | Go 动态类型 | 含义 |
| --- | --- | --- |
| `file_name` | `string` | 当前文件的基名 |
| `lines` | `[]string` | 按时间顺序排列的行内容 |
| `cursor` | `int64` | 后续增量读取的游标 |
| `start_cursor` | `int64` | 当前页起点，用于向前翻页 |
| `has_previous` | `bool` | 当前页之前是否仍有内容 |
| `end` | `bool` | 当前读取方向是否到边界；首次/历史页对应最早端，增量页对应最新端 |

游标表示 `previous... + path` 拼接后的字节偏移，不是行号。滚动文件集合或顺序改变后，不应继续把旧游标解释为相同位置。文件不存在时忽略该段并返回空页，不会创建文件。

## 示例：最新内容、增量和历史

```go
package example

import fileLog "server/app/util/log"

func ReadLatest(path string) (map[string]interface{}, error) {
	return fileLog.Read(path, 0, 0, 200)
}

func ReadNext(path string, cursor int64) (map[string]interface{}, error) {
	return fileLog.Read(path, cursor, 0, 200)
}

func ReadPrevious(path string, startCursor int64) (map[string]interface{}, error) {
	if startCursor <= 0 {
		// 已经到最早端，避免传入两个 0 后又读取最新页。
		return map[string]interface{}{"lines": []string{}, "has_previous": false}, nil
	}
	return fileLog.Read(path, 0, startCursor, 200)
}
```

有滚动文件时，所有请求都应传入一致且有序的 `previous` 参数，例如 `Read(current, cursor, 0, 200, old1, old2)`。该函数不会解压 `.gz`、`.zst`；不要直接把压缩文件作为 previous 传入。

## 读取限制

- 单次最多扫描 8 MiB，逐块读取大小为 64 KiB，因此不承诺一次恰好返回 pageSize 行。
- 单行最多保留 256 KiB；过长或读取窗口左侧不完整的行使用 `[日志行过长，已省略]` 替代。
- CRLF 行尾的 `\r` 会移除；无效 UTF-8 替换为替代字符。返回内容用于文本展示，不是原始日志的二进制导出。
- 当前文件尾部未完成的物理行也可能被显示，后续追加不能假设每次返回都是新的完整业务事件。
- after 超过新的总文件长度时，会重新读取最新一页，适应截断后的恢复。
- 路径必须指向普通文件。路径授权、目录边界和读写/滚动时的并发协调由调用方负责。

## Clear API

`Clear(path string) error` 截断当前文件，并删除同目录中匹配 `基名-2006-01-02T15-04-05.000-size/time.扩展名` 的滚动日志，包含 `.gz` 和 `.zst` 后缀。它不是只清理当前文件。

不存在的文件会创建为 0600，已有当前文件会调整为 0600；不自动创建父目录。空路径或符号链接会返回错误。部分滚动文件删除失败时，当前文件可能已经清空，错误通过 `errors.Join` 汇总。

调用 Clear 应放在已鉴权的管理操作中；与负责写入/滚动的组件协调锁，完成后让查看器清除旧游标。不要把任意请求路径直接传给 Clear。

相关：[file](../file/README.md)、[process](../process/README.md)。
