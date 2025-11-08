package server

import (
	"encoding/json"
	"github.com/SinKingCloud/sinking-go/sinking-websocket"
	"github.com/gorilla/websocket"
	"net/http"
	"server/app/constant"
	"server/app/model"
	"server/app/service"
	server2 "server/app/service/server"
	"server/app/util/context"
	"server/app/util/webssh"
	"strconv"
	"strings"
	"time"
)

func Ssh(c *context.Context) {
	type Form struct {
		Id     int `json:"id" default:"" validate:"omitempty,numeric,min=1" label:"记录ID"`
		Width  int `json:"width" default:"200" validate:"omitempty,numeric,min=1" label:"宽度"`
		Height int `json:"height" default:"120" validate:"omitempty,numeric,min=1" label:"高度"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
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
			AuthType: c.GetIntWithDefault(configs[constant.SshAuthType], int(server2.Password)),
			Password: configs[constant.SshPassword],
		}
		if s.Ip == "" || s.User == "" {
			c.Error("获取服务器信息失败")
			return
		}
	}
	client := webssh.NewSshClient(s.Ip, s.Port, 10*time.Second)
	if s.AuthType == int(server2.Password) {
		err = client.AuthWithPassword(s.User, s.Password)
	} else {
		err = client.AuthWithPrivateKey(s.User, s.Password)
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	type Message struct {
		Event   string `json:"event"`   //消息事件
		Content string `json:"content"` //消息内容
	}
	var session *webssh.SshSession
	exit := make(chan bool)
	wsServer := sinking_websocket.WebSocket{
		OnError: func(id string, err error) {
			_ = session.Close()
			close(exit)
		},
		OnConnect: func(id string, ws *sinking_websocket.Conn) {
			session, err = client.NewSession(form.Height, form.Width)
			if err == nil {
				go func() {
					tick := time.NewTicker(10 * time.Millisecond)
					defer tick.Stop()
					for {
						select {
						case <-tick.C:
							if payload := session.Read(); payload != nil {
								_ = ws.WriteMessage(websocket.TextMessage, payload)
							}
						case <-exit:
							return
						}
					}
				}()
			}
		},
		OnClose: func(id string, err error) {
			_ = session.Close()
			close(exit)
		},
		OnMessage: func(id string, ws *sinking_websocket.Conn, messageType int, data []byte) {
			var msg *Message
			if json.Unmarshal(data, &msg) == nil {
				switch msg.Event {
				case "ping":
					_ = ws.WriteMessage(websocket.TextMessage, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
					break
				case "resize":
					arr := strings.Split(msg.Content, "|")
					if len(arr) == 2 {
						width, _ := strconv.Atoi(arr[0])
						height, _ := strconv.Atoi(arr[1])
						if width > 0 && height > 0 {
							_ = session.Resize(height, width)
						}
					}
					break
				case "write":
					_ = session.Write([]byte(msg.Content))
					break
				}
			}
		},
	}
	token := c.Request.Header.Get("Sec-Websocket-Protocol")
	resp := make(http.Header)
	if token != "" {
		resp.Set("Sec-Websocket-Protocol", token)
	}
	wsServer.Listen(c.Writer, c.Request, resp)
}
