package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/http/controller/auth"
	"server/app/http/middleware"
	"server/app/util/server"
	"server/public"
)

func loadApp(s *sinking_web.Engine) {
	s.Use(middleware.Cors())
	loadAuthRoute(s)
	loadSystemRoute(s)
	loadStaticRoute(s)
}

func loadStaticRoute(s *sinking_web.Engine) {
	s.ANY("/", server.HandleFunc(func(c *server.Context) {
		c.SetHeader("content-type", "text/html;charset=utf-8;")
		c.Data(200, public.Index())
	}))
	s.ANY("/*", server.HandleFunc(func(c *server.Context) {
		c.Request.URL.Path = public.Path() + c.Request.URL.Path
		public.FileServer.ServeHTTP(c.Writer, c.Request)
	}))
}

func loadAuthRoute(s *sinking_web.Engine) {
	g := s.Group("/auth")
	web := g.Group("/web")
	web.ANY("/info", server.HandleFunc(auth.Web.Info))
	web.ANY("/login", server.HandleFunc(auth.Web.Login))
	out := web.Group("/out_login")
	out.Use(middleware.CheckLogin())
	out.ANY("/", server.HandleFunc(auth.Web.OutLogin))
}

func loadSystemRoute(s *sinking_web.Engine) {
	g := s.Group("/admin")
	g.Use(middleware.CheckLogin())
	file := g.Group("/file")
	{
		file.ANY("/test", server.HandleFunc(func(c *server.Context) {
			c.Success("success")
		}))
	}
}
