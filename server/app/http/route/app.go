package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/http/controller/auth"
	"server/app/http/controller/socket"
	"server/app/http/middleware"
	"server/app/util"
	"server/app/util/server"
	"server/public"
)

func loadApp(s *sinking_web.Engine) {
	if util.IsDebug() {
		s.Use(server.HandleFunc(middleware.Cors))
	}
	loadAuthRoute(s)
	loadFileRoute(s)
	loadStaticRoute(s)
}

func loadStaticRoute(s *sinking_web.Engine) {
	s.ANY("/", server.HandleFunc(func(c *server.Context) {
		c.SetHeader("content-type", "text/html;charset=utf-8;")
		c.Data(200, public.ReadDistFile("index.html"))
	}))
	s.ANY("/*", server.HandleFunc(func(c *server.Context) {
		c.Request.URL.Path = public.Path() + c.Request.URL.Path
		public.FileServer.ServeHTTP(c.Writer, c.Request)
	}))
}

func loadAuthRoute(s *sinking_web.Engine) {
	s.ANY("/login", server.HandleFunc(auth.Login))
	s.ANY("/logout", server.HandleFunc(auth.Logout))
	s.ANY("/captcha", server.HandleFunc(auth.Captcha))
	s.ANY("/socket", server.HandleFunc(socket.Server))
}

func loadFileRoute(s *sinking_web.Engine) {
	g := s.Group("/file")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", server.HandleFunc(func(c *server.Context) {
		c.Success("success")
	}))
}
