//go:build !windows

package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os/exec"
	"strings"
	"syscall"
)

func newScriptCommand(path string) *exec.Cmd {
	command := exec.Command("bash", path)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return command
}

// scanOutput 实时读取完整输出行并记录日志。
func (se *scriptExec) scanOutput(rc io.ReadCloser, buf *bytes.Buffer) error {
	defer rc.Close()
	reader := bufio.NewReader(rc)
	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			buf.WriteString(line)
			buf.WriteByte('\n')
			if se.writeLog != nil {
				se.writeLog(line)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return readErr
		}
	}
}

type commandControl struct {
	pid int
}

func newCommandControl(command *exec.Cmd) *commandControl {
	return &commandControl{pid: command.Process.Pid}
}

func (c *commandControl) terminate() error {
	if c.pid <= 0 {
		return nil
	}
	err := syscall.Kill(-c.pid, syscall.SIGTERM)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func (c *commandControl) kill() error {
	if c.pid <= 0 {
		return nil
	}
	err := syscall.Kill(-c.pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func (c *commandControl) alive() bool {
	if c.pid <= 0 {
		return false
	}
	err := syscall.Kill(-c.pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func (c *commandControl) close() {}
