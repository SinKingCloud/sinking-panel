package route

import (
	"path"
	"strings"

	"server/app/http/controller/auth"
	"server/app/http/controller/cert"
	"server/app/http/controller/config"
	"server/app/http/controller/container"
	"server/app/http/controller/file"
	"server/app/http/controller/mongo"
	"server/app/http/controller/mysql"
	"server/app/http/controller/pgsql"
	"server/app/http/controller/recycle"
	"server/app/http/controller/redis"
	"server/app/http/controller/script"
	"server/app/http/controller/secret"
	"server/app/http/controller/server"
	"server/app/http/controller/site"
	"server/app/http/controller/system"
	"server/app/http/controller/task"
	"server/app/http/controller/types"
	"server/app/http/middleware"
	"server/app/util/context"
	"server/app/util/platform"
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
	loadSiteRoute(s)
	loadCertRoute(s)
	loadSecretRoute(s)
	// 支持轻量容器时才注册容器及依赖它的软件接口。
	if platform.CanRunContainer() {
		loadRedisRoute(s)
		loadMysqlRoute(s)
		loadPgsqlRoute(s)
		loadMongoRoute(s)
		loadContainerRoute(s)
	}
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

// loadSiteRoute 网站管理路由
func loadSiteRoute(s *sinking_web.Engine) {
	g := s.Group("/site")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(site.List))               // 网站列表
	g.ANY("/info", context.HandleFunc(site.Info))               // 网站详情
	g.ANY("/create", context.HandleFunc(site.Create))           // 创建网站
	g.ANY("/update", context.HandleFunc(site.Update))           // 修改基础配置
	g.ANY("/delete", context.HandleFunc(site.Delete))           // 删除网站
	g.ANY("/enable", context.HandleFunc(site.Enable))           // 启用网站
	g.ANY("/disable", context.HandleFunc(site.Disable))         // 停用网站
	g.ANY("/log", context.HandleFunc(site.Log))                 // 网站日志
	g.ANY("/domain", context.HandleFunc(site.Domain))           // 域名配置
	g.ANY("/ssl", context.HandleFunc(site.SSL))                 // SSL 配置
	g.ANY("/waf", context.HandleFunc(site.Waf))                 // WAF 配置
	g.ANY("/cache", context.HandleFunc(site.Cache))             // 缓存配置
	g.ANY("/rate", context.HandleFunc(site.RateLimit))          // 访问频率限制
	g.ANY("/traffic", context.HandleFunc(site.TrafficLimit))    // 流量限制
	g.ANY("/header", context.HandleFunc(site.Header))           // 请求和响应头
	g.ANY("/compression", context.HandleFunc(site.Compression)) // 响应压缩
	g.ANY("/redirect", context.HandleFunc(site.Redirect))       // 重定向配置
	g.ANY("/route", context.HandleFunc(site.Route))             // 自定义路由
	g.ANY("/static", context.HandleFunc(site.Static))           // 静态网站配置
	g.ANY("/proxy", context.HandleFunc(site.Proxy))             // 反向代理配置
	g.ANY("/fastcgi", context.HandleFunc(site.FastCGI))         // PHP-FPM 配置
	g.ANY("/process", context.HandleFunc(site.Process))         // 通用网站进程配置
}

// loadCertRoute 证书管理路由
func loadCertRoute(s *sinking_web.Engine) {
	g := s.Group("/cert")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(cert.List))     // 证书列表
	g.ANY("/info", context.HandleFunc(cert.Info))     // 证书详情
	g.ANY("/create", context.HandleFunc(cert.Create)) // 导入证书
	g.ANY("/update", context.HandleFunc(cert.Update)) // 修改证书
	g.ANY("/delete", context.HandleFunc(cert.Delete)) // 删除证书
	g.ANY("/obtain", context.HandleFunc(cert.Obtain)) // 申请证书
	g.ANY("/renew", context.HandleFunc(cert.Renew))   // 续签证书
	g.ANY("/log", context.HandleFunc(cert.Log))       // 自动续签任务日志
}

// loadSecretRoute 密钥管理路由。
func loadSecretRoute(s *sinking_web.Engine) {
	g := s.Group("/secret")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(secret.List))     // 密钥列表
	g.ANY("/info", context.HandleFunc(secret.Info))     // 密钥详情
	g.ANY("/create", context.HandleFunc(secret.Create)) // 添加密钥
	g.ANY("/update", context.HandleFunc(secret.Update)) // 修改密钥
	g.ANY("/delete", context.HandleFunc(secret.Delete)) // 删除密钥
}

