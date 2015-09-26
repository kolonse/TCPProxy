// RTCPProxyClient project RTCPProxyClient.go
package main

import (
	"TCPProxy/RTCPProxyServer/TCPServer"
	"TCPProxy/TCPProxyProto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
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

func ServerStart(serverInfo *TCPServer.ServerInfo) {
	// 连接服务端
	serverAddr := fmt.Sprintf("%v:%v", serverInfo.Ip, serverInfo.Port)
	fmt.Println("CONN ", serverAddr)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil { // 连接不上服务就直接退出
		fmt.Println(err.Error())
		return
	}
	// 发送连接数据包
	pp := TCPProxyProto.NewProxyProto()
	pp.StringifyConn().
		StringifyEnd()
	_, err = conn.Write(pp.GetBuff())
	if err != nil { // 如果出现连接失败 那么直接退出
		conn.Close()
		fmt.Println(err.Error())
		return
	}
	buff := make([]byte, 10000)
	bExit := make(chan bool, 1)
	for {
		n, err := conn.Read(buff)
		if err != nil { //连接出现错误 关闭连接退出
			fmt.Println(err.Error())
			conn.Close()
			bExit <- false
			break
		}
		fmt.Println(n)
	}
	<-bExit
}

func main() {
	for {
		// 获取服务信息
		respInfo := RegisterProxy()
		if !CheckServerStatus(respInfo) {
			fmt.Println("code:", respInfo.Code, " message:", respInfo.Message)
			return
		}

		serverInfo := respInfo.Extern.(*TCPServer.ServerInfo)
		fmt.Println(serverInfo.Dump())

		ServerStart(serverInfo)
		time.Sleep(500 * time.Millisecond)
	}
}
