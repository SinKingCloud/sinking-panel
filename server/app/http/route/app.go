package route

import (
	"path"
	"server/app/util/platform"
	"strings"

	"server/app/http/controller/auth"
	"server/app/http/controller/cert"
	"server/app/http/controller/config"
	"server/app/http/controller/container"
	"server/app/http/controller/database"
	"server/app/http/controller/file"
	"server/app/http/controller/recycle"
	"server/app/http/controller/script"
	"server/app/http/controller/secret"
	"server/app/http/controller/server"
	"server/app/http/controller/site"
	"server/app/http/controller/software"
	"server/app/http/controller/system"
	"server/app/http/controller/task"
	"server/app/http/controller/types"
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
	loadSiteRoute(s)
	loadCertRoute(s)
	loadSecretRoute(s)
	loadDatabaseRoute(s)
	loadSoftwareRoute(s)
	loadContainerRoute(s)
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

// loadDatabaseRoute 数据库实例与远端数据库管理，不依赖容器运行环境。
func loadDatabaseRoute(s *sinking_web.Engine) {
	g := s.Group("/database")
	g.Use(context.HandleFunc(middleware.CheckLogin))

	g.ANY("/list", context.HandleFunc(database.List))     // 实例列表
	g.ANY("/info", context.HandleFunc(database.Info))     // 实例详情
	g.ANY("/create", context.HandleFunc(database.Create)) // 添加实例
	g.ANY("/update", context.HandleFunc(database.Update)) // 修改实例
	g.ANY("/delete", context.HandleFunc(database.Delete)) // 删除实例
	g.ANY("/test", context.HandleFunc(database.Test))     // 测试连接
	g.ANY("/config", context.HandleFunc(database.Config)) // 连接配置表单

	mysql := g.Group("/mysql")
	{
		mysql.ANY("/databases", context.HandleFunc(database.Mysql.Databases))                 // 实时数据库列表
		mysql.ANY("/database_info", context.HandleFunc(database.Mysql.DatabaseInfo))          // 真实数据库信息
		mysql.ANY("/charsets", context.HandleFunc(database.Mysql.Charsets))                   // 字符集和排序规则
		mysql.ANY("/create_database", context.HandleFunc(database.Mysql.CreateDatabase))      // 创建真实数据库
		mysql.ANY("/drop_database", context.HandleFunc(database.Mysql.DropDatabase))          // 删除真实数据库
		mysql.ANY("/update_database", context.HandleFunc(database.Mysql.UpdateDatabase))      // 修改数据库属性
		mysql.ANY("/version", context.HandleFunc(database.Mysql.Version))                     // 版本及可用功能
		mysql.ANY("/create_table", context.HandleFunc(database.Mysql.CreateTable))            // 创建数据表
		mysql.ANY("/rename_table", context.HandleFunc(database.Mysql.RenameTable))            // 重命名数据表
		mysql.ANY("/add_column", context.HandleFunc(database.Mysql.AddColumn))                // 添加字段
		mysql.ANY("/update_column", context.HandleFunc(database.Mysql.UpdateColumn))          // 修改字段
		mysql.ANY("/drop_column", context.HandleFunc(database.Mysql.DropColumn))              // 删除字段
		mysql.ANY("/create_index", context.HandleFunc(database.Mysql.CreateIndex))            // 创建索引
		mysql.ANY("/update_index", context.HandleFunc(database.Mysql.UpdateIndex))            // 修改索引备注
		mysql.ANY("/drop_index", context.HandleFunc(database.Mysql.DropIndex))                // 删除索引
		mysql.ANY("/ddl", context.HandleFunc(database.Mysql.DDL))                             // 查看建表语句
		mysql.ANY("/import", context.HandleFunc(database.Mysql.Import))                       // 导入数据
		mysql.ANY("/export", context.HandleFunc(database.Mysql.Export))                       // 导出数据到服务器
		mysql.ANY("/objects", context.HandleFunc(database.Mysql.Objects))                     // 数据库对象
		mysql.ANY("/object_definition", context.HandleFunc(database.Mysql.ObjectDefinition))  // 查看对象定义
		mysql.ANY("/save_object", context.HandleFunc(database.Mysql.SaveObject))              // 保存数据库对象
		mysql.ANY("/drop_object", context.HandleFunc(database.Mysql.DropObject))              // 删除数据库对象
		mysql.ANY("/create_foreign_key", context.HandleFunc(database.Mysql.CreateForeignKey)) // 创建外键
		mysql.ANY("/update_foreign_key", context.HandleFunc(database.Mysql.UpdateForeignKey)) // 修改外键
		mysql.ANY("/drop_foreign_key", context.HandleFunc(database.Mysql.DropForeignKey))     // 删除外键
		mysql.ANY("/tables", context.HandleFunc(database.Mysql.Tables))                       // 数据表列表
		mysql.ANY("/columns", context.HandleFunc(database.Mysql.Columns))                     // 字段信息
		mysql.ANY("/indexes", context.HandleFunc(database.Mysql.Indexes))                     // 索引信息
		mysql.ANY("/rows", context.HandleFunc(database.Mysql.Rows))                           // 分页数据
		mysql.ANY("/execute", context.HandleFunc(database.Mysql.Execute))                     // 执行语句
		mysql.ANY("/insert", context.HandleFunc(database.Mysql.Insert))                       // 添加数据
		mysql.ANY("/update", context.HandleFunc(database.Mysql.Update))                       // 修改数据
		mysql.ANY("/delete", context.HandleFunc(database.Mysql.Delete))                       // 删除数据
		mysql.ANY("/users", context.HandleFunc(database.Mysql.Users))                         // 账户权限
		mysql.ANY("/maintenance", context.HandleFunc(database.Mysql.Maintenance))             // 维护工具
	}
	postgres := g.Group("/postgres")
	{
		postgres.ANY("/databases", context.HandleFunc(database.Postgres.Databases))                 // 实时数据库列表
		postgres.ANY("/database_info", context.HandleFunc(database.Postgres.DatabaseInfo))          // 真实数据库信息
		postgres.ANY("/create_database", context.HandleFunc(database.Postgres.CreateDatabase))      // 创建真实数据库
		postgres.ANY("/drop_database", context.HandleFunc(database.Postgres.DropDatabase))          // 删除真实数据库
		postgres.ANY("/rename_database", context.HandleFunc(database.Postgres.RenameDatabase))      // 重命名数据库
		postgres.ANY("/version", context.HandleFunc(database.Postgres.Version))                     // 版本及可用功能
		postgres.ANY("/schemas", context.HandleFunc(database.Postgres.Schemas))                     // 架构列表
		postgres.ANY("/create_table", context.HandleFunc(database.Postgres.CreateTable))            // 创建数据表
		postgres.ANY("/drop_table", context.HandleFunc(database.Postgres.DropTable))                // 删除数据表
		postgres.ANY("/rename_table", context.HandleFunc(database.Postgres.RenameTable))            // 重命名数据表
		postgres.ANY("/add_column", context.HandleFunc(database.Postgres.AddColumn))                // 添加字段
		postgres.ANY("/update_column", context.HandleFunc(database.Postgres.UpdateColumn))          // 修改字段
		postgres.ANY("/drop_column", context.HandleFunc(database.Postgres.DropColumn))              // 删除字段
		postgres.ANY("/create_index", context.HandleFunc(database.Postgres.CreateIndex))            // 创建索引
		postgres.ANY("/drop_index", context.HandleFunc(database.Postgres.DropIndex))                // 删除索引
		postgres.ANY("/ddl", context.HandleFunc(database.Postgres.DDL))                             // 查看建表语句
		postgres.ANY("/create_view", context.HandleFunc(database.Postgres.CreateView))              // 创建视图
		postgres.ANY("/import", context.HandleFunc(database.Postgres.Import))                       // 导入数据
		postgres.ANY("/export", context.HandleFunc(database.Postgres.Export))                       // 导出数据到服务器
		postgres.ANY("/objects", context.HandleFunc(database.Postgres.Objects))                     // 数据库对象
		postgres.ANY("/object_definition", context.HandleFunc(database.Postgres.ObjectDefinition))  // 查看对象定义
		postgres.ANY("/save_object", context.HandleFunc(database.Postgres.SaveObject))              // 保存数据库对象
		postgres.ANY("/drop_object", context.HandleFunc(database.Postgres.DropObject))              // 删除数据库对象
		postgres.ANY("/create_foreign_key", context.HandleFunc(database.Postgres.CreateForeignKey)) // 创建外键
		postgres.ANY("/update_foreign_key", context.HandleFunc(database.Postgres.UpdateForeignKey)) // 修改外键
		postgres.ANY("/drop_foreign_key", context.HandleFunc(database.Postgres.DropForeignKey))     // 删除外键
		postgres.ANY("/tables", context.HandleFunc(database.Postgres.Tables))                       // 数据表列表
		postgres.ANY("/columns", context.HandleFunc(database.Postgres.Columns))                     // 字段信息
		postgres.ANY("/indexes", context.HandleFunc(database.Postgres.Indexes))                     // 索引信息
		postgres.ANY("/rows", context.HandleFunc(database.Postgres.Rows))                           // 分页数据
		postgres.ANY("/query", context.HandleFunc(database.Postgres.Query))                         // 执行查询
		postgres.ANY("/execute", context.HandleFunc(database.Postgres.Execute))                     // 执行语句
		postgres.ANY("/insert", context.HandleFunc(database.Postgres.Insert))                       // 添加数据
		postgres.ANY("/update", context.HandleFunc(database.Postgres.Update))                       // 修改数据
		postgres.ANY("/delete", context.HandleFunc(database.Postgres.Delete))                       // 删除数据
		postgres.ANY("/users", context.HandleFunc(database.Postgres.Users))                         // 账户权限
		postgres.ANY("/maintenance", context.HandleFunc(database.Postgres.Maintenance))             // 维护工具
	}
	sqlserver := g.Group("/sqlserver")
	{
		sqlserver.ANY("/databases", context.HandleFunc(database.SqlServer.Databases))                 // 实时数据库列表
		sqlserver.ANY("/database_info", context.HandleFunc(database.SqlServer.DatabaseInfo))          // 真实数据库信息
		sqlserver.ANY("/create_database", context.HandleFunc(database.SqlServer.CreateDatabase))      // 创建真实数据库
		sqlserver.ANY("/drop_database", context.HandleFunc(database.SqlServer.DropDatabase))          // 删除真实数据库
		sqlserver.ANY("/rename_database", context.HandleFunc(database.SqlServer.RenameDatabase))      // 重命名数据库
		sqlserver.ANY("/version", context.HandleFunc(database.SqlServer.Version))                     // 版本及可用功能
		sqlserver.ANY("/schemas", context.HandleFunc(database.SqlServer.Schemas))                     // 架构列表
		sqlserver.ANY("/create_table", context.HandleFunc(database.SqlServer.CreateTable))            // 创建数据表
		sqlserver.ANY("/drop_table", context.HandleFunc(database.SqlServer.DropTable))                // 删除数据表
		sqlserver.ANY("/rename_table", context.HandleFunc(database.SqlServer.RenameTable))            // 重命名数据表
		sqlserver.ANY("/add_column", context.HandleFunc(database.SqlServer.AddColumn))                // 添加字段
		sqlserver.ANY("/update_column", context.HandleFunc(database.SqlServer.UpdateColumn))          // 修改字段
		sqlserver.ANY("/drop_column", context.HandleFunc(database.SqlServer.DropColumn))              // 删除字段
		sqlserver.ANY("/create_index", context.HandleFunc(database.SqlServer.CreateIndex))            // 创建索引
		sqlserver.ANY("/drop_index", context.HandleFunc(database.SqlServer.DropIndex))                // 删除索引
		sqlserver.ANY("/ddl", context.HandleFunc(database.SqlServer.DDL))                             // 查看建表语句
		sqlserver.ANY("/create_view", context.HandleFunc(database.SqlServer.CreateView))              // 创建视图
		sqlserver.ANY("/import", context.HandleFunc(database.SqlServer.Import))                       // 导入数据
		sqlserver.ANY("/export", context.HandleFunc(database.SqlServer.Export))                       // 导出数据到服务器
		sqlserver.ANY("/objects", context.HandleFunc(database.SqlServer.Objects))                     // 数据库对象
		sqlserver.ANY("/object_definition", context.HandleFunc(database.SqlServer.ObjectDefinition))  // 查看对象定义
		sqlserver.ANY("/save_object", context.HandleFunc(database.SqlServer.SaveObject))              // 保存数据库对象
		sqlserver.ANY("/drop_object", context.HandleFunc(database.SqlServer.DropObject))              // 删除数据库对象
		sqlserver.ANY("/create_foreign_key", context.HandleFunc(database.SqlServer.CreateForeignKey)) // 创建外键
		sqlserver.ANY("/update_foreign_key", context.HandleFunc(database.SqlServer.UpdateForeignKey)) // 修改外键
		sqlserver.ANY("/drop_foreign_key", context.HandleFunc(database.SqlServer.DropForeignKey))     // 删除外键
		sqlserver.ANY("/tables", context.HandleFunc(database.SqlServer.Tables))                       // 数据表列表
		sqlserver.ANY("/columns", context.HandleFunc(database.SqlServer.Columns))                     // 字段信息
		sqlserver.ANY("/indexes", context.HandleFunc(database.SqlServer.Indexes))                     // 索引信息
		sqlserver.ANY("/rows", context.HandleFunc(database.SqlServer.Rows))                           // 分页数据
		sqlserver.ANY("/query", context.HandleFunc(database.SqlServer.Query))                         // 执行查询
		sqlserver.ANY("/execute", context.HandleFunc(database.SqlServer.Execute))                     // 执行语句
		sqlserver.ANY("/insert", context.HandleFunc(database.SqlServer.Insert))                       // 添加数据
		sqlserver.ANY("/update", context.HandleFunc(database.SqlServer.Update))                       // 修改数据
		sqlserver.ANY("/delete", context.HandleFunc(database.SqlServer.Delete))                       // 删除数据
		sqlserver.ANY("/users", context.HandleFunc(database.SqlServer.Users))                         // 账户权限
		sqlserver.ANY("/maintenance", context.HandleFunc(database.SqlServer.Maintenance))             // 维护工具
	}
	sqlite := g.Group("/sqlite")
	{
		sqlite.ANY("/databases", context.HandleFunc(database.Sqlite.Databases))                 // 实时数据库列表
		sqlite.ANY("/database_info", context.HandleFunc(database.Sqlite.DatabaseInfo))          // 真实数据库信息
		sqlite.ANY("/version", context.HandleFunc(database.Sqlite.Version))                     // 版本及可用功能
		sqlite.ANY("/schemas", context.HandleFunc(database.Sqlite.Schemas))                     // 架构列表
		sqlite.ANY("/create_table", context.HandleFunc(database.Sqlite.CreateTable))            // 创建数据表
		sqlite.ANY("/drop_table", context.HandleFunc(database.Sqlite.DropTable))                // 删除数据表
		sqlite.ANY("/rename_table", context.HandleFunc(database.Sqlite.RenameTable))            // 重命名数据表
		sqlite.ANY("/add_column", context.HandleFunc(database.Sqlite.AddColumn))                // 添加字段
		sqlite.ANY("/update_column", context.HandleFunc(database.Sqlite.UpdateColumn))          // 修改字段
		sqlite.ANY("/drop_column", context.HandleFunc(database.Sqlite.DropColumn))              // 删除字段
		sqlite.ANY("/create_index", context.HandleFunc(database.Sqlite.CreateIndex))            // 创建索引
		sqlite.ANY("/drop_index", context.HandleFunc(database.Sqlite.DropIndex))                // 删除索引
		sqlite.ANY("/ddl", context.HandleFunc(database.Sqlite.DDL))                             // 查看建表语句
		sqlite.ANY("/create_view", context.HandleFunc(database.Sqlite.CreateView))              // 创建视图
		sqlite.ANY("/import", context.HandleFunc(database.Sqlite.Import))                       // 导入数据
		sqlite.ANY("/export", context.HandleFunc(database.Sqlite.Export))                       // 导出数据到服务器
		sqlite.ANY("/objects", context.HandleFunc(database.Sqlite.Objects))                     // 数据库对象
		sqlite.ANY("/object_definition", context.HandleFunc(database.Sqlite.ObjectDefinition))  // 查看对象定义
		sqlite.ANY("/save_object", context.HandleFunc(database.Sqlite.SaveObject))              // 保存数据库对象
		sqlite.ANY("/drop_object", context.HandleFunc(database.Sqlite.DropObject))              // 删除数据库对象
		sqlite.ANY("/create_foreign_key", context.HandleFunc(database.Sqlite.CreateForeignKey)) // 创建外键
		sqlite.ANY("/update_foreign_key", context.HandleFunc(database.Sqlite.UpdateForeignKey)) // 修改外键
		sqlite.ANY("/drop_foreign_key", context.HandleFunc(database.Sqlite.DropForeignKey))     // 删除外键
		sqlite.ANY("/tables", context.HandleFunc(database.Sqlite.Tables))                       // 数据表列表
		sqlite.ANY("/columns", context.HandleFunc(database.Sqlite.Columns))                     // 字段信息
		sqlite.ANY("/indexes", context.HandleFunc(database.Sqlite.Indexes))                     // 索引信息
		sqlite.ANY("/rows", context.HandleFunc(database.Sqlite.Rows))                           // 分页数据
		sqlite.ANY("/query", context.HandleFunc(database.Sqlite.Query))                         // 执行查询
		sqlite.ANY("/execute", context.HandleFunc(database.Sqlite.Execute))                     // 执行语句
		sqlite.ANY("/insert", context.HandleFunc(database.Sqlite.Insert))                       // 添加数据
		sqlite.ANY("/update", context.HandleFunc(database.Sqlite.Update))                       // 修改数据
		sqlite.ANY("/delete", context.HandleFunc(database.Sqlite.Delete))                       // 删除数据
		sqlite.ANY("/maintenance", context.HandleFunc(database.Sqlite.Maintenance))             // 维护工具
	}
	redis := g.Group("/redis")
	{
		redis.ANY("/databases", context.HandleFunc(database.Redis.Databases))        // 实时数据库列表
		redis.ANY("/database_info", context.HandleFunc(database.Redis.DatabaseInfo)) // 真实数据库信息
		redis.ANY("/version", context.HandleFunc(database.Redis.Version))            // 版本及可用功能
		redis.ANY("/keys", context.HandleFunc(database.Redis.Keys))                  // 键列表
		redis.ANY("/value", context.HandleFunc(database.Redis.Value))                // 键值详情
		redis.ANY("/save", context.HandleFunc(database.Redis.Save))                  // 保存键值
		redis.ANY("/delete", context.HandleFunc(database.Redis.Delete))              // 删除键值
		redis.ANY("/rename", context.HandleFunc(database.Redis.Rename))              // 重命名键
		redis.ANY("/expire", context.HandleFunc(database.Redis.Expire))              // 设置有效期
		redis.ANY("/command", context.HandleFunc(database.Redis.Command))            // 执行命令
		redis.ANY("/info", context.HandleFunc(database.Redis.Info))                  // 运行信息
		redis.ANY("/users", context.HandleFunc(database.Redis.Users))                // 账户权限
		redis.ANY("/maintenance", context.HandleFunc(database.Redis.Maintenance))    // 维护工具
		redis.ANY("/import", context.HandleFunc(database.Redis.Import))              // 导入数据
		redis.ANY("/export", context.HandleFunc(database.Redis.Export))              // 导出数据
	}
	mongo := g.Group("/mongo")
	{
		mongo.ANY("/databases", context.HandleFunc(database.Mongo.Databases))                // 实时数据库列表
		mongo.ANY("/database_info", context.HandleFunc(database.Mongo.DatabaseInfo))         // 真实数据库信息
		mongo.ANY("/create_database", context.HandleFunc(database.Mongo.CreateDatabase))     // 创建真实数据库
		mongo.ANY("/drop_database", context.HandleFunc(database.Mongo.DropDatabase))         // 删除真实数据库
		mongo.ANY("/version", context.HandleFunc(database.Mongo.Version))                    // 版本及可用功能
		mongo.ANY("/collections", context.HandleFunc(database.Mongo.Collections))            // 集合列表
		mongo.ANY("/indexes", context.HandleFunc(database.Mongo.Indexes))                    // 索引列表
		mongo.ANY("/documents", context.HandleFunc(database.Mongo.Documents))                // 文档列表
		mongo.ANY("/insert", context.HandleFunc(database.Mongo.Insert))                      // 添加文档
		mongo.ANY("/update", context.HandleFunc(database.Mongo.Update))                      // 修改文档
		mongo.ANY("/delete", context.HandleFunc(database.Mongo.Delete))                      // 删除文档
		mongo.ANY("/aggregate", context.HandleFunc(database.Mongo.Aggregate))                // 聚合查询
		mongo.ANY("/command", context.HandleFunc(database.Mongo.Command))                    // 执行命令
		mongo.ANY("/create_collection", context.HandleFunc(database.Mongo.CreateCollection)) // 创建集合
		mongo.ANY("/drop_collection", context.HandleFunc(database.Mongo.DropCollection))     // 删除集合
		mongo.ANY("/create_index", context.HandleFunc(database.Mongo.CreateIndex))           // 创建索引
		mongo.ANY("/drop_index", context.HandleFunc(database.Mongo.DropIndex))               // 删除索引
		mongo.ANY("/users", context.HandleFunc(database.Mongo.Users))                        // 账户权限
		mongo.ANY("/maintenance", context.HandleFunc(database.Mongo.Maintenance))            // 维护工具
		mongo.ANY("/import", context.HandleFunc(database.Mongo.Import))                      // 导入数据
		mongo.ANY("/export", context.HandleFunc(database.Mongo.Export))                      // 导出数据
		mongo.ANY("/create_view", context.HandleFunc(database.Mongo.CreateView))             // 创建视图
		mongo.ANY("/update_view", context.HandleFunc(database.Mongo.UpdateView))             // 修改视图
		mongo.ANY("/view_definition", context.HandleFunc(database.Mongo.ViewDefinition))     // 视图定义
	}
	elasticsearch := g.Group("/elasticsearch")
	{
		elasticsearch.ANY("/version", context.HandleFunc(database.Elasticsearch.Version))                // 版本及可用功能
		elasticsearch.ANY("/indices", context.HandleFunc(database.Elasticsearch.Indices))                // 索引列表
		elasticsearch.ANY("/info", context.HandleFunc(database.Elasticsearch.Info))                      // 索引详情
		elasticsearch.ANY("/create_index", context.HandleFunc(database.Elasticsearch.CreateIndex))       // 创建索引
		elasticsearch.ANY("/drop_index", context.HandleFunc(database.Elasticsearch.DropIndex))           // 删除索引
		elasticsearch.ANY("/documents", context.HandleFunc(database.Elasticsearch.Documents))            // 分页文档
		elasticsearch.ANY("/close_search", context.HandleFunc(database.Elasticsearch.CloseSearch))       // 释放分页会话
		elasticsearch.ANY("/insert", context.HandleFunc(database.Elasticsearch.Insert))                  // 添加文档
		elasticsearch.ANY("/update", context.HandleFunc(database.Elasticsearch.Update))                  // 修改文档
		elasticsearch.ANY("/delete", context.HandleFunc(database.Elasticsearch.Delete))                  // 删除文档
		elasticsearch.ANY("/mapping", context.HandleFunc(database.Elasticsearch.Mapping))                // 字段映射
		elasticsearch.ANY("/analyzers", context.HandleFunc(database.Elasticsearch.Analyzers))            // 分析器选项
		elasticsearch.ANY("/update_mapping", context.HandleFunc(database.Elasticsearch.UpdateMapping))   // 修改字段映射
		elasticsearch.ANY("/refresh_fields", context.HandleFunc(database.Elasticsearch.RefreshFields))   // 回填历史文档
		elasticsearch.ANY("/settings", context.HandleFunc(database.Elasticsearch.Settings))              // 索引设置
		elasticsearch.ANY("/update_settings", context.HandleFunc(database.Elasticsearch.UpdateSettings)) // 修改索引设置
		elasticsearch.ANY("/status", context.HandleFunc(database.Elasticsearch.Status))                  // 集群状态
		elasticsearch.ANY("/request", context.HandleFunc(database.Elasticsearch.Request))                // 执行 REST 请求
	}
}

// loadSoftwareRoute 软件包统一管理，运行能力按软件声明检查。
func loadSoftwareRoute(s *sinking_web.Engine) {
	g := s.Group("/software")
	g.Use(context.HandleFunc(middleware.CheckLogin))
	g.ANY("/list", context.HandleFunc(software.List))
	g.ANY("/info", context.HandleFunc(software.Info))
	g.ANY("/install", context.HandleFunc(software.Install))
	g.ANY("/update", context.HandleFunc(software.Update))
	g.ANY("/uninstall", context.HandleFunc(software.Uninstall))
	g.ANY("/:code/page", context.HandleFunc(software.Page))
	g.ANY("/:code/page/*path", context.HandleFunc(software.Page))
	g.ANY("/:code/:method", context.HandleFunc(software.Invoke))
}

// loadContainerRoute 容器通用操作。
func loadContainerRoute(s *sinking_web.Engine) {
	if !platform.CanRunContainer() {
		return
	}
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
	g.ANY("/enum", context.HandleFunc(types.Enum))     //类型枚举
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
