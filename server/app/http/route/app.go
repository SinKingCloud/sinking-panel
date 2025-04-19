package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/http/controller/auth"
	"server/app/http/controller/config"
	"server/app/http/controller/cron"
	"server/app/http/controller/file"
	"server/app/http/controller/recycle"
	server2 "server/app/http/controller/server"
	"server/app/http/controller/system"
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
	loadRecycleRoute(s)
	loadServerRoute(s)
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
	s.ANY("/login", server.HandleFunc(auth.Login))     //账号登录
	s.ANY("/logout", server.HandleFunc(auth.Logout))   //注销登录
	s.ANY("/captcha", server.HandleFunc(auth.Captcha)) //验证码
}

func loadConfigRoute(s *sinking_web.Engine) {
	g := s.Group("/config")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/get", server.HandleFunc(config.Get)) //获取配置
	g.ANY("/set", server.HandleFunc(config.Set)) //修改配置
}

func loadServerRoute(s *sinking_web.Engine) {
	g := s.Group("/server")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/ssh", server.HandleFunc(server2.Ssh))       //连接ssh
	g.ANY("/list", server.HandleFunc(server2.List))     //服务器列表
	g.ANY("/info", server.HandleFunc(server2.Info))     //服务器信息
	g.ANY("/create", server.HandleFunc(server2.Create)) //添加服务器
	g.ANY("/update", server.HandleFunc(server2.Update)) //更新服务器
	g.ANY("/delete", server.HandleFunc(server2.Delete)) //删除服务器
}

func loadCronRoute(s *sinking_web.Engine) {
	g := s.Group("/cron")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", server.HandleFunc(cron.List))       //任务列表
	g.ANY("/info", server.HandleFunc(cron.Info))       //任务详情
	g.ANY("/create", server.HandleFunc(cron.Create))   //创建任务
	g.ANY("/delete", server.HandleFunc(cron.Delete))   //删除任务
	g.ANY("/update", server.HandleFunc(cron.Update))   //更新任务
	g.ANY("/run", server.HandleFunc(cron.Run))         //执行任务
	g.ANY("/stop", server.HandleFunc(cron.Stop))       //暂停任务
	g.ANY("/restore", server.HandleFunc(cron.Restore)) //恢复任务
	g.ANY("/log", server.HandleFunc(cron.Log))         //任务日志
}

func loadRecycleRoute(s *sinking_web.Engine) {
	g := s.Group("/recycle")
	g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", server.HandleFunc(recycle.List))       //回收站列表
	g.ANY("/count", server.HandleFunc(recycle.Count))     //数据统计
	g.ANY("/clear", server.HandleFunc(recycle.Clear))     //清空回收站
	g.ANY("/create", server.HandleFunc(recycle.Create))   //添加文件
	g.ANY("/delete", server.HandleFunc(recycle.Delete))   //删除文件
	g.ANY("/restore", server.HandleFunc(recycle.Restore)) //恢复文件
}

func loadFileRoute(s *sinking_web.Engine) {
	g := s.Group("/file")
	//g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/disk", server.HandleFunc(file.Disk))         //分区信息
	g.ANY("/list", server.HandleFunc(file.List))         //文件列表
	g.ANY("/info", server.HandleFunc(file.Info))         //文件信息
	g.ANY("/create", server.HandleFunc(file.Create))     //创建文件
	g.ANY("/count", server.HandleFunc(file.Count))       //统计信息
	g.ANY("/delete", server.HandleFunc(file.Delete))     //删除文件
	g.ANY("/update", server.HandleFunc(file.Update))     //更新信息
	g.ANY("/preview", server.HandleFunc(file.Preview))   //预览文件
	g.ANY("/copy", server.HandleFunc(file.Copy))         //复制文件
	g.ANY("/move", server.HandleFunc(file.Move))         //移动文件
	g.ANY("/extract", server.HandleFunc(file.Extract))   //解压文件
	g.ANY("/compress", server.HandleFunc(file.Compress)) //压缩文件
	g.ANY("/upload", nil)                                //上传文件
	g.ANY("/download", server.HandleFunc(file.Download)) //下载文件
}

func loadSystemRoute(s *sinking_web.Engine) {
	g := s.Group("/system")
	//g.Use(server.HandleFunc(middleware.CheckLogin))
	g.ANY("/info", nil)                            //系统信息
	g.ANY("/status", nil)                          //系统状态
	g.ANY("/task", server.HandleFunc(system.Task)) //系统任务
	g.ANY("/log", server.HandleFunc(system.Log))   //系统日志
	g.ANY("/enum", server.HandleFunc(system.Enum)) //枚举类型
}
