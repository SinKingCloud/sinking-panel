package server

type message struct {
	Event   string `json:"event"`
	Content string `json:"content,omitempty"`
	Code    string `json:"code,omitempty"`
}
