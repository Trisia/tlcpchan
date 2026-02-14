package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Trisia/tlcpchan-ui/server"
)

var (
	listen    = flag.String("listen", ":30000", "监听地址")
	apiAddr   = flag.String("api", "http://localhost:30080", "后端API地址")
	staticDir = flag.String("static", "./dist", "静态文件目录")
	showVer   = flag.Bool("version", false, "显示版本信息")
	version   = "1.0.0"
)

func main() {
	flag.Parse()

	if *showVer {
		log.Printf("tlcpchan-ui version %s\n", version)
		return
	}

	srv := server.New(*staticDir, *apiAddr, version)
	log.Printf("UI服务启动，监听地址: %s", *listen)
	if err := http.ListenAndServe(*listen, srv); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
