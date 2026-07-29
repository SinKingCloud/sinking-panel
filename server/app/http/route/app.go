package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/http/controller/auth"
	"server/app/http/controller/config"
	"server/app/http/controller/file"
	"server/app/http/controller/recycle"
	server2 "server/app/http/controller/server"
	"server/app/http/controller/system"
	"server/app/http/controller/task"
	"server/app/http/middleware"
	"server/app/util/context"
	"server/public"
)

func loadApp(s *sinking_web.Engine) {
	loadMiddleware(s)
	loadAuthRoute(s)
	loadFileRoute(s)
	loadRecycleRoute(s)
	loadServerRoute(s)
	loadConfigRoute(s)
	loadSystemRoute(s)
	loadTaskRoute(s)
	loadStaticRoute(s)
}

// loadMiddleware 中间件
func loadMiddleware(s *sinking_web.Engine) {
	s.Use(context.HandleFunc(middleware.Cors))
}

// loadStaticRoute 静态资源
func loadStaticRoute(s *sinking_web.Engine) {
	s.ANY("/", context.HandleFunc(func(c *context.Context) {
		c.SetHeader("content-type", "text/html;charset=utf-8;")
		c.Data(200, public.ReadDistFile("index.html"))
	}))
	s.ANY("/*", context.HandleFunc(func(c *context.Context) {
		c.Request.URL.Path = public.Path() + c.Request.URL.Path
		public.FileServer.ServeHTTP(c.Writer, c.Request)
	}))
}

// loadAuthRoute 授权路由
func loadAuthRoute(s *sinking_web.Engine) {
	s.ANY("/info", context.HandleFunc(auth.Info))       //网站信息
	s.ANY("/login", context.HandleFunc(auth.Login))     //账号登录
	s.ANY("/logout", context.HandleFunc(auth.Logout))   //注销登录
	s.ANY("/captcha", context.HandleFunc(auth.Captcha)) //验证码
}

// loadConfigRoute 配置路由
func loadConfigRoute(s *sinking_web.Engine) {
	g := s.Group("/config")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/get", context.HandleFunc(config.Get)) //获取配置
	g.ANY("/set", context.HandleFunc(config.Set)) //修改配置
}

// loadServerRoute 服务器路由
func loadServerRoute(s *sinking_web.Engine) {
	g := s.Group("/server")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/ssh", context.HandleFunc(server2.Ssh))       //连接ssh
	g.ANY("/list", context.HandleFunc(server2.List))     //服务器列表
	g.ANY("/info", context.HandleFunc(server2.Info))     //服务器信息
	g.ANY("/create", context.HandleFunc(server2.Create)) //添加服务器
	g.ANY("/update", context.HandleFunc(server2.Update)) //更新服务器
	g.ANY("/delete", context.HandleFunc(server2.Delete)) //删除服务器
}

// loadTaskRoute 任务路由
func loadTaskRoute(s *sinking_web.Engine) {
	g := s.Group("/task")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(task.List))       //任务列表
	g.ANY("/info", context.HandleFunc(task.Info))       //任务详情
	g.ANY("/create", context.HandleFunc(task.Create))   //创建任务
	g.ANY("/delete", context.HandleFunc(task.Delete))   //删除任务
	g.ANY("/update", context.HandleFunc(task.Update))   //更新任务
	g.ANY("/run", context.HandleFunc(task.Run))         //执行任务
	g.ANY("/stop", context.HandleFunc(task.Stop))       //暂停任务
	g.ANY("/restore", context.HandleFunc(task.Restore)) //恢复任务
	g.ANY("/log", context.HandleFunc(task.Log))         //任务日志
}

// loadRecycleRoute 回收站路由
func loadRecycleRoute(s *sinking_web.Engine) {
	g := s.Group("/recycle")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(recycle.List))       //回收站列表
	g.ANY("/count", context.HandleFunc(recycle.Count))     //数据统计
	g.ANY("/clear", context.HandleFunc(recycle.Clear))     //清空回收站
	g.ANY("/create", context.HandleFunc(recycle.Create))   //添加文件
	g.ANY("/delete", context.HandleFunc(recycle.Delete))   //删除文件
	g.ANY("/restore", context.HandleFunc(recycle.Restore)) //恢复文件
}

// loadFileRoute 文件路由
func loadFileRoute(s *sinking_web.Engine) {
	g := s.Group("/file")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/disk", context.HandleFunc(file.Disk))         //分区信息
	g.ANY("/list", context.HandleFunc(file.List))         //文件列表
	g.ANY("/info", context.HandleFunc(file.Info))         //文件信息
	g.ANY("/create", context.HandleFunc(file.Create))     //创建文件
	g.ANY("/count", context.HandleFunc(file.Count))       //统计信息
	g.ANY("/delete", context.HandleFunc(file.Delete))     //删除文件
	g.ANY("/update", context.HandleFunc(file.Update))     //更新信息
	g.ANY("/preview", context.HandleFunc(file.Preview))   //预览文件
	g.ANY("/copy", context.HandleFunc(file.Copy))         //复制文件
	g.ANY("/move", context.HandleFunc(file.Move))         //移动文件
	g.ANY("/extract", context.HandleFunc(file.Extract))   //解压文件
	g.ANY("/compress", context.HandleFunc(file.Compress)) //压缩文件
	g.ANY("/upload", context.HandleFunc(file.Upload))     //上传文件
	g.ANY("/download", context.HandleFunc(file.Download)) //下载文件
}

// loadSystemRoute 系统路由
func loadSystemRoute(s *sinking_web.Engine) {
	g := s.Group("/system")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/info", context.HandleFunc(system.Info))     //系统信息
	g.ANY("/status", context.HandleFunc(system.Status)) //系统状态
	g.ANY("/task", context.HandleFunc(system.Task))     //系统任务
	g.ANY("/log", context.HandleFunc(system.Log))       //系统日志
	g.ANY("/enum", context.HandleFunc(system.Enum))     //枚举类型
}
