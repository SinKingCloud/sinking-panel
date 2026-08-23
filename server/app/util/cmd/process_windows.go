package cmd

import (
	"os/exec"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func prepareCommand(_ *exec.Cmd) {}

type commandControl struct {
	command *exec.Cmd
	job     windows.Handle
}

func newCommandControl(command *exec.Cmd) *commandControl {
	control := &commandControl{command: command}
	job, err := windows.CreateJobObject(nil, nil)
	if err == nil {
		process, openErr := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(command.Process.Pid))
		if openErr == nil {
			err = windows.AssignProcessToJobObject(job, process)
			_ = windows.CloseHandle(process)
		} else {
			err = openErr
		}
		if err == nil {
			control.job = job
		} else {
			_ = windows.CloseHandle(job)
		}
	}
	return control
}

func (c *commandControl) terminate() error {
	if c.command.Process == nil {
		return nil
	}
	return exec.Command("taskkill", "/PID", strconv.Itoa(c.command.Process.Pid), "/T").Run()
}

func (c *commandControl) kill() error {
	if c.job != 0 {
		return windows.TerminateJobObject(c.job, 1)
	}
	if c.command.Process == nil {
		return nil
	}
	return exec.Command("taskkill", "/PID", strconv.Itoa(c.command.Process.Pid), "/T", "/F").Run()
}

func (c *commandControl) alive() bool {
	if c.job != 0 {
		var info struct {
			TotalUserTime             int64
			TotalKernelTime           int64
			ThisPeriodTotalUserTime   int64
			ThisPeriodTotalKernelTime int64
			TotalPageFaultCount       uint32
			TotalProcesses            uint32
			ActiveProcesses           uint32
			TotalTerminatedProcesses  uint32
		}
		if windows.QueryInformationJobObject(c.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info)), nil) == nil {
			return info.ActiveProcesses > 0
		}
	}
	if c.command.Process == nil {
		return false
	}
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(c.command.Process.Pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	event, err := windows.WaitForSingleObject(handle, 0)
	return err == nil && event == 0x102
}

func (c *commandControl) close() {
	if c.job != 0 {
		_ = windows.CloseHandle(c.job)
	}
}
