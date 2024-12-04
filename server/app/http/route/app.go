package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/http/controller/admin"
	"server/app/http/controller/auth"
	"server/app/http/controller/user"
	"server/app/http/middleware"
	"server/app/util/server"
	"server/public"
)

func loadApp(s *sinking_web.Engine) {
	s.Use(middleware.Cors())
	loadAuthRoute(s)
	loadUserRoute(s)
	loadAdminRoute(s)
	loadStaticRoute(s)
}

func loadStaticRoute(s *sinking_web.Engine) {
	s.ANY("/", server.HandleFunc(func(c *server.Context) {
		c.SetHeader("content-type", "text/html;charset=utf-8;")
		c.Data(200, public.Index())
	}))
	s.ANY("/captcha.js", server.HandleFunc(func(c *server.Context) {
		c.SetHeader("content-type", "application/javascript;charset=utf-8;")
		c.Data(200, public.Captcha())
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

	verify := g.Group("/verify")
	verify.ANY("/code", server.HandleFunc(auth.Verify.Code))
	verify.ANY("/sms", server.HandleFunc(auth.Verify.Sms))

	callback := g.Group("/callback")
	callback.ANY("/epay/:method", server.HandleFunc(auth.Callback.Epay))

	pay := g.Group("/pay")
	pay.ANY("/submit", server.HandleFunc(auth.Pay.Submit))
}

func loadUserRoute(s *sinking_web.Engine) {
	g := s.Group("/user")
	g.Use(middleware.CheckLogin())

	person := g.Group("/person")
	{
		person.ANY("/info", server.HandleFunc(user.Person.Info))
		person.ANY("/password", server.HandleFunc(user.Person.Password))
		person.ANY("/phone", server.HandleFunc(user.Person.Phone))
		person.ANY("/log", server.HandleFunc(user.Person.Log))
		person.ANY("/token", server.HandleFunc(user.Person.GenToken))
	}

	notice := g.Group("/notice")
	{
		notice.ANY("/list", server.HandleFunc(user.Notice.List))
		notice.ANY("/info", server.HandleFunc(user.Notice.Info))
	}

	pay := g.Group("/pay")
	{
		pay.ANY("/log", server.HandleFunc(user.Pay.Log))
		pay.ANY("/order", server.HandleFunc(user.Pay.Order))
		pay.ANY("/recharge", server.HandleFunc(user.Pay.Recharge))
	}

	uin := g.Group("/uin")
	{
		uin.ANY("/price", server.HandleFunc(user.Uin.Price))
		uin.ANY("/server", server.HandleFunc(user.Uin.Server))
		uin.ANY("/project", server.HandleFunc(user.Uin.Project))
		uin.ANY("/log", server.HandleFunc(user.Uin.Log))
		uin.ANY("/order", server.HandleFunc(user.Uin.Order))
		uin.ANY("/refund", server.HandleFunc(user.Uin.Refund))
		uin.ANY("/info", server.HandleFunc(user.Uin.Info))
		uin.ANY("/login", server.HandleFunc(user.Uin.Login))
		uin.ANY("/notify", server.HandleFunc(user.Uin.Notify))
		uin.ANY("/sms", server.HandleFunc(user.Uin.Sms))
		uin.ANY("/run", server.HandleFunc(user.Uin.Run))
		uin.ANY("/buy", server.HandleFunc(user.Uin.Buy))
		uin.ANY("/list", server.HandleFunc(user.Uin.List))
		uin.ANY("/delete", server.HandleFunc(user.Uin.Delete))
		uin.ANY("/edit_device", server.HandleFunc(user.Uin.ReSetDevice))
		uin.ANY("/edit_server", server.HandleFunc(user.Uin.EditServer))
		uin.ANY("/edit_timing", server.HandleFunc(user.Uin.EditTiming))
		uin.ANY("/edit_password", server.HandleFunc(user.Uin.EditPassword))
		uin.ANY("/edit_status", server.HandleFunc(user.Uin.EditStatus))
		uin.ANY("/edit_setting", server.HandleFunc(user.Uin.EditSetting))
		uin.ANY("/level", server.HandleFunc(user.Uin.LevelInfo))
	}
}

func loadAdminRoute(s *sinking_web.Engine) {
	g := s.Group("/admin")
	g.Use(middleware.CheckLogin(), middleware.CheckAdmin())

	users := g.Group("/user")
	{
		users.ANY("/list", server.HandleFunc(admin.User.List))
		users.ANY("/update", server.HandleFunc(admin.User.Update))
		users.ANY("/log", server.HandleFunc(admin.User.Log))
		users.ANY("/money", server.HandleFunc(admin.User.Money))
		users.ANY("/config/:method", server.HandleFunc(admin.User.Config))
	}

	notice := g.Group("/notice")
	{
		notice.ANY("/list", server.HandleFunc(admin.Notice.List))
		notice.ANY("/update", server.HandleFunc(admin.Notice.Update))
		notice.ANY("/create", server.HandleFunc(admin.Notice.Create))
		notice.ANY("/delete", server.HandleFunc(admin.Notice.Delete))
	}

	pay := g.Group("/pay")
	{
		pay.ANY("/log", server.HandleFunc(admin.Pay.Log))
		pay.ANY("/order", server.HandleFunc(admin.Pay.Order))
	}

	config := g.Group("/config")
	{
		config.ANY("/list", server.HandleFunc(admin.Config.List))
		config.ANY("/update", server.HandleFunc(admin.Config.Update))
	}

	srv := g.Group("/server")
	{
		srv.ANY("/list", server.HandleFunc(admin.Server.List))
		srv.ANY("/update", server.HandleFunc(admin.Server.Update))
		srv.ANY("/create", server.HandleFunc(admin.Server.Create))
		srv.ANY("/delete", server.HandleFunc(admin.Server.Delete))
	}

	uin := g.Group("/uin")
	{
		uin.ANY("/list", server.HandleFunc(admin.Uin.List))
		uin.ANY("/buy", server.HandleFunc(admin.Uin.Buy))
	}
}
