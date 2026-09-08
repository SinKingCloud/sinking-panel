package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func newScriptCommand(path string) *exec.Cmd {
	command := exec.Command("cmd.exe")
	// 路径通过环境变量展开一次，保留中文、空格以及 %、! 等字符。
	command.Env = append(os.Environ(), "SINKING_PANEL_EXEC_SCRIPT="+path)
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:       `cmd.exe /D /S /V:OFF /C ""%SINKING_PANEL_EXEC_SCRIPT%""`,
		HideWindow:    true,
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	// 独立控制台让脚本内的 chcp 不影响面板或其他并发任务，也支持后台运行。
	return command
}

// scanOutput 先将 Windows 输出转成 UTF-8，再分行写入日志。
func (se *scriptExec) scanOutput(rc io.ReadCloser, buf *bytes.Buffer) error {
	defer rc.Close()
	reader := bufio.NewReader(rc)
	prefix, _ := reader.Peek(2)
	if bytes.HasPrefix(prefix, []byte{0xef, 0xbb}) {
		prefix, _ = reader.Peek(3)
		if bytes.HasPrefix(prefix, []byte{0xef, 0xbb, 0xbf}) {
			_, _ = reader.Discard(3)
		}
	} else if bytes.HasPrefix(prefix, []byte{0xff, 0xfe}) {
		reader = bufio.NewReader(transform.NewReader(reader, unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()))
	} else if bytes.HasPrefix(prefix, []byte{0xfe, 0xff}) {
		reader = bufio.NewReader(transform.NewReader(reader, unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder()))
	} else if len(prefix) >= 2 && prefix[1] == 0 && (prefix[0] >= 0x20 && prefix[0] <= 0x7e || prefix[0] >= '\t' && prefix[0] <= '\r') {
		// cmd /U 可能不写 BOM；只识别 ASCII 字符后跟 NUL 的常见 UTF-16LE 前缀。
		reader = bufio.NewReader(transform.NewReader(reader, unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()))
	}

	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if !utf8.ValidString(line) {
				if len(line) > math.MaxInt32 {
					return errors.New("输出行过长，无法转换 Windows 代码页")
				}
				const mbErrInvalidChars = 0x8 // MB_ERR_INVALID_CHARS
				var decodeErr error
				// 无编码标记时无法通用推断；先按系统 OEM 页读取，失败后尝试 ANSI 页。
				for _, codePage := range []uint32{1, 0} { // CP_OEMCP、CP_ACP
					length, err := windows.MultiByteToWideChar(codePage, mbErrInvalidChars, unsafe.StringData(line), int32(len(line)), nil, 0)
					if err != nil {
						decodeErr = errors.Join(decodeErr, err)
						continue
					}
					decoded := make([]uint16, length)
					if length > 0 {
						length, err = windows.MultiByteToWideChar(codePage, mbErrInvalidChars, unsafe.StringData(line), int32(len(line)), &decoded[0], length)
						if err != nil {
							decodeErr = errors.Join(decodeErr, err)
							continue
						}
					}
					line = string(utf16.Decode(decoded[:length]))
					decodeErr = nil
					break
				}
				if decodeErr != nil {
					return decodeErr
				}
			}
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
