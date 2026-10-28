# job：生产者与消费者任务

`server/app/util/job` 将一个生产者和多个消费者组成可取消的后台任务，适用于扫描文件、批量处理记录等工作。它是进程内工作队列，不持久化任务，也不负责进程保活；进程保活请使用 [process](../process/README.md)。

## 生命周期和 API

| API | 行为 |
| --- | --- |
| `NewTask() *Task` | 创建任务，默认消费者数量为 1 |
| `SetThread(num int) *Task` | 配置消费者数量，非正数按 1 处理 |
| `SetProducer(func(context.Context, chan interface{}) error) *Task` | 设置唯一的生产者，返回 Task 以便链式调用 |
| `SetConsumer(func(context.Context, interface{}) error) *Task` | 设置消费者，每个任务项由一个消费者处理 |
| `Start() error` | 使用后台 context 异步启动 |
| `StartWithContext(ctx context.Context) error` | 使用调用方 context 异步启动；nil 按后台 context 处理 |
| `Wait()` / `WaitWithTimeout(timeout time.Duration) error` | 等待当前一轮结束；等待超时不会自动停止任务 |
| `Stop() error` / `StopWithTimeout(timeout time.Duration) error` | 取消当前一轮并等待退出；Stop 默认最多等 30 秒 |
| `GetStatus() TaskStatus` / `IsRunning() bool` | 读取状态；Stopping 也属于运行中 |
| `GetStats() TaskStats` | 返回当前一轮的统计快照 |

状态为 `TaskStopped`、`TaskRunning`、`TaskStopping`、`TaskCompleted`。重复启动运行中的任务会返回错误；一轮结束后可以再次 Start，统计重新计数。启动时复制生产者、消费者和并发数，运行中调用 setter 不会修改已经启动的这一轮。

启动前应设置生产者和消费者；未设置生产者时任务会直接结束，未设置消费者时任务项会被跳过，不会计入成功或失败。内部通道容量为消费者数量的两倍。**生产者不要关闭 channel**，包会在生产者返回后关闭它，并等待全部消费者退出。不要在生产者返回后继续向 channel 发送数据。

## 示例：并行处理任务

```go
package example

import (
	"context"
	"fmt"
	"server/app/util/job"
)

func ProcessNumbers(ctx context.Context) error {
	task := job.NewTask().SetThread(4)
	task.SetProducer(func(ctx context.Context, output chan interface{}) error {
		for i := 0; i < 100; i++ {
			select {
			case output <- i:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})
	task.SetConsumer(func(ctx context.Context, value interface{}) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		number := value.(int) // 生产者约定只发送 int。
		fmt.Println(number * number)
		return nil
	})
	if err := task.StartWithContext(ctx); err != nil {
		return err
	}
	task.Wait()
	stats := task.GetStats()
	if err := ctx.Err(); err != nil {
		return err
	}
	if stats.FailedJobs > 0 {
		return fmt.Errorf("有 %d 项处理失败", stats.FailedJobs)
	}
	return nil
}
```

## 统计、错误和停止语义

- `ProcessedJobs` 只统计消费者返回 nil 的项；`FailedJobs` 统计消费者返回错误的项。单项失败不会中止其他项，也不会自动重试。
- `ActiveWorkers` 是尚未退出的工作协程数量，不是正在执行回调的数量。`StartTime`、`EndTime` 记录一轮起止时间。
- 当前实现忽略生产者返回的错误。需要传回生产失败原因时，由调用方使用结果 channel、受锁保护的错误变量或取消原因单独记录，再在 Wait 后读取。
- 取消后，工作协程继续排空通道但跳过尚未处理的项；正在执行的消费者必须自行响应 context。外部 context 取消后，状态可能最终显示 Completed，不能只根据状态判断业务成功。
- Stop 超时只表示未等到退出，不会强制终止回调。不要在同一个任务的生产者/消费者内调用阻塞的 Stop 或 Wait。
- 回调中的 panic 不会被恢复。共享结果应由调用方加锁，不要在运行时直接写入导出的 `Thread`、`Producer`、`Consumer` 字段，优先使用 setter。

## 集成建议

业务层负责存储任务 ID、错误详情和进度；本包仅负责一轮进程内调度。耗时 I/O 应使用带 context 的接口。服务关闭时先停止生产者，再等待任务退出，避免已退出服务仍有协程写资源。

相关：[cmd](../cmd/README.md)、[process](../process/README.md)、[oss](../oss/README.md)。
