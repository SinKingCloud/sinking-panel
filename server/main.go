package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		os.Args = []string{"", "start"}
	}
	daemonIns, err := NewUnixDaemon("server.pid", "server.log", func() {
		// 启动HTTP服务
		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprintf(writer, "Hello, World! This is a daemon-powered HTTP service.")
		})
		_ = http.ListenAndServe(":8080", nil)
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	switch os.Args[1] {
	case "start":
		if err = daemonIns.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "stop":
		if err = daemonIns.Stop(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "reload":
		if err = daemonIns.Reload(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println("无效的命令，支持的命令start|stop|reload")
		os.Exit(1)
	}
}
