package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/http/controller/auth"
	"server/app/http/controller/ssh"
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
	loadSshRoute(s)
	loadConfigRoute(s)
	loadSystemRoute(s)
	loadCronRoute(s)
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
}

func loadSystemRoute(s *sinking_web.Engine) {
	g := s.Group("/system")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/version", nil)
	g.ANY("/status", nil)
	g.ANY("/log", nil)
}

func loadCronRoute(s *sinking_web.Engine) {
	g := s.Group("/cron")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", nil)
	g.ANY("/create", nil)
	g.ANY("/update", nil)
	g.ANY("/delete", nil)
	g.ANY("/info", nil)
	g.ANY("/run", nil)
	g.ANY("/log", nil)
}

func loadConfigRoute(s *sinking_web.Engine) {
	g := s.Group("/config")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/get", nil)
	g.ANY("/set", nil)
}

func loadSshRoute(s *sinking_web.Engine) {
	g := s.Group("/ssh")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/socket", server.HandleFunc(ssh.Socket))
	g.ANY("/list", nil)
	g.ANY("/info", nil)
	g.ANY("/create", nil)
	g.ANY("/update", nil)
	g.ANY("/delete", nil)
}

func loadFileRoute(s *sinking_web.Engine) {
	g := s.Group("/file")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/disk", nil)
	g.ANY("/list", nil)
	g.ANY("/info", nil)
	g.ANY("/create", nil)
	g.ANY("/update", nil)
	g.ANY("/delete", nil)
	g.ANY("/copy", nil)
	g.ANY("/cut", nil)
	g.ANY("/download", nil)
}
