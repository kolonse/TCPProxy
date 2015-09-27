// TCPServer
package TCPServer

import (
	"KolonseWeb"
	. "TCPProxy/TCPProxyProto"
	"encoding/json"
	"fmt"
	//	"io"
	"net"
	"strconv"
	"time"
)

const (
	SERVER_STATUS_NOTEXIST       = "not exist"
	SERVER_STATUS_WAIT           = "wait"
	SERVER_STATUS_RUNNING        = "running"
	SERVER_STATUS_TCP_ADDR_ERROR = "tcp addr error"
	SERVER_STATUS_LISTEN_FAIL    = "listen fail"
)

type Who struct {
	Name string       // 服务者名字
	Ip   string       // 客户端 IP
	conn *net.TCPConn // 连接套接字
}

func (w *Who) Close() {
	if w.conn != nil {
		(*w.conn).Close()
	}
}

func (w *Who) Dump() string {
	ret := ""
	ret += fmt.Sprintln("\tName:", w.Name)
	ret += fmt.Sprintln("\tIp:", w.Ip)
	return ret
}

func (w *Who) Set(name string, ip string, conn *net.TCPConn) {
	w.Close()
	w.conn = conn
	w.Name = name
	w.Ip = ip
}

type ServerInfo struct {
	Domain  string
	Port    uint16
	Desc    string
	Status  string // 服务状态
	Ip      string
	ForWho  Who
	Parters map[string]Who
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
	ret += fmt.Sprintln("ForWho:")
	ret += si.ForWho.Dump()
	return ret
}

func (si *ServerInfo) handleConnection(conn *net.TCPConn) {
	KolonseWeb.DefaultLogs().Info("Recv Conn,RemoteAddr:%v %v %v %v",
		conn.RemoteAddr().Network(), conn.RemoteAddr().String(),
		conn.LocalAddr().Network(), conn.LocalAddr().String())
	// 收到一个连接 读取开始 i'm proxy server
	buff := make([]byte, 10240) //  缓存数据
	// 检测是否是 PROXY SERVER
	for {
		n, err := conn.Read(buff)
		if err != nil { // 连接出现错误
			conn.Close()
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
			break
		}
		status := CheckHeadMark(buff[:n])
		if status == RPOXY_PROTO_ERROR_LENGTH { // 长度不足 需要继续接收
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Length Not Enougth", conn.RemoteAddr(), conn.LocalAddr())
			continue
		} else if status == RPOXY_PROTO_SUCCESS { // 说明属于 代理协议
			pp := NewProxyProto()
			Err := pp.Parse(buff[:n]).GetError()
			if Err.GetCode() == RPOXY_PROTO_ERROR_LENGTH {
				continue
			} else if Err.GetCode() != RPOXY_PROTO_SUCCESS {
				KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), Err.Error())
				conn.Close()
				break
			}
			KolonseWeb.DefaultLogs().Info("RECV:\r\n%v", pp.HeaderString())
			KolonseWeb.DefaultLogs().Info("")
			si.ProcessProxyConn(conn, pp)
			break
		} else {
			si.ProcessReqConn(conn, buff[:n])
			break
		}
	}
}

func (si *ServerInfo) ProcessProxyConn(conn *net.TCPConn, pp *ProxyProto) {
	KolonseWeb.DefaultLogs().Info("Addr r_%v|l_%v 处理代理协议", conn.RemoteAddr(), conn.LocalAddr())
	KolonseWeb.DefaultLogs().Info("RECV:%v %v", pp.GetMethod(), pp.GetVersion())
	switch pp.GetMethod() {
	case PROXY_PROTO_METHOD_CONN:
		si.ProcessProtoConn(conn, pp)
	case PROXY_PROTO_METHOD_RES:
		si.ProcessProtoRes(conn, pp)
	}
	buff := make([]byte, 10240)
	for {

		index := 0
		n, err := conn.Read(buff[index:])
		if err != nil { // 连接出现错误
			conn.Close()
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
			break
		}
		pp = NewProxyProto()
		Err := pp.Parse(buff[:n]).GetError()
		if Err.GetCode() == RPOXY_PROTO_ERROR_LENGTH {
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), Err.Error())
			conn.Close()
			break
		} else if Err.GetCode() != RPOXY_PROTO_SUCCESS {
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), Err.Error())
			conn.Close()
			break
		}
		KolonseWeb.DefaultLogs().Info("RECV:%v %v", pp.GetMethod(), pp.GetVersion())
		//KolonseWeb.DefaultLogs().Info("")

		switch pp.GetMethod() {
		case PROXY_PROTO_METHOD_CONN:
			si.ProcessProtoConn(conn, pp)
		case PROXY_PROTO_METHOD_RES:
			si.ProcessProtoRes(conn, pp)
		}
	}
}

