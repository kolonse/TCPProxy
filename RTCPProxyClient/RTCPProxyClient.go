// RTCPProxyClient project RTCPProxyClient.go
package main

import (
	"RTCPProxyServer/TCPServer"
	"TCPProxyProto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func RegisterProxy() *TCPProxyProto.RespProto {
	var httpHost = fmt.Sprintf("http://%v/RegisterProxy?domain=%v&name=%v", *Server, *ProxyDomain, *ProxyName)
	fmt.Println("Request " + httpHost)
	resp, err := http.Get(httpHost)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buff))
	serverInfo := TCPServer.NewServerInfo()
	respInfo := TCPProxyProto.NewRespProto(0, "", serverInfo)
	err = json.Unmarshal(buff, &respInfo)
	if err != nil {
		panic(err)
	}
	return respInfo
}

func CheckServerStatus(rp *TCPProxyProto.RespProto) bool {
	if rp.Code != TCPProxyProto.RPOXY_PROTO_SUCCESS {
		return false
	}
	return true
}

func main() {
	// 获取服务信息
	respInfo := RegisterProxy()
	if !CheckServerStatus(respInfo) {
		fmt.Println("code:", respInfo.Code, " message:", respInfo.Message)
		return
	}

	serverInfo := respInfo.Extern.(*TCPServer.ServerInfo)
	fmt.Println(serverInfo.Dump())
}
