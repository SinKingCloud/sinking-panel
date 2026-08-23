package route

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/service"
	"server/global"
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
func Init(stop <-chan struct{}) {
	service.Init()
	shutdown := service.Boot()
	defer shutdown()

	//实例化一个http server
	r := sinking_web.Default()
	//设置是否以debug模式运行
	r.SetDebugMode(global.App.IsDebug())
	//加载错误handle
	loadErrorHandle(r)
	//加载app
	loadApp(r)
	//启动http server
	host, port := global.App.ServerAddr()
	address := net.JoinHostPort(host, port)
	sinking_web.Author(r, address)
	httpServer := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Minute,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
	}
	done := make(chan error, 1)
	go func() {
		err := httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			panic(err)
		}
		return
	case <-stop:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := httpServer.Shutdown(ctx)
	cancel()
	if err != nil {
		log.Printf("停止面板 HTTP 服务超时: %v", err)
		if closeErr := httpServer.Close(); closeErr != nil {
			log.Printf("强制停止面板 HTTP 服务失败: %v", closeErr)
		}
	}
	if err = <-done; err != nil {
		panic(err)
	}
}
