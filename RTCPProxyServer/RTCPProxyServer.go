// RTCPProxyServer project RTCPProxyServer.go
package main

import (
	"KolonseWeb"
	"KolonseWeb/HttpLib"
	"KolonseWeb/Type"
	"TCPProxy/RTCPProxyServer/TCPServer"
	. "TCPProxy/TCPProxyProto"
)

func main() {
	LoadCfg()
	//TCPServerManager.TCPServerManagerStart()
	KolonseWeb.DefaultLogs().Info("加载代理服务配置:\n%v", TCPServerManager.Dump())
	KolonseWeb.DefaultApp.Get("/RegisterProxy", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		domain := req.URL.Query().Get("domain")
		name := req.URL.Query().Get("name")
		KolonseWeb.DefaultLogs().Info("处理客户端请求 Req Domain:%v ProxyName:%v,Client Addr:%v", domain, name, req.RemoteAddr)
		serverInfo := TCPServerManager.GetServerInfo(domain)
		// 检查一下域名是否是 wait 状态
		if serverInfo.GetStatus() != TCPServer.SERVER_STATUS_WAIT {
			KolonseWeb.DefaultLogs().Error("服务已经被代理 信息如下:\n", serverInfo.Dump())
			res.Json(NewRespProto(RPOXY_PROTO_ERROR_LOGIC, "server is "+serverInfo.GetStatus(), serverInfo)) // 返回服务状态
			return
		}

		// 启动服务
		err := serverInfo.Start()
		if err != nil {
			KolonseWeb.DefaultLogs().Error("服务启动失败 Req Domain:%v ProxyName:%v,Client Addr:%v err:%v",
				domain, name, req.RemoteAddr, err.Error())
			res.Json(NewRespProto(RPOXY_PROTO_ERROR_NET, err.Error(), serverInfo)) // 返回服务状态
			return
		}
		res.Json(NewRespProto(RPOXY_PROTO_SUCCESS, "", serverInfo)) // 返回服务状态
	})
	KolonseWeb.DefaultApp.Listen("0.0.0.0", *Port)
}
