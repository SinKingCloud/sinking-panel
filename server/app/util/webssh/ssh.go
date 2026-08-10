package webssh

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSshClient(ip string, port int, timeout time.Duration) *SshClient {
	return &SshClient{
		Addr:    ip,
		Port:    port,
		Timeout: timeout,
	}
}

type SshClient struct {
	Addr     string //连接地址
	Port     int    //连接端口
	User     string //账户
	Password string //密码
	Key      string //密钥
	Timeout  time.Duration
	client   *ssh.Client
}

type SshSession struct {
	Session   *ssh.Session
	StdinPipe io.WriteCloser
	write     *sshBufWriter
}

type sshBufWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
	cond   *sync.Cond
	ready  chan struct{}
	closed bool
}

func newSshBufWriter() *sshBufWriter {
	write := &sshBufWriter{ready: make(chan struct{}, 1)}
	write.cond = sync.NewCond(&write.mu)
	return write
}

func (w *sshBufWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	written := 0
	for len(p) > 0 {
		for !w.closed && w.buffer.Len() >= 1024*1024 {
			w.cond.Wait()
		}
		if w.closed {
			return written, io.ErrClosedPipe
		}
		size := 1024*1024 - w.buffer.Len()
		if size > len(p) {
			size = len(p)
		}
		n, err := w.buffer.Write(p[:size])
		written += n
		p = p[n:]
		if err != nil {
			return written, err
		}
		select {
		case w.ready <- struct{}{}:
		default:
		}
	}
	return written, nil
}

func (w *sshBufWriter) Read() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer.Len() == 0 {
		return nil
	}
	payload := append([]byte(nil), w.buffer.Bytes()...)
	w.buffer.Reset()
	w.cond.Broadcast()
	return payload
}

func (w *sshBufWriter) Close() {
	w.mu.Lock()
	w.closed = true
	w.cond.Broadcast()
	w.mu.Unlock()
}

// AuthWithPassword 账号密码登录
func (s *SshClient) AuthWithPassword(user string, password string) error {
	s.User = user
	s.Password = password
	return s.Auth(s.User, ssh.Password(s.Password))
}

// AuthWithPrivateKey 证书登录
func (s *SshClient) AuthWithPrivateKey(user string, key string) error {
	s.User = user
	s.Key = key
	signer, err := ssh.ParsePrivateKey([]byte(s.Key))
	if err == nil {
		return s.Auth(s.User, ssh.PublicKeys(signer))
	}
	return err
}

// Auth 执行命令
func (s *SshClient) Auth(user string, method ssh.AuthMethod) error {
	s.User = user
	config := &ssh.ClientConfig{
		Timeout:         s.Timeout,
		User:            s.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{method}
	addr := net.JoinHostPort(s.Addr, strconv.Itoa(s.Port))
	c, err := ssh.Dial("tcp", addr, config)
	if err == nil {
		s.client = c
		return nil
	}
	return err
}

// NewSession 新建session
func (s *SshClient) NewSession(height int, width int) (*SshSession, error) {
	temp, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	stdinPipe, err := temp.StdinPipe()
	if err != nil {
		_ = temp.Close()
		return nil, err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err = temp.RequestPty("xterm", height, width, modes); err != nil {
		_ = temp.Close()
		return nil, err
	}
	write := newSshBufWriter()
	temp.Stdout = write
	temp.Stderr = write
	if err = temp.Shell(); err != nil {
		_ = temp.Close()
		return nil, err
	}
	session := &SshSession{
		Session:   temp,
		StdinPipe: stdinPipe,
		write:     write,
	}
	return session, nil
}

// Close 关闭client
func (s *SshClient) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

// Exec 执行命令
func (s *SshClient) Exec(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = session.Close()
	}()
	output, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Close 关闭session
func (s *SshSession) Close() error {
	if s == nil {
		return nil
	}
	if s.write != nil {
		s.write.Close()
	}
	if s.Session == nil {
		return nil
	}
	return s.Session.Close()
}

// Write 写入数据
func (s *SshSession) Write(payload []byte) error {
	if s.StdinPipe != nil {
		_, err := s.StdinPipe.Write(payload)
		return err
	}
	return errors.New("the stdin not init")
}

// Read 读取数据
func (s *SshSession) Read() []byte {
	if s.write != nil {
		return s.write.Read()
	}
	return nil
}

// OutputReady 终端有可读输出时通知。
func (s *SshSession) OutputReady() <-chan struct{} {
	if s == nil || s.write == nil {
		return nil
	}
	return s.write.ready
}

// Wait 等待session
func (s *SshSession) Wait() error {
	err := s.Session.Wait()
	if err != nil {
		return err
	}
	return nil
}

// Resize 重置大小
func (s *SshSession) Resize(height int, width int) error {
	return s.Session.WindowChange(height, width)
}
