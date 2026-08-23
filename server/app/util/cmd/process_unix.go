//go:build !windows

package cmd

import (
	"errors"
	"os/exec"
	"syscall"
)

func prepareCommand(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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
