package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"server/app/enum/log_type"
	"server/app/enum/server_auth_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/webssh"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	sshReadLimit    = 1024 * 1024
	sshWriteTimeout = 30 * time.Second
	sshPongTimeout  = 60 * time.Second
	sshPingInterval = 54 * time.Second
)

type connection struct {
	client     *webssh.SshClient
	session    *webssh.SshSession
	connection *websocket.Conn
	writeMu    sync.Mutex
	closeOnce  sync.Once
	done       chan struct{}
}

type message struct {
	Event   string `json:"event"`
	Content string `json:"content,omitempty"`
	Code    string `json:"code,omitempty"`
}

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	WriteBufferSize:  4096,
	CheckOrigin:      func(*http.Request) bool { return true },
}

func (s *connection) send(messageType int, payload []byte) error {
	select {
	case <-s.done:
		return websocket.ErrCloseSent
	default:
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	select {
	case <-s.done:
		return websocket.ErrCloseSent
	default:
	}
	if err := s.connection.SetWriteDeadline(time.Now().Add(sshWriteTimeout)); err != nil {
		return err
	}
	return s.connection.WriteMessage(messageType, payload)
}

func (s *connection) close() {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.connection != nil {
			_ = s.connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""), time.Now().Add(time.Second))
			_ = s.connection.Close()
		}
		if s.client != nil {
			_ = s.client.Close()
		}
		if s.session != nil {
			_ = s.session.Close()
		}
	})
}

func Ssh(c *context.Context) {
	var form struct {
		Id     int64 `json:"id" default:"0" validate:"gte=0" label:"记录ID"`
		Width  int   `json:"width" default:"200" validate:"min=1,max=1000" label:"宽度"`
		Height int   `json:"height" default:"120" validate:"min=1,max=1000" label:"高度"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	header := make(http.Header)
	if protocols := websocket.Subprotocols(c.Request); len(protocols) > 0 {
		header.Set("Sec-WebSocket-Protocol", protocols[0])
	}
	socket, err := upgrader.Upgrade(c.Writer, c.Request, header)
	if err != nil {
		return
	}
	ssh := &connection{connection: socket, done: make(chan struct{})}
	serverData, err := service.Server.FindById(form.Id)
	if err != nil || serverData == nil {
		response, marshalErr := json.Marshal(message{Event: "error", Code: "server_not_found", Content: "获取服务器信息失败"})
		if marshalErr == nil {
			_ = ssh.send(websocket.BinaryMessage, response)
		}
		ssh.close()
		return
	}
	code, errorMessage := "", ""
	if strings.TrimSpace(serverData.User) == "" {
		code, errorMessage = "credential_invalid", "SSH登录账号未配置，请补充认证信息"
	} else if strings.TrimSpace(serverData.Password) == "" {
		if serverData.AuthType == server_auth_type.Password {
			code, errorMessage = "password_invalid", "SSH密码未配置，请输入密码"
		} else {
			code, errorMessage = "private_key_invalid", "SSH私钥未配置，请输入私钥"
		}
	} else {
		ssh.client = webssh.NewSshClient(serverData.Ip, serverData.Port, 10*time.Second)
		if serverData.AuthType == server_auth_type.Password {
			err = ssh.client.AuthWithPassword(serverData.User, serverData.Password)
		} else {
			err = ssh.client.AuthWithPrivateKey(serverData.User, serverData.Password)
		}
		if err != nil {
			switch {
			case errors.Is(err, webssh.ErrAuthentication):
				code = "credential_invalid"
				if serverData.AuthType == server_auth_type.Password {
					errorMessage = "SSH账号或密码认证失败，请检查后重试"
				} else {
					errorMessage = "SSH账号或私钥认证失败，请检查后重试"
				}
			case errors.Is(err, webssh.ErrInvalidPrivateKey):
				code, errorMessage = "private_key_invalid", "SSH私钥格式无效，请重新配置"
			case errors.Is(err, webssh.ErrConnectionTimeout):
				code, errorMessage = "connection_timeout", "SSH连接超时，请检查地址、端口和网络"
			case errors.Is(err, webssh.ErrConnectionRefused):
				code, errorMessage = "connection_refused", "SSH服务拒绝连接，请检查地址和端口"
			case errors.Is(err, webssh.ErrHostUnreachable):
				code, errorMessage = "host_unreachable", "无法访问SSH主机，请检查地址和网络"
			case errors.Is(err, webssh.ErrHandshake):
				code, errorMessage = "handshake_failed", "SSH握手失败，请检查SSH服务和协议配置"
			default:
				code, errorMessage = "connection_failed", "SSH连接失败，请检查服务端配置"
			}
		} else if ssh.session, err = ssh.client.NewSession(form.Height, form.Width); err != nil {
			code, errorMessage = "session_failed", "SSH会话启动失败，请检查服务端终端配置"
		}
	}
	if code != "" {
		response, marshalErr := json.Marshal(message{Event: "error", Code: code, Content: errorMessage})
		if marshalErr == nil {
			_ = ssh.send(websocket.BinaryMessage, response)
		}
		ssh.close()
		return
	}
	defer ssh.close()
	response, err := json.Marshal(message{Event: "ready"})
	if err != nil || ssh.send(websocket.BinaryMessage, response) != nil {
		return
	}

	name := serverData.Name
	if name == "" {
		name = serverData.Ip
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventLogin, "连接SSH终端", "连接服务器["+name+"]")

	_ = socket.SetReadDeadline(time.Now().Add(sshPongTimeout))
	socket.SetReadLimit(sshReadLimit)
	socket.SetPongHandler(func(string) error {
		return socket.SetReadDeadline(time.Now().Add(sshPongTimeout))
	})

	sessionDone := make(chan struct{})
	go func() {
		_ = ssh.session.Wait()
		close(sessionDone)
	}()
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		ticker := time.NewTicker(sshPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ssh.session.OutputReady():
				if payload := ssh.session.Read(); payload != nil && ssh.send(websocket.BinaryMessage, payload) != nil {
					ssh.close()
					return
				}
			case <-sessionDone:
				for payload := ssh.session.Read(); payload != nil; payload = ssh.session.Read() {
					if ssh.send(websocket.BinaryMessage, payload) != nil {
						break
					}
				}
				ssh.close()
				return
			case <-ticker.C:
				if writeErr := socket.WriteControl(websocket.PingMessage, nil, time.Now().Add(sshWriteTimeout)); writeErr != nil {
					ssh.close()
					return
				}
			case <-ssh.done:
				return
			}
		}
	}()

	for {
		_, payload, readErr := socket.ReadMessage()
		if readErr != nil {
			break
		}
		var data message
		if json.Unmarshal(payload, &data) != nil {
			continue
		}
		switch data.Event {
		case "ping":
			response, err = json.Marshal(message{Event: "pong", Content: strconv.FormatInt(time.Now().Unix(), 10)})
			if err == nil {
				err = ssh.send(websocket.BinaryMessage, response)
			}
		case "resize":
			size := strings.Split(data.Content, "|")
			if len(size) == 2 {
				width, widthErr := strconv.Atoi(size[0])
				height, heightErr := strconv.Atoi(size[1])
				if widthErr == nil && heightErr == nil && width >= 1 && width <= 1000 && height >= 1 && height <= 1000 {
					err = ssh.session.Resize(height, width)
				}
			}
		case "write":
			err = ssh.session.Write([]byte(data.Content))
		}
		if err != nil {
			break
		}
	}
	ssh.close()
	<-sessionDone
	<-outputDone
}
