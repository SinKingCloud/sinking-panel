package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	rand2 "math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// scriptExec 安全脚本执行器
// writeLog: 日志回调函数，参数为实时日志内容
// TempDirPattern: 临时目录名前缀
// maxScriptSize: 允许的最大脚本大小（字节）
// timeout: 默认执行超时时间（秒）
type scriptExec struct {
	tempPath      string
	writeLog      func(string)
	maxScriptSize int64
	timeout       int
}

// NewScriptExec 创建脚本执行器实例
// writeLog: 日志回调函数，可以为nil
func NewScriptExec(TempPath string, timeout int, writeLog func(string)) *scriptExec {
	return &scriptExec{
		tempPath:      TempPath,
		maxScriptSize: 1024 * 1024 * 2,
		timeout:       timeout,
		writeLog:      writeLog,
	}
}

// Execute 执行脚本内容
// script: 要执行的脚本内容
// timeout: 可选参数，执行超时时间（秒）
// 返回值: 标准输出内容，标准错误内容，错误对象
func (se *scriptExec) Execute(script string) (string, string, error) {
	if len(script) == 0 {
		return "", "", errors.New("脚本内容不能为空")
	}
	if int64(len(script)) > se.maxScriptSize {
		return "", "", fmt.Errorf("脚本大小超过限制(%d字节)", se.maxScriptSize)
	}
	path := se.tempPath + "/exec/"
	if _, err := os.Stat(path); err != nil {
		_ = os.MkdirAll(path, 0755)
	}
	tempDir, err := os.MkdirTemp(path, "*")
	if err != nil {
		return "", "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)
	scriptPath, err := se.createScriptFile(tempDir, script)
	if err != nil {
		return "", "", err
	}
	return se.executeScript(scriptPath, se.timeout)
}

// createScriptFile 创建可执行脚本文件
func (se *scriptExec) createScriptFile(dir, content string) (string, error) {
	ext := ".sh"
	if runtime.GOOS == "windows" {
		ext = ".bat"
	}

	scriptFile := filepath.Join(dir, "script_"+se.secureRandomString(8)+ext)
	if err := os.WriteFile(scriptFile, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("写入脚本文件失败: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(scriptFile, 0700); err != nil {
			return "", fmt.Errorf("设置执行权限失败: %w", err)
		}
	}

	return scriptFile, nil
}

// executeScript 执行脚本核心逻辑
func (se *scriptExec) executeScript(path string, timeout int) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", path)
	} else {
		cmd = exec.CommandContext(ctx, "bash", path)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("获取标准输出管道失败: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("获取错误输出管道失败: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("启动进程失败: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var stdoutBuf, stderrBuf bytes.Buffer
	go func() { defer wg.Done(); se.scanOutput(stdoutPipe, &stdoutBuf) }()
	go func() { defer wg.Done(); se.scanOutput(stderrPipe, &stderrBuf) }()

	execErr := cmd.Wait()
	wg.Wait()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", "", fmt.Errorf("执行超时（%d秒）", timeout)
	}

	if execErr != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("执行错误: %w", execErr)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// scanOutput 实时扫描输出流并记录日志
func (se *scriptExec) scanOutput(rc io.ReadCloser, buf *bytes.Buffer) {
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line + "\n")
		if se.writeLog != nil {
			se.writeLog(line)
		}
	}
	if err := scanner.Err(); err != nil && se.writeLog != nil {
		return
	}
}

// secureRandomString 生成安全随机字符串
func (se *scriptExec) secureRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str := make([]byte, n)
	if _, err := rand.Read(str); err != nil {
		return strconv.FormatInt(time.Now().UnixMilli(), 10) + strconv.Itoa(100000+rand2.Intn(899999))
	}
	for i, b := range str {
		str[i] = letters[b%byte(len(letters))]
	}
	return string(str)
}
