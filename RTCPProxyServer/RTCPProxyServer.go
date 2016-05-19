// RTCPProxyServer project RTCPProxyServer.go
package main

import (
	"github.com/kardianos/service"
	"github.com/kolonse/KolonseWeb"
	"github.com/kolonse/KolonseWeb/HttpLib"
	"github.com/kolonse/KolonseWeb/Type"
	"github.com/kolonse/TCPProxy/RTCPProxyServer/TCPServer"
	. "github.com/kolonse/TCPProxy/TCPProxyProto"
	"strconv"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// Do work here
	KolonseWeb.DefaultApp.Get("/RegisterProxy", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		port := req.URL.Query().Get("port") // 启动监听端口
		KolonseWeb.DefaultLogs().Info("处理客户端请求 Req Client Addr:%v Register Port:%v",
			req.RemoteAddr, port)
		// 创建一个TCP服务 然后启动
		portInt, _ := strconv.Atoi(port)
		ts := TCPServer.NewServerInfo()
		err := ts.Start(portInt)
		if err != nil {
			res.Json(NewRespProto(RPOXY_PROTO_ERROR_NET, err.Error(), nil)) // 返回服务状态

		} else {
			res.Json(NewRespProto(RPOXY_PROTO_SUCCESS, "", nil)) // 返回服务状态
		}
	})
	KolonseWeb.DefaultApp.Listen("0.0.0.0", *Port)
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "RTCPProxyServer",
		DisplayName: "逆向TCP代理服务",
		Description: "主要用来处理第三方回调的问题 方便调试",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		KolonseWeb.DefaultLogs().Error(err.Error())
	}
	logger, err := s.Logger(nil)
	if err != nil {
		KolonseWeb.DefaultLogs().Error(err.Error())
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
