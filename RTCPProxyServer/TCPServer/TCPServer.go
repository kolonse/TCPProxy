// TCPServer
package TCPServer

import (
	"KolonseWeb"
	. "TCPProxy/TCPProxyProto"
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
	Domain string
	Port   uint16
	Desc   string
	Status string // 服务状态
	ForWho Who
	Parter Who
}

func (si *ServerInfo) GetStatus() string {
	return si.Status
}

func (si *ServerInfo) Dump() string {
	ret := ""
	ret += fmt.Sprintln("Domain:", si.Domain)
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
	buff := make([]byte, 512) //  缓存长度
	// 检测是否是 PROXY SERVER
	for {
		n, err := conn.Read(buff)
		if err != nil { // 连接出现错误
			conn.Close()
			KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
			return
		}
		status := CheckHeadMark(buff[:n])
		if status == RPOXY_PROTO_ERROR_LENGTH { // 长度不足 需要继续接收
			continue
		} else if status == RPOXY_PROTO_SUCCESS { // 说明属于 代理协议
			si.ProcessProxyConn(conn, buff[:n])
		} else {
			si.ProcessReqConn(conn, buff[:n])
		}
	}
}

func (si *ServerInfo) ProcessProxyConn(conn *net.TCPConn, buff []byte) {
	KolonseWeb.DefaultLogs().Info("Addr r_%v|l_%v 处理代理协议", conn.RemoteAddr(), conn.LocalAddr())
}

func (si *ServerInfo) ProcessReqConn(conn *net.TCPConn, buff []byte) {
	KolonseWeb.DefaultLogs().Info("Addr r_%v|l_%v 处理请求", conn.RemoteAddr(), conn.LocalAddr())
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
		Status: SERVER_STATUS_WAIT,
	}
}
