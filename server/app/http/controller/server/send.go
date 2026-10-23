package server

import (
	"time"

	"github.com/gorilla/websocket"
)

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
