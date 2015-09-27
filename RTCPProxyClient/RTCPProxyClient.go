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
	connInfo := make(map[string]string)
	connInfo["name"] = *ProxyName
	body, _ := json.Marshal(connInfo)
	pp := TCPProxyProto.NewProxyProto()
	pp.StringifyConn().
		StringifyBody(body).
		StringifyEnd()
	_, err = conn.Write(pp.GetBuff())
	if err != nil { // 如果出现连接失败 那么直接退出
		conn.Close()
		fmt.Println(err.Error())
		return
	}

	//bExit := make(chan bool, 1)
	buff := make([]byte, 10000)
	for {

		n, err := conn.Read(buff)
		if err != nil { //连接出现错误 关闭连接退出
			fmt.Println(err.Error())
			conn.Close()
			//bExit <- false
			break
		}
		pp = TCPProxyProto.NewProxyProto()
		Err := pp.Parse(buff[:n]).GetError() //.GetCode()
		if Err.GetCode() != 0 {
			fmt.Println("Parse Error:", Err.Error())
			conn.Close()
			//bExit <- false
			break
		}
		switch pp.GetMethod() {
		case TCPProxyProto.PROXY_PROTO_METHOD_REQ:
			Req(conn, pp)
		}
	}
	//<-bExit
}

type ConnPair struct {
	proxyConn  net.Conn
	dstConn    net.Conn
	remoteAddr string
}

func NewConnPair(pc net.Conn, dc net.Conn, ra string) *ConnPair {
	return &ConnPair{
		proxyConn:  pc,
		dstConn:    dc,
		remoteAddr: ra,
	}
}

var ConnMap map[string]*ConnPair

func ConnRun(cp *ConnPair) {
	buff := make([]byte, 10000)
	for {
		n, err := cp.dstConn.Read(buff)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		pp := TCPProxyProto.NewProxyProto()
		pp.StringifyRES().
			StringifyRemoteAddr(cp.remoteAddr).
			StringifyRemoteAddr(cp.dstConn.LocalAddr().String()).
			StringifyBody(buff[:n]).
			StringifyEnd()
		cp.proxyConn.Write(pp.GetBuff())
	}
}

func Create(pc net.Conn, ra string) (*ConnPair, error) {
	c, err := net.Dial("tcp", *ProxyServer)
	if err != nil {
		return nil, err
	}
	cp := NewConnPair(pc, c, ra)
	go ConnRun(cp)
	return cp, nil
}

func Req(conn net.Conn, pp *TCPProxyProto.ProxyProto) {
	cp, ok := ConnMap[pp.GetRemoteAddr()]
	if !ok {
		// 创建一个新的连接
		cpNew, err := Create(conn, pp.GetRemoteAddr())
		if err != nil {
			// 发送关闭连接请求
			return
		}
		ConnMap[pp.GetRemoteAddr()] = cpNew
		cp = cpNew
	}
	// 将数据发送到连接上
	fmt.Println(cp)
	cp.dstConn.Write(pp.GetBody())
}

func main() {
	ConnMap = make(map[string]*ConnPair)
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