// loadRedisRoute Redis 管理路由。
func loadRedisRoute(s *sinking_web.Engine) {
	g := s.Group("/redis")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/image", context.HandleFunc(redis.Image))             // 解析本机镜像
	g.ANY("/install", context.HandleFunc(redis.Install))         // 导入镜像并安装
	g.ANY("/uninstall", context.HandleFunc(redis.Uninstall))     // 卸载并删除实例
	g.ANY("/start", context.HandleFunc(redis.Start))             // 启动服务
	g.ANY("/stop", context.HandleFunc(redis.Stop))               // 停止服务
	g.ANY("/restart", context.HandleFunc(redis.Restart))         // 重启服务
	g.ANY("/reload", context.HandleFunc(redis.Reload))           // 重载配置
	g.ANY("/performance", context.HandleFunc(redis.Performance)) // 性能配置
	g.ANY("/persistence", context.HandleFunc(redis.Persistence)) // 持久化配置
	g.ANY("/config", context.HandleFunc(redis.Config))           // 配置文件
	g.ANY("/status", context.HandleFunc(redis.Status))           // 运行状态
	g.ANY("/log", context.HandleFunc(redis.Log))                 // 运行日志
}

// loadMysqlRoute MySQL 安装、调优和数据库管理。
func loadMysqlRoute(s *sinking_web.Engine) {
	g := s.Group("/mysql")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.Use(context.HandleFunc(func(c *context.Context) { c.SetHeader("Cache-Control", "no-store") }))
	g.ANY("/image", context.HandleFunc(mysql.Image))             // 解析镜像
	g.ANY("/install", context.HandleFunc(mysql.Install))         // 安装服务
	g.ANY("/uninstall", context.HandleFunc(mysql.Uninstall))     // 卸载服务
	g.ANY("/start", context.HandleFunc(mysql.Start))             // 启动
	g.ANY("/stop", context.HandleFunc(mysql.Stop))               // 停止
	g.ANY("/restart", context.HandleFunc(mysql.Restart))         // 重启
	g.ANY("/status", context.HandleFunc(mysql.Status))           // 状态与负载
	g.ANY("/performance", context.HandleFunc(mysql.Performance)) // 可视化调优
	g.ANY("/config", context.HandleFunc(mysql.Config))           // 配置文件
	g.ANY("/reload", context.HandleFunc(mysql.Reload))           // 重载配置
	g.ANY("/log", context.HandleFunc(mysql.Log))                 // 运行日志
	g.ANY("/databases", context.HandleFunc(mysql.Databases))     // 数据库管理
	g.ANY("/users", context.HandleFunc(mysql.Users))             // 账号管理
}

// loadPgsqlRoute PostgreSQL 安装、调优和数据库管理。
func loadPgsqlRoute(s *sinking_web.Engine) {
	g := s.Group("/pgsql")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.Use(context.HandleFunc(func(c *context.Context) { c.SetHeader("Cache-Control", "no-store") }))
	g.ANY("/image", context.HandleFunc(pgsql.Image))             // 解析镜像
	g.ANY("/install", context.HandleFunc(pgsql.Install))         // 安装服务
	g.ANY("/uninstall", context.HandleFunc(pgsql.Uninstall))     // 卸载服务
	g.ANY("/start", context.HandleFunc(pgsql.Start))             // 启动
	g.ANY("/stop", context.HandleFunc(pgsql.Stop))               // 停止
	g.ANY("/restart", context.HandleFunc(pgsql.Restart))         // 重启
	g.ANY("/status", context.HandleFunc(pgsql.Status))           // 状态与负载
	g.ANY("/performance", context.HandleFunc(pgsql.Performance)) // 可视化调优
	g.ANY("/config", context.HandleFunc(pgsql.Config))           // 配置文件
	g.ANY("/reload", context.HandleFunc(pgsql.Reload))           // 重载配置
	g.ANY("/log", context.HandleFunc(pgsql.Log))                 // 运行日志
	g.ANY("/databases", context.HandleFunc(pgsql.Databases))     // 数据库管理
	g.ANY("/users", context.HandleFunc(pgsql.Users))             // 账号管理
}

