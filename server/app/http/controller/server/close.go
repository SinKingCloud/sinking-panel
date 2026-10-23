package server

import (
	"time"

	"github.com/gorilla/websocket"
)

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
