package controller

import (
	"github.com/SinKingCloud/sinking-go/sinking-websocket"
	"github.com/gorilla/websocket"
	"server/app/util/server"
	"server/app/util/webssh"
	"time"
)

func Ssh(c *server.Context) {
	client := webssh.NewSshClient("222.187.238.201", 22, 10*time.Second)
	err := client.AuthWithPassword("root", "qwe@123")
	if err != nil {
		return
	}
	var session *webssh.SshSession
	exit := make(chan bool)
	wsServer := sinking_websocket.WebSocket{
		OnError: func(id string, err error) {
			_ = session.Close()
			close(exit)
		},
		OnConnect: func(id string, ws *sinking_websocket.Conn) {
			session, err = client.NewSession()
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
			_ = session.Write(data)
		},
	}
	wsServer.Listen(c.Writer, c.Request, nil)
}
