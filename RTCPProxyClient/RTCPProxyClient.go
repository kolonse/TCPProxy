// RTCPProxyClient project RTCPProxyClient.go
package main

import (
	. "TCPProxy/TCPProxyProto"
	"encoding/json"
	"fmt"
	"github.com/kolonse/kdp"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func RegisterProxy() *RespProto {
	var httpHost = fmt.Sprintf("http://%v/RegisterProxy?domain=%v&name=%v&port=%v",
		*Server, *ProxyDomain, *ProxyName,
		*UsePort)
	fmt.Println("Request " + httpHost)
	resp, err := http.Get(httpHost)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	fmt.Println(string(buff))
	respInfo := NewRespProto(0, "", nil)
	err = json.Unmarshal(buff, &respInfo)
	if err != nil {
		return nil
	}
	return respInfo
}

func CheckServerStatus(rp *RespProto) bool {
	if rp.Code != RPOXY_PROTO_SUCCESS {
		return false
	}
	return true
}

func ServerStart() {
	// 连接服务端
	ip := *Server
	index := strings.Index(ip, ":")
	ip = ip[:index]
	serverAddr := fmt.Sprintf("%v:%v", ip, *UsePort+1)
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
	pp := kdp.NewKDP()
	pp.Add("Method", "CONN").
		StringifyBody(body).
		Stringify()
	fmt.Println(string(pp.GetBuff()))
	_, err = conn.Write(pp.GetBuff())
	if err != nil { // 如果出现连接失败 那么直接退出
		conn.Close()
		fmt.Println(err.Error())
		return
	}

	//bExit := make(chan bool, 1)
	buff := make([]byte, 10000)
	var cache []byte
	for {
		n, err := conn.Read(buff)
		if err != nil { //连接出现错误 关闭连接退出
			fmt.Println(err.Error())
			conn.Close()
			//bExit <- false
			break
		}
		cache = append(cache, buff[:n]...)
		pp = kdp.NewKDP()
		Err := pp.Parse(cache).GetError() //.GetCode()
		if Err.GetCode() != 0 {
			fmt.Println("Parse Error:", Err.Error())
			conn.Close()
			//bExit <- false
			break
		}
		cache = cache[pp.GetProtoLen():]
		fmt.Println("收到请求:\n" + pp.HeaderString())
		switch method, _ := pp.Get("Method"); method {
		case "REQ":
			Req(conn, pp)
		case "Close":
			Close(conn, pp)
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
		pp := kdp.NewKDP()
		pp.Add("Method", "RES").
			Add("RemoteAddr", cp.remoteAddr).
			StringifyBody(buff[:n]).
			Stringify()
		fmt.Println("向代理服务发送:\n" + pp.HeaderString())
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

func Close(conn net.Conn, pp *kdp.KDP) {
	remoteAddr, ok := pp.Get("RemoteAddr")
	if !ok {
		panic("协议中没有找到 RemoteAddr")
	}
	cp, ok := ConnMap[remoteAddr]
	if !ok {
		return
	}
	cp.dstConn.Close()
}

func Req(conn net.Conn, pp *kdp.KDP) {
	remoteAddr, ok := pp.Get("RemoteAddr")
	if !ok {
		panic("协议中没有找到 RemoteAddr")
	}
	cp, ok := ConnMap[remoteAddr]
	if !ok {
		// 创建一个新的连接
		cpNew, err := Create(conn, remoteAddr)
		if err != nil {
			// 发送关闭连接请求
			return
		}
		ConnMap[remoteAddr] = cpNew
		cp = cpNew
	}
	// 将数据发送到连接上
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

		ServerStart()
		time.Sleep(500 * time.Millisecond)
	}
}
