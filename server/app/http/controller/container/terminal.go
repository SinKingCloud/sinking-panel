package container

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"

	"github.com/gorilla/websocket"
)

// Terminal 连接容器交互终端，复用 Manager 的 PTY，不依赖容器内安装 SSH 服务。
func Terminal(c *context.Context) {
	var form struct {
		Id     string `json:"id" validate:"required,max=128" label:"容器ID"`
		Width  int    `json:"width" default:"80" validate:"min=1,max=1000" label:"宽度"`
		Height int    `json:"height" default:"24" validate:"min=1,max=1000" label:"高度"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		WriteBufferSize:  4096,
		Subprotocols:     websocket.Subprotocols(c.Request),
		CheckOrigin:      func(*http.Request) bool { return true },
	}
	socket, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer socket.Close()
	var writeMu sync.Mutex
	send := func(messageType int, payload []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		if err := socket.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return err
		}
		return socket.WriteMessage(messageType, payload)
	}
	session, err := service.Container.Terminal(form.Id, form.Height, form.Width)
	if err != nil {
		payload, _ := json.Marshal(map[string]string{"event": "error", "content": err.Error()})
		_ = send(websocket.TextMessage, payload)
		return
	}
	done := make(chan struct{})
	var closeOnce sync.Once
	closeSession := func() {
		closeOnce.Do(func() {
			close(done)
			_ = socket.Close()
			_ = session.Close()
		})
	}
	var workers sync.WaitGroup
	defer func() {
		closeSession()
		workers.Wait()
	}()
	if send(websocket.TextMessage, []byte(`{"event":"ready"}`)) != nil {
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventLogin, "连接容器终端", "连接容器["+form.Id+"]")

	socket.SetReadLimit(1024 * 1024)
	_ = socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	socket.SetPongHandler(func(string) error {
		return socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	ack := make(chan struct{}, 1)
	input := make(chan []byte, 4)
	workers.Add(3)
	go func() {
		defer workers.Done()
		for {
			select {
			case payload := <-input:
				for len(payload) > 0 {
					n, writeErr := session.Write(payload)
					if writeErr != nil || n == 0 {
						closeSession()
						return
					}
					payload = payload[n:]
				}
			case <-done:
				return
			}
		}
	}()
	go func() {
		defer workers.Done()
		defer closeSession()
		buffer := make([]byte, 32*1024)
		for {
			n, readErr := session.Read(buffer)
			if n > 0 {
				if send(websocket.BinaryMessage, buffer[:n]) != nil {
					return
				}
				// 浏览器完成当前帧的终端解析后再读取，持续输出不会堆积在内存中。
				select {
				case <-ack:
				case <-done:
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()
	go func() {
		defer workers.Done()
		ticker := time.NewTicker(54 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if socket.WriteControl(websocket.PingMessage, nil, time.Now().Add(30*time.Second)) != nil {
					closeSession()
					return
				}
			case <-done:
				return
			}
		}
	}()
	for {
		_, payload, readErr := socket.ReadMessage()
		if readErr != nil {
			return
		}
		var data struct {
			Event   string `json:"event"`
			Content string `json:"content"`
		}
		if json.Unmarshal(payload, &data) != nil {
			continue
		}
		switch data.Event {
		case "ack":
			select {
			case ack <- struct{}{}:
			default:
			}
		case "ping":
			err = send(websocket.TextMessage, []byte(`{"event":"pong"}`))
		case "resize":
			if width, height, ok := strings.Cut(data.Content, "|"); ok {
				cols, widthErr := strconv.Atoi(width)
				rows, heightErr := strconv.Atoi(height)
				if widthErr == nil && heightErr == nil && cols >= 1 && cols <= 1000 && rows >= 1 && rows <= 1000 {
					err = session.Resize(rows, cols)
				}
			}
		case "write":
			// PTY 输入可能阻塞，独立写入才能及时响应浏览器断开；队列满时结束会话。
			select {
			case input <- []byte(data.Content):
			default:
				return
			}
		}
		if err != nil {
			return
		}
	}
}
