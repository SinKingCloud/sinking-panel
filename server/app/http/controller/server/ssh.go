package server

import (
	"encoding/json"
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

type sshConnection struct {
	client     *webssh.SshClient
	session    *webssh.SshSession
	connection *websocket.Conn
	writeMu    sync.Mutex
	closeOnce  sync.Once
	done       chan struct{}
}

var sshUpgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	WriteBufferSize:  4096,
	CheckOrigin:      func(*http.Request) bool { return true },
}

func (s *sshConnection) send(messageType int, payload []byte) error {
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

func (s *sshConnection) close() {
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
	s, err := service.Server.FindById(form.Id)
	if err != nil || s == nil {
		c.Error("获取服务器信息失败")
		return
	}
	if strings.TrimSpace(s.User) == "" || strings.TrimSpace(s.Password) == "" {
		c.Error("SSH连接配置不完整")
		return
	}
	client := webssh.NewSshClient(s.Ip, s.Port, 10*time.Second)
	if s.AuthType == server_auth_type.Password {
		err = client.AuthWithPassword(s.User, s.Password)
	} else {
		err = client.AuthWithPrivateKey(s.User, s.Password)
	}
	if err != nil {
		_ = client.Close()
		c.Error(err.Error())
		return
	}
	response := make(http.Header)
	if protocols := websocket.Subprotocols(c.Request); len(protocols) > 0 {
		response.Set("Sec-WebSocket-Protocol", protocols[0])
	}
	connection, err := sshUpgrader.Upgrade(c.Writer, c.Request, response)
	if err != nil {
		_ = client.Close()
		return
	}
	ssh := &sshConnection{client: client, connection: connection, done: make(chan struct{})}
	ssh.session, err = client.NewSession(form.Height, form.Width)
	if err != nil {
		ssh.close()
		return
	}
	defer ssh.close()

	name := s.Name
	if name == "" {
		name = s.Ip
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventLogin, "连接SSH终端", "连接服务器["+name+"]")
	_ = connection.SetReadDeadline(time.Now().Add(sshPongTimeout))
	connection.SetReadLimit(sshReadLimit)
	connection.SetPongHandler(func(string) error {
		return connection.SetReadDeadline(time.Now().Add(sshPongTimeout))
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
				if err := connection.WriteControl(websocket.PingMessage, nil, time.Now().Add(sshWriteTimeout)); err != nil {
					ssh.close()
					return
				}
			case <-ssh.done:
				return
			}
		}
	}()

	for {
		_, payload, readErr := connection.ReadMessage()
		if readErr != nil {
			break
		}
		var data struct {
			Event   string `json:"event"`
			Content string `json:"content"`
		}
		if json.Unmarshal(payload, &data) != nil {
			continue
		}
		switch data.Event {
		case "ping":
			err = ssh.send(websocket.TextMessage, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
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
