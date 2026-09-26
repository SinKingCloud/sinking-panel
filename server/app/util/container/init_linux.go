//go:build linux && cgo

package container

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/opencontainers/runc/libcontainer"
	"golang.org/x/sys/unix"
)

const initOutputHeader = "sinking-container-init-stderr\x00\n"

// Init 在 Go 包初始化后恢复真实 stderr，再交给 libcontainer 进入容器。
func Init() {
	header := make([]byte, len(initOutputHeader))
	if _, err := os.Stderr.ReadAt(header, 0); err == nil && bytes.Equal(header, []byte(initOutputHeader)) {
		if err := unix.Dup3(3, 2, 0); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "恢复容器标准错误输出失败: %v\n", err)
			os.Exit(255)
		}
		// libcontainer 保留 ExtraFiles；标记为 exec 时关闭，避免传给容器应用。
		unix.CloseOnExec(3)
	}
	libcontainer.Init()
}

// prepareInitOutput 暂存包初始化输出；启动后立即恢复原输出，不过滤应用日志。
func (m *Manager) prepareInitOutput(process *libcontainer.Process) (func() error, error) {
	stderr, ok := process.Stderr.(*os.File)
	if !ok || stderr == nil {
		return nil, errors.New("容器初始化的标准错误输出必须为文件")
	}
	if len(process.ExtraFiles) != 0 {
		return nil, errors.New("容器初始化不能重复传入额外文件")
	}
	output, err := os.CreateTemp(m.runtimeRoot, "init-stderr-")
	if err != nil {
		return nil, fmt.Errorf("创建容器初始化输出失败: %w", err)
	}
	// 仅通过文件描述符传递，不留下临时文件，也不依赖容器内的目录权限。
	if err := os.Remove(output.Name()); err != nil {
		return nil, errors.Join(fmt.Errorf("移除容器初始化临时文件失败: %w", err), output.Close())
	}
	if _, err := output.WriteString(initOutputHeader); err != nil {
		return nil, errors.Join(fmt.Errorf("写入容器初始化标记失败: %w", err), output.Close())
	}
	extraFiles := process.ExtraFiles
	process.ExtraFiles = []*os.File{stderr}
	process.Stderr = output
	return func() (returnErr error) {
		process.Stderr = stderr
		process.ExtraFiles = extraFiles
		defer func() { returnErr = errors.Join(returnErr, output.Close()) }()
		if _, err := output.Seek(int64(len(initOutputHeader)), io.SeekStart); err != nil {
			return fmt.Errorf("读取容器初始化输出失败: %w", err)
		}
		reader := bufio.NewReader(output)
		filtered, longLine := false, false
		for {
			line, readErr := reader.ReadSlice('\n')
			skip := false
			if !filtered && !longLine && readErr != bufio.ErrBufferFull {
				var entry map[string]interface{}
				if json.Unmarshal(line, &entry) == nil && entry["level"] == "warn" &&
					entry["msg"] == "unable to determine directory for user configuration; falling back to current directory" &&
					entry["error"] == "neither $XDG_CONFIG_HOME nor $HOME are defined" {
					skip, filtered = true, true
				}
			}
			if !skip && len(line) != 0 {
				if _, err := stderr.Write(line); err != nil {
					return fmt.Errorf("转发容器初始化输出失败: %w", err)
				}
			}
			longLine = readErr == bufio.ErrBufferFull
			if readErr != nil && !longLine {
				if errors.Is(readErr, io.EOF) {
					return nil
				}
				return fmt.Errorf("读取容器初始化输出失败: %w", readErr)
			}
		}
	}, nil
}
