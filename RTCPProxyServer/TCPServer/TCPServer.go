// TCPServer
package TCPServer

import (
	. "TCPProxy/TCPProxyProto"
	"encoding/json"
	"fmt"
	"github.com/kolonse/KolonseWeb"
	t "github.com/kolonse/TCPServer"
	"github.com/kolonse/function"
	//	"io"
	"net"
	"strconv"
	//"time"
)

const (
	SERVER_STATUS_NOTEXIST       = "not exist"
	SERVER_STATUS_WAIT           = "wait"
	SERVER_STATUS_RUNNING        = "running"
	SERVER_STATUS_TCP_ADDR_ERROR = "tcp addr error"
	SERVER_STATUS_LISTEN_FAIL    = "listen fail"
)

type Who struct {
	Name   string   // 服务者名字
	Ip     string   // 客户端 IP
	conn   net.Conn // 连接套接字
	Active bool     // 标识该身份是否活跃
}

//func (w *Who) Close() {
//	if w.conn != nil {
//		(*w.conn).Close()
//	}
//}

func (w *Who) IsActive() bool {
	return w.Active
}

//func (w *Who) Dump() string {
//	ret := ""
//	ret += fmt.Sprintln("\tName:", w.Name)
//	ret += fmt.Sprintln("\tIp:", w.Ip)
//	return ret
//}

func (w *Who) SetConn(conn net.Conn) {
	w.conn = conn
	w.Active = true
}

func (w *Who) SetActive(active bool) {
	w.Active = active
}

type ServerInfo struct {
	Domain      string
	Port        uint16
	Desc        string
	Status      string // 服务状态
	Ip          string
	Channel     *t.TCPServer
	ChannelConn Who
	Server      *t.TCPServer
	Parters     map[string]Who
}

func (si *ServerInfo) GetStatus() string {
	return si.Status
}

func (si *ServerInfo) Dump() string {
	ret := ""
	ret += fmt.Sprintln("Domain:", si.Domain)
	ret += fmt.Sprintln("Ip:", si.Ip)
	ret += fmt.Sprintln("Port:", si.Port)
	ret += fmt.Sprintln("Desc:", si.Desc)
	ret += fmt.Sprintln("Status:", si.Status)
	//ret += fmt.Sprintln("ForWho:")
	//ret += si.ForWho.Dump()
	return ret
}

func (si *ServerInfo) ProcessProtoConn(conn net.Conn, pp *ProxyProto) {
	// 将连接信息进行赋值
	connInfo := make(map[string]string)
	json.Unmarshal(pp.GetBody(), &connInfo) // 读取出连接时填写的 name 字段
	KolonseWeb.DefaultLogs().Info("收到连接请求数据 Remote:%v Local:%v info:%v",
		conn.RemoteAddr().String(),
		conn.LocalAddr().String(),
		string(pp.GetBody()))
	si.ChannelConn.SetConn(conn)
}

func (si *ServerInfo) ProcessProtoRes(conn net.Conn, pp *ProxyProto) {
	remoteAddr := pp.GetRemoteAddr()
	pWho, ok := si.Parters[remoteAddr]
	if !ok {
		//  连接已经不存在 通知客户端关闭连接
		KolonseWeb.DefaultLogs().Error("RemoteAddr:%v not found", remoteAddr)
		return //
	}
	// 发送消息
	pWho.conn.Write(pp.GetBody())
}

func (si *ServerInfo) channelConn(conn *t.TCPConn) {
	// 收到隧道的连接
	KolonseWeb.DefaultLogs().Info("收到隧道连接 Remote:%v Local:%v",
		conn.RemoteAddr().String(),
		conn.LocalAddr().String())
}