// loadMongoRoute MongoDB 安装、调优和数据库管理。
func loadMongoRoute(s *sinking_web.Engine) {
	g := s.Group("/mongo")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.Use(context.HandleFunc(func(c *context.Context) { c.SetHeader("Cache-Control", "no-store") }))
	g.ANY("/image", context.HandleFunc(mongo.Image))             // 解析镜像
	g.ANY("/install", context.HandleFunc(mongo.Install))         // 安装服务
	g.ANY("/uninstall", context.HandleFunc(mongo.Uninstall))     // 卸载服务
	g.ANY("/start", context.HandleFunc(mongo.Start))             // 启动
	g.ANY("/stop", context.HandleFunc(mongo.Stop))               // 停止
	g.ANY("/restart", context.HandleFunc(mongo.Restart))         // 重启
	g.ANY("/status", context.HandleFunc(mongo.Status))           // 状态与负载
	g.ANY("/performance", context.HandleFunc(mongo.Performance)) // 可视化调优
	g.ANY("/config", context.HandleFunc(mongo.Config))           // 配置文件
	g.ANY("/reload", context.HandleFunc(mongo.Reload))           // 重载配置
	g.ANY("/log", context.HandleFunc(mongo.Log))                 // 运行日志
	g.ANY("/databases", context.HandleFunc(mongo.Databases))     // 数据库管理
	g.ANY("/users", context.HandleFunc(mongo.Users))             // 账号管理
}

// loadContainerRoute 容器通用操作。
func loadContainerRoute(s *sinking_web.Engine) {
	g := s.Group("/container")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(container.List))         // 容器列表
	g.ANY("/info", context.HandleFunc(container.Info))         // 容器详情
	g.ANY("/status", context.HandleFunc(container.Status))     // 动态状态
	g.ANY("/run", context.HandleFunc(container.Run))           // 创建并启动容器
	g.ANY("/start", context.HandleFunc(container.Start))       // 启动容器
	g.ANY("/stop", context.HandleFunc(container.Stop))         // 停止容器
	g.ANY("/restart", context.HandleFunc(container.Restart))   // 重启容器
	g.ANY("/remove", context.HandleFunc(container.Remove))     // 删除容器
	g.ANY("/images", context.HandleFunc(container.Images))     // 镜像列表、详情、导入和删除
	g.ANY("/exec", context.HandleFunc(container.Exec))         // 执行容器命令
	g.ANY("/log", context.HandleFunc(container.Log))           // 读取或清理容器日志
	g.ANY("/terminal", context.HandleFunc(container.Terminal)) // 交互终端
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
	g.ANY("/list", context.HandleFunc(script.List))     //脚本列表
	g.ANY("/info", context.HandleFunc(script.Info))     //脚本详情
	g.ANY("/create", context.HandleFunc(script.Create)) //添加脚本
	g.ANY("/update", context.HandleFunc(script.Update)) //更新脚本
	g.ANY("/delete", context.HandleFunc(script.Delete)) //删除脚本
}

// loadTypeRoute 类型路由
func loadTypeRoute(s *sinking_web.Engine) {
	g := s.Group("/type")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(types.List))     //类型列表
	g.ANY("/create", context.HandleFunc(types.Create)) //添加类型
	g.ANY("/update", context.HandleFunc(types.Update)) //更新类型
	g.ANY("/delete", context.HandleFunc(types.Delete)) //删除类型
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
	g.ANY("/sign", context.HandleFunc(file.Sign))         //获取预览签名
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
	g.ANY("/account", context.HandleFunc(system.Account))   //账户管理
	g.ANY("/info", context.HandleFunc(system.Info))         //主机、系统和运行时信息
	g.ANY("/timezone", context.HandleFunc(system.Timezone)) //系统时区
	g.ANY("/statis", context.HandleFunc(system.Statis))     //系统资源统计
	g.ANY("/io", context.HandleFunc(system.IO))             //网络与磁盘 I/O 统计
	g.ANY("/task", context.HandleFunc(system.Task))         //系统任务
	g.ANY("/log", context.HandleFunc(system.Log))           //系统日志
	g.ANY("/enum", context.HandleFunc(system.Enum))         //枚举类型
	g.ANY("/http", context.HandleFunc(system.HTTP))         //HTTP服务管理
	g.ANY("/restart", context.HandleFunc(system.Restart))   //重启面板或服务器
}
