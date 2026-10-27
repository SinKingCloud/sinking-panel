package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"server/app/constant"
	"server/app/service/software"
	"server/app/util/ipc"
)

// pluginCommandRequest 通过本机通信传递操作及服务器路径。
type pluginCommandRequest struct {
	Action string `json:"action"`
	Target string `json:"target,omitempty"`
}

type pluginCommandResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// runPluginCommand 统一离线命令和本机通信的执行结果，业务仍由软件服务处理。
func (s *Server) runPluginCommand(service software.Service, action, target string) (string, error) {
	switch action {
	case "install":
		installed, err := service.Install(target)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("插件已安装: %s %s", installed.Code, installed.Version), nil
	case "uninstall":
		if _, err := service.Uninstall(target); err != nil {
			return "", err
		}
		return "卸载完成", nil
	default:
		return "", errors.New("不支持的插件命令")
	}
}

// sendPluginCommand 将安装、卸载交给正在运行的面板；通信服务未运行时返回 false，供命令行离线执行。
func (s *Server) sendPluginCommand(ctx context.Context, action, target string) (bool, string, error) {
	dial, cancel := context.WithTimeout(ctx, 5*time.Second)
	connection, err := ipc.Dial(dial, constant.PluginCommandPath)
	cancel()
	if errors.Is(err, ipc.ErrUnavailable) {
		return false, "", nil
	}
	if err != nil {
		return true, "", fmt.Errorf("连接面板插件服务失败: %w", err)
	}
	defer connection.Close()
	stop := context.AfterFunc(ctx, func() { _ = connection.Close() })
	defer stop()
	_ = connection.SetDeadline(time.Now().Add(30 * time.Second))
	request := pluginCommandRequest{
		Action: action,
		Target: target,
	}
	if err = json.NewEncoder(connection).Encode(request); err != nil {
		return true, "", fmt.Errorf("提交插件命令失败，请查看插件列表确认状态: %w", err)
	}
	// 命令行中断等待不会再次提交或中断面板内的操作。
	_ = connection.SetDeadline(time.Now().Add(11 * time.Minute))
	var response pluginCommandResponse
	if err = json.NewDecoder(io.LimitReader(connection, 64<<10)).Decode(&response); err != nil {
		return true, "", fmt.Errorf("等待插件操作结果失败，操作可能仍在面板执行，请查看插件列表确认状态: %w", err)
	}
	if response.Error != "" {
		return true, "", errors.New(response.Error)
	}
	return true, response.Message, nil
}

// listenPluginCommand 接收本机插件命令；关闭函数停止接收并等待在途操作完成，再由调用方关闭业务服务。
func (s *Server) listenPluginCommand(service software.Service) (func(), error) {
	listener, err := ipc.Listen(constant.PluginCommandPath)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		var requests sync.WaitGroup
		defer requests.Wait()
		for {
			connection, err := listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Printf("接收插件命令失败: %v", err)
				}
				return
			}
			requests.Add(1)
			go func() {
				defer requests.Done()
				defer connection.Close()
				_ = connection.SetReadDeadline(time.Now().Add(30 * time.Second))
				var request pluginCommandRequest
				var response pluginCommandResponse
				err := json.NewDecoder(io.LimitReader(connection, 16<<10)).Decode(&request)
				if err != nil {
					err = fmt.Errorf("读取插件命令失败: %w", err)
				} else {
					response.Message, err = s.runPluginCommand(service, request.Action, request.Target)
				}
				if err != nil {
					response.Error = err.Error()
				}
				_ = connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
				_ = json.NewEncoder(connection).Encode(response)
			}()
		}
	}()
	var closeOnce sync.Once
	return func() {
		closeOnce.Do(func() {
			_ = listener.Close()
			<-done
		})
	}, nil
}
