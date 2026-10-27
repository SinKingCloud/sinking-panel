//go:build !linux || !cgo

package command

// executeInit 在不支持 libcontainer init 的平台上返回 false，继续走正常命令解析。
func (s *Server) executeInit(args []string) bool {
	return false
}