func (si *ServerInfo) ProcessProtoConn(conn *net.TCPConn, pp *ProxyProto) {
	// 将连接信息进行赋值
	connInfo := make(map[string]string)
	json.Unmarshal(pp.GetBody(), &connInfo) // 读取出连接时填写的 name 字段
	name := connInfo["name"]
	si.ForWho.Set(name, conn.RemoteAddr().String(), conn)
}

func (si *ServerInfo) ProcessProtoRes(conn *net.TCPConn, pp *ProxyProto) {
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

func (si *ServerInfo) ProcessReqConn(conn *net.TCPConn, buff []byte) {
	KolonseWeb.DefaultLogs().Info("Addr r_%v|l_%v 处理请求", conn.RemoteAddr(), conn.LocalAddr())
	si.Parters[conn.RemoteAddr().String()] = Who{
		conn: conn,
	}
	/** 通知反向代理请求
	*	发送连接请求 并将数据发送过去
	 */
	pp := NewProxyProto()
	pp.StringifyREQ().
		StringifyRemoteAddr(conn.RemoteAddr().String()).
		StringifyBody(buff).
		StringifyEnd()
	_, err := si.ForWho.conn.Write(pp.GetBuff())
	if err != nil {
		KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v",
			conn.RemoteAddr(),
			conn.LocalAddr(),
			err.Error())
		return
	}
	for {
		_, err := conn.Read(buff)
		if err != nil { // 连接出现错误
			conn.Close()
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
			// 通知对端也要关闭连接
			pp = NewProxyProto()
			pp.StringifyClose().
				StringifyRemoteAddr(conn.RemoteAddr().String()).
				StringifyEnd()
			// 将数据发送出去
			_, err := si.ForWho.conn.Write(pp.GetBuff())
			if err != nil {
				KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v",
					conn.RemoteAddr(),
					conn.LocalAddr(),
					err.Error())
			}
			break
		}
		pp = NewProxyProto()
		pp.StringifyREQ().
			StringifyRemoteAddr(conn.RemoteAddr().String()).
			StringifyBody(buff).
			StringifyEnd()
		// 将数据发送出去
		_, err = si.ForWho.conn.Write(pp.GetBuff())
		if err != nil {
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v",
				conn.RemoteAddr(),
				conn.LocalAddr(),
				err.Error())
			return
		}
	}
	return
}

func (si *ServerInfo) Start() error {
	host := ":" + strconv.Itoa(int(si.Port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		si.Status = SERVER_STATUS_TCP_ADDR_ERROR
		return err
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		si.Status = SERVER_STATUS_LISTEN_FAIL
		return err
	}
	si.Status = SERVER_STATUS_RUNNING
	KolonseWeb.DefaultLogs().Info("TCP Server Listen On %v", si.Port)
	go func() {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			go si.handleConnection(conn)
		}
	}()
	return nil
}

func NULLServerInfo() *ServerInfo {
	return &ServerInfo{
		Status: SERVER_STATUS_NOTEXIST,
	}
}

func NewServerInfo() *ServerInfo {
	return &ServerInfo{
		Status:  SERVER_STATUS_WAIT,
		Parters: make(map[string]Who),
	}
}