func (si *ServerInfo) channelRecv(conn *t.TCPConn, buff []byte, err error) {
	if err != nil {
		// 关闭所有的服务连接
		//conn.Close()
		KolonseWeb.DefaultLogs().Error("隧道连接 Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
		return
	}
	// 将接收到的数据进行解析 并发送到对应的连接上
	pp := NewProxyProto()
	Err := pp.Parse(buff).GetError()
	if Err.GetCode() == RPOXY_PROTO_ERROR_LENGTH {
		KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), Err.Error())
		conn.Close()
		return
	} else if Err.GetCode() != RPOXY_PROTO_SUCCESS {
		KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), Err.Error())
		conn.Close()
		return
	}
	KolonseWeb.DefaultLogs().Debug("RECV:%v %v", pp.GetMethod(), pp.GetVersion())
	switch pp.GetMethod() { // 处理协议
	case PROXY_PROTO_METHOD_CONN:
		si.ProcessProtoConn(conn.Conn, pp)
	case PROXY_PROTO_METHOD_RES:
		si.ProcessProtoRes(conn.Conn, pp)
	}
}

func (si *ServerInfo) serverConn(conn *t.TCPConn) {
	// 收到用户数据的连接
	KolonseWeb.DefaultLogs().Info("收到服务连接 Remote:%v Local:%v",
		conn.RemoteAddr().String(),
		conn.LocalAddr().String())
	who := Who{}
	who.SetConn(conn.Conn)
	si.Parters[conn.RemoteAddr().String()] = who
}

func (si *ServerInfo) serverRecv(conn *t.TCPConn, buff []byte, err error) {
	// 收到用户数据
	if err != nil {
		pp := NewProxyProto()
		pp.StringifyClose().
			StringifyRemoteAddr(conn.RemoteAddr().String()).
			StringifyEnd()
		// 将数据发送出去
		_, err := si.ChannelConn.conn.Write(pp.GetBuff())
		if err != nil {
			KolonseWeb.DefaultLogs().Error("服务连接 Addr r_%v|l_%v Error:%v",
				conn.RemoteAddr(),
				conn.LocalAddr(),
				err.Error())
		}
		return
	}
	KolonseWeb.DefaultLogs().Info("服务连接 Addr r_%v|l_%v 收到数据,size:%v",
		conn.RemoteAddr(),
		conn.LocalAddr(),
		len(buff))
	pp := NewProxyProto()
	pp.StringifyREQ().
		StringifyRemoteAddr(conn.RemoteAddr().String()).
		StringifyBody(buff).
		StringifyEnd()
	// 将数据发送出去
	_, err = si.ChannelConn.conn.Write(pp.GetBuff())
	if err != nil {
		KolonseWeb.DefaultLogs().Error("服务连接 Addr r_%v|l_%v Error:%v",
			conn.RemoteAddr(),
			conn.LocalAddr(),
			err.Error())
		return
	}
}

func (si *ServerInfo) Start(port int) error {
	channelHost := ":" + strconv.Itoa(port+1)
	serverHost := ":" + strconv.Itoa(port)
	si.Channel = t.NewTCPServer(channelHost)
	si.Server = t.NewTCPServer(serverHost)

	si.Channel.Register("newConnCB", function.Bind(si.channelConn, function.PH(function.P_1)))
	si.Channel.Register("recvCB", function.Bind(si.channelRecv,
		function.PH(function.P_1),
		function.PH(function.P_2),
		function.PH(function.P_3)))

	si.Server.Register("newConnCB", function.Bind(si.serverConn, function.PH(function.P_1)))
	si.Server.Register("recvCB", function.Bind(si.serverRecv,
		function.PH(function.P_1),
		function.PH(function.P_2),
		function.PH(function.P_3)))
	go si.Channel.Server()
	go si.Server.Server()
	return nil
}

func NULLServerInfo() *ServerInfo {
	return &ServerInfo{
		Status: SERVER_STATUS_NOTEXIST,
	}
}

func NewServerInfo() *ServerInfo {
	return &ServerInfo{
		Parters: make(map[string]Who),
	}
}
