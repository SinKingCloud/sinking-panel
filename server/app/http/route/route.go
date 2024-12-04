package route

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"net/http"
	"server/app/constant"
	"server/app/util"
	"time"
)

// loadErrorHandle 设置错误回调
func loadErrorHandle(s *sinking_web.Engine) {
	//设置错误回调
	s.SetErrorHandle(&sinking_web.ErrorHandel{
		//资源不存在错误
		NotFound: func(c *sinking_web.Context) {
			c.JSON(404, sinking_web.H{"code": 404, "message": "请求资源不存在"})
		},
		//系统错误
		Fail: func(c *sinking_web.Context, code int, message string) {
			c.JSON(500, sinking_web.H{"code": code, "message": message})
		},
	})
}

// Init 初始化server
func Init() {
	//设置是否以debug模式运行
	sinking_web.SetDebugMode(util.Conf.GetString(constant.ServerMode) == "dev")
	//实例化一个http server
	r := sinking_web.New()
	r.Use(sinking_web.Recovery())
	//加载错误handle
	loadErrorHandle(r)
	//加载app
	loadApp(r)
	//启动http server
	host := util.Conf.GetString(constant.ServerHost)
	port := util.Conf.GetString(constant.ServerPort)
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" {
		port = "5678"
	}
	server := &http.Server{
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
		Addr:         host + ":" + port,
		Handler:      r,
	}
	util.Log.Println("服务端启动成功", host+":"+port)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
		return
	}
}
