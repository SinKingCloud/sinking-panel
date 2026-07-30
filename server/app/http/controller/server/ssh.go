package server

import (
	stdContext "context"
	"encoding/json"
	"errors"
	"net/http"
	"server/app/constant"
	"server/app/enum/server_auth_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/webssh"
	"strconv"
	"strings"
	"time"

	"github.com/SinKingCloud/sinking-go/sinking-websocket"
	"github.com/gorilla/websocket"
)

type sshContextKey struct{}

type sshConnection struct {
	client  *webssh.SshClient
	session *webssh.SshSession
	width   int
	height  int
}

var sshServer = sinking_websocket.NewServer(
	sinking_websocket.WithReadLimit(1024*1024),
	sinking_websocket.WithConnectHandler(func(connection *sinking_websocket.Connection) error {
		ssh, ok := connection.Request().Context().Value(sshContextKey{}).(*sshConnection)
		if !ok || ssh == nil {
			return errors.New("ssh connection context not found")
		}
		var err error
		ssh.session, err = ssh.client.NewSession(ssh.height, ssh.width)
		if err != nil {
			return err
		}
		go func() {
			tick := time.NewTicker(10 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-tick.C:
					if payload := ssh.session.Read(); payload != nil {
						if connection.Send(sinking_websocket.TextMessage, payload) != nil {
							return
						}
					}
				case <-connection.Done():
					return
				}
			}
		}()
		return nil
	}),
	sinking_websocket.WithMessageHandler(func(connection *sinking_websocket.Connection, message sinking_websocket.Message) error {
		ssh, ok := connection.Request().Context().Value(sshContextKey{}).(*sshConnection)
		if !ok || ssh == nil || ssh.session == nil {
			return errors.New("ssh connection context not found")
		}
		var data struct {
			Event   string `json:"event"`
			Content string `json:"content"`
		}
		if json.Unmarshal(message.Payload, &data) != nil {
			return nil
		}
		switch data.Event {
		case "ping":
			return connection.Send(sinking_websocket.TextMessage, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
		case "resize":
			arr := strings.Split(data.Content, "|")
			if len(arr) == 2 {
				width, _ := strconv.Atoi(arr[0])
				height, _ := strconv.Atoi(arr[1])
				if width > 0 && height > 0 {
					return ssh.session.Resize(height, width)
				}
			}
		case "write":
			return ssh.session.Write([]byte(data.Content))
		}
		return nil
	}),
	sinking_websocket.WithDisconnectHandler(func(connection *sinking_websocket.Connection, _ error) {
		ssh, ok := connection.Request().Context().Value(sshContextKey{}).(*sshConnection)
		if ok && ssh != nil && ssh.session != nil {
			_ = ssh.session.Close()
		}
	}),
)

func Ssh(c *context.Context) {
	var form struct {
		Id     int64 `json:"id" default:"" validate:"omitempty,numeric,min=1" label:"记录ID"`
		Width  int   `json:"width" default:"200" validate:"omitempty,numeric,min=1" label:"宽度"`
		Height int   `json:"height" default:"120" validate:"omitempty,numeric,min=1" label:"高度"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var s *model.Server
	var err error
	if form.Id > 0 {
		s, err = service.Server.FindById(form.Id)
		if err != nil || s == nil {
			c.Error("获取服务器信息失败")
			return
		}
	} else {
		configs := service.Config.Group(constant.SshGroup)
		s = &model.Server{
			Ip:       configs[constant.SshIP],
			Port:     c.GetIntWithDefault(configs[constant.SshPort], 22),
			User:     configs[constant.SshUser],
			AuthType: c.GetIntWithDefault(configs[constant.SshAuthType], server_auth_type.Password),
			Password: configs[constant.SshPassword],
		}
		if s.Ip == "" || s.User == "" {
			c.Error("获取服务器信息失败")
			return
		}
	}
	client := webssh.NewSshClient(s.Ip, s.Port, 10*time.Second)
	defer func() {
		_ = client.Close()
	}()
	if s.AuthType == server_auth_type.Password {
		err = client.AuthWithPassword(s.User, s.Password)
	} else {
		err = client.AuthWithPrivateKey(s.User, s.Password)
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	ssh := &sshConnection{
		client: client,
		width:  form.Width,
		height: form.Height,
	}
	resp := make(http.Header)
	protocols := websocket.Subprotocols(c.Request)
	if len(protocols) > 0 {
		resp.Set("Sec-WebSocket-Protocol", protocols[0])
	}
	request := c.Request.WithContext(stdContext.WithValue(c.Request.Context(), sshContextKey{}, ssh))
	_ = sshServer.Handle(c.Writer, request, resp)
}
