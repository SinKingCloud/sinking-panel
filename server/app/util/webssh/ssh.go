package webssh

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"sync"
	"time"
)

func NewSshClient(ip string, port int, timeout time.Duration) *SshClient {
	return &SshClient{
		Addr:       ip,
		Port:       port,
		Timeout:    timeout,
		sessionMap: &sync.Map{},
	}
}

type SshClient struct {
	Addr       string //连接地址
	Port       int    //连接端口
	User       string //账户
	Password   string //密码
	Key        string //密钥
	Timeout    time.Duration
	client     *ssh.Client
	sessionMap *sync.Map
}

type SshSession struct {
	Session   *ssh.Session
	StdinPipe io.WriteCloser
	write     *sshBufWriter
}

type sshBufWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *sshBufWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
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
	addr := fmt.Sprintf("%s:%d", s.Addr, s.Port)
	c, err := ssh.Dial("tcp", addr, config)
	if err == nil {
		s.client = c
		return nil
	}
	return err
}

// Session 获取一个session
func (s *SshClient) Session(sessionID string, height int, width int) (*SshSession, error) {
	if s.sessionMap == nil {
		s.sessionMap = &sync.Map{}
	}
	var session *SshSession
	var err error
	if values, ok := s.sessionMap.Load(sessionID); ok {
		session = values.(*SshSession)
	} else {
		session, err = s.NewSession(height, width)
		if err != nil {
			return nil, err
		}
		s.sessionMap.Store(sessionID, session)
	}
	return session, err
}

// NewSession 新建session
func (s *SshClient) NewSession(height int, width int) (*SshSession, error) {
	temp, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	stdinPipe, err := temp.StdinPipe()
	if err != nil {
		return nil, err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err = temp.RequestPty("xterm", height, width, modes); err != nil {
		return nil, err
	}
	if err = temp.Shell(); err != nil {
		return nil, err
	}
	write := new(sshBufWriter)
	temp.Stdout = write
	temp.Stderr = write
	session := &SshSession{
		Session:   temp,
		StdinPipe: stdinPipe,
		write:     write,
	}
	return session, nil
}

// Close 关闭client
func (s *SshClient) Close() error {
	return s.client.Close()
}

// Exec 执行命令
func (s *SshClient) Exec(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	output, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Close 关闭session
func (s *SshSession) Close() error {
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

// Write 写入数据
func (s *SshSession) Read() []byte {
	if s.write != nil && s.write.buffer.Len() != 0 {
		defer s.write.buffer.Reset()
		return s.write.buffer.Bytes()
	}
	return nil
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
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := s.Session.RequestPty("xterm", height, width, modes); err != nil {
		return err
	}
	return nil
}
