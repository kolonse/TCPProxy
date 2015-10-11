// RTCPProxyServer project RTCPProxyServer.go
package main

import (
	"TCPProxy/RTCPProxyServer/TCPServer"
	. "TCPProxy/TCPProxyProto"
	"github.com/kolonse/KolonseWeb"
	"github.com/kolonse/KolonseWeb/HttpLib"
	"github.com/kolonse/KolonseWeb/Type"
	"strconv"
)

func main() {
	KolonseWeb.DefaultApp.Get("/RegisterProxy", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		port := req.URL.Query().Get("port") // 启动监听端口
		KolonseWeb.DefaultLogs().Info("处理客户端请求 Req Client Addr:%v Register Port:%v",
			req.RemoteAddr, port)
		// 创建一个TCP服务 然后启动
		portInt, _ := strconv.Atoi(port)
		ts := TCPServer.NewServerInfo()
		ts.Start(portInt)
		res.Json(NewRespProto(RPOXY_PROTO_SUCCESS, "", nil)) // 返回服务状态
	})
	KolonseWeb.DefaultApp.Listen("0.0.0.0", *Port)
}
