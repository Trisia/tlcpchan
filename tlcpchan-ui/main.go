package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Trisia/tlcpchan-ui/server"
)

var (
	listen    = flag.String("listen", ":3000", "监听地址")
	apiAddr   = flag.String("api", "http://localhost:8080", "后端API地址")
	staticDir = flag.String("static", "./dist", "静态文件目录")
)

func main() {
	flag.Parse()

	srv := server.New(*staticDir, *apiAddr)
	log.Printf("UI服务启动，监听地址: %s", *listen)
	if err := http.ListenAndServe(*listen, srv); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
