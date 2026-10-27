package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"server/app/util/ipc"
)

// control 使用固定本机端点，PID 只作为响应返回，不写文件。
func (u *Daemon) control(action string, timeout time.Duration) (int, error) {
	if timeout <= 0 {
		timeout = defaultStopTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	connection, err := ipc.Dial(ctx, u.ControlPath)
	if err != nil {
		return 0, err
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(timeout))
	if _, err = fmt.Fprintln(connection, action); err != nil {
		return 0, err
	}
	var response controlResponse
	if err = json.NewDecoder(io.LimitReader(connection, 4096)).Decode(&response); err != nil {
		return 0, fmt.Errorf("读取服务控制结果失败: %w", err)
	}
	if response.Error != "" {
		return 0, errors.New(response.Error)
	}
	return response.Pid, nil
}

func (u *Daemon) serve(listener net.Listener, stop func(), done <-chan struct{}, closed chan<- struct{}) {
	defer close(closed)
	var requests sync.WaitGroup
	defer requests.Wait()
	for {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		requests.Add(1)
		go func() {
			defer requests.Done()
			defer connection.Close()
			_ = connection.SetReadDeadline(time.Now().Add(5 * time.Second))
			command, err := bufio.NewReaderSize(connection, 128).ReadSlice('\n')
			if err != nil {
				return
			}
			response := controlResponse{Pid: os.Getpid()}
			value := strings.TrimSpace(string(command))
			switch {
			case value == "status":
			case strings.HasPrefix(value, "stop "):
				pid, err := strconv.Atoi(strings.TrimPrefix(value, "stop "))
				if err != nil || pid != os.Getpid() {
					response.Error = "服务进程已变更，请重新执行操作"
				} else {
					stop()
					<-done
				}
			default:
				response.Error = "不支持的进程控制命令"
			}
			_ = connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_ = json.NewEncoder(connection).Encode(response)
		}()
	}
}
