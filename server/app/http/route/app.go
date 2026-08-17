package route

import (
	"path"
	"strings"

	"server/app/http/controller/auth"
	"server/app/http/controller/config"
	"server/app/http/controller/file"
	"server/app/http/controller/recycle"
	scriptController "server/app/http/controller/script"
	"server/app/http/controller/server"
	"server/app/http/controller/system"
	"server/app/http/controller/task"
	typeController "server/app/http/controller/types"
	"server/app/http/middleware"
	"server/app/util/context"
	"server/public"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

func loadApp(s *sinking_web.Engine) {
	loadMiddleware(s)
	loadAuthRoute(s)
	loadFileRoute(s)
	loadRecycleRoute(s)
	loadServerRoute(s)
	loadScriptRoute(s)
	loadConfigRoute(s)
	loadSystemRoute(s)
	loadTaskRoute(s)
	loadTypeRoute(s)
	loadStaticRoute(s)
}

// loadMiddleware 中间件
func loadMiddleware(s *sinking_web.Engine) {
	s.Use(context.HandleFunc(middleware.Cors))
}

// loadStaticRoute 静态资源
func loadStaticRoute(s *sinking_web.Engine) {
	s.ANY("/", context.HandleFunc(func(c *context.Context) {
		c.SetHeader("Cache-Control", "no-cache, must-revalidate")
		c.SetHeader("content-type", "text/html;charset=utf-8;")
		c.Data(200, public.ReadDistFile("index.html"))
	}))
	s.ANY("/*", context.HandleFunc(func(c *context.Context) {
		switch strings.ToLower(path.Ext(c.Request.URL.Path)) {
		case ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".svg", ".ico", ".bmp", ".tif", ".tiff":
			c.SetHeader("Cache-Control", "public, max-age=86400")
		case "", ".html", ".htm":
			c.SetHeader("Cache-Control", "no-cache, must-revalidate")
		}
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
	s.ANY("/preview", context.HandleFunc(auth.Preview)) //预览文件
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
	g.ANY("/ssh", context.HandleFunc(server.Ssh))       //连接ssh
	g.ANY("/list", context.HandleFunc(server.List))     //服务器列表
	g.ANY("/info", context.HandleFunc(server.Info))     //服务器信息
	g.ANY("/create", context.HandleFunc(server.Create)) //添加服务器
	g.ANY("/update", context.HandleFunc(server.Update)) //更新服务器
	g.ANY("/delete", context.HandleFunc(server.Delete)) //删除服务器
}

// loadScriptRoute 常用脚本路由
func loadScriptRoute(s *sinking_web.Engine) {
	g := s.Group("/script")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(scriptController.List))     //脚本列表
	g.ANY("/info", context.HandleFunc(scriptController.Info))     //脚本详情
	g.ANY("/create", context.HandleFunc(scriptController.Create)) //添加脚本
	g.ANY("/update", context.HandleFunc(scriptController.Update)) //更新脚本
	g.ANY("/delete", context.HandleFunc(scriptController.Delete)) //删除脚本
}

// loadTypeRoute 类型路由
func loadTypeRoute(s *sinking_web.Engine) {
	g := s.Group("/type")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(typeController.List))     //类型列表
	g.ANY("/create", context.HandleFunc(typeController.Create)) //添加类型
	g.ANY("/update", context.HandleFunc(typeController.Update)) //更新类型
	g.ANY("/delete", context.HandleFunc(typeController.Delete)) //删除类型
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
	g.ANY("/delete", context.HandleFunc(recycle.Delete))   //批量删除文件
	g.ANY("/restore", context.HandleFunc(recycle.Restore)) //批量恢复文件
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
	g.GET("/sign", context.HandleFunc(file.Sign))         //获取预览签名
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
	g.ANY("/account", context.HandleFunc(system.Account)) //账户管理
	g.ANY("/info", context.HandleFunc(system.Info))       //系统信息
	g.ANY("/status", context.HandleFunc(system.Status))   //系统状态
	g.ANY("/task", context.HandleFunc(system.Task))       //系统任务
	g.ANY("/log", context.HandleFunc(system.Log))         //系统日志
	g.ANY("/enum", context.HandleFunc(system.Enum))       //枚举类型
}
